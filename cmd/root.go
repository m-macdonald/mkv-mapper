/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"m-macdonald/mkv-mapper/cmd/disc"
	"m-macdonald/mkv-mapper/internal/config"

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
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, cfg, "", "Path to the config file")
	rootCmd.PersistentFlags().String(logLevel, "info", "The level at which we should log any messages")
}

func initConfig(cmd *cobra.Command) error {
	viper.BindPFlag(config.LogLevel, cmd.Flags().Lookup(logLevel))
	viper.SetEnvPrefix(config.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

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
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}

		return fmt.Errorf("failed to read config: %w", err)
	}

	return nil
}
