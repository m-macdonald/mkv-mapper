package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"m-macdonald/mkv-mapper/cmd/disc"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	logLevel = "log-level"
	cfg      = "config"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mkv-mapper",
	Short: "Rip optical discs with MakeMKV and name the output using TheDiscDB",
	Long: `mkv-mapper scans the inserted disc with makemkvcon, identifies its titles
using TheDiscDB (https://thediscdb.com), then rips the selected titles and
names the resulting .mkv files from configurable templates.

Settings are read from flags, MKVMAP_* environment variables, and an optional
config.json in the current directory or the user config directory.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cfgFile string

func init() {
	cobra.EnableTraverseRunHooks = true
	rootCmd.AddCommand(disc.Cmd)
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return util.WithUsageHint(cmd, err)
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, cfg, "", "Path to the config file")
	rootCmd.PersistentFlags().String(logLevel, "info", "The level at which we should log any messages")
}

func initConfig(cmd *cobra.Command) error {
	viper.BindPFlag(config.LogLevel, cmd.Flags().Lookup(logLevel))
	if err := util.BindConfigFlags(cmd); err != nil {
		return err
	}
	viper.SetEnvPrefix(config.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	for _, key := range config.Keys {
		if err := viper.BindEnv(key); err != nil {
			return err
		}
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(configDir, config.ProgramDirname))
		viper.SetConfigName(config.ConfigFilename)
		viper.SetConfigType("json")
	}

	if err := viper.ReadInConfig(); err != nil {
		// A config file is optional when searching the default locations, but one named via --config must exist.
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil
		}

		return fmt.Errorf("failed to read config: %w", err)
	}

	return nil
}
