/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"m-macdonald/mkv-mapper/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	PersistentPreRunE: initContext,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var defaultCfgFile = "$HOME/.config/mkv-mapper/config.json"

var cfgFile string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfgFile, fmt.Sprintf("Path to the config config file (default is %s)", defaultCfgFile))
	rootCmd.PersistentFlags().String("output-dir", "", "Output directory for ripped files")
	rootCmd.PersistentFlags().String("disc-root", "", "Disc mount root directory")
	rootCmd.PersistentFlags().String("log-level", "info", "The level at which we should log any messages. Info is the default and probably does not ned to be changed")
	rootCmd.PersistentFlags().String("makemkv-path", "makemkvcon", "The location of the makemkvcon binary. Defaults to assuming the binary is already available on the path")

	// viper.BindPFlags(rootCmd.PersistentFlags())

	viper.BindPFlag(config.OutputDir, rootCmd.PersistentFlags().Lookup("output-dir"))
	viper.BindPFlag(config.DiscRoot, rootCmd.PersistentFlags().Lookup("disc-root"))
	viper.BindPFlag(config.LogLevel, rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag(config.MakeMkvPath, rootCmd.PersistentFlags().Lookup("makemkv-path"))
}

type contextKey struct {}
var appContextKey = contextKey {}

type AppContext struct {
	Config *config.Config
	Logger *zap.SugaredLogger
}

func initContext(cmd *cobra.Command, args []string) error {
	config, err := initConfig()
	if err != nil {
		return err
	}
	println(fmt.Sprintf("%v", config))
	logger, err := initLogger()
	if err != nil {
		return err
	}
	defer logger.Sync()

	appContext := AppContext{
		Config: config,
		Logger: logger,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, appContextKey, appContext)

	cmd.SetContext(ctx)

	return nil
}

func initConfig() (*config.Config, error) {
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("MKVMAP")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s %w", cfgFile, err)
	}

	var cfg config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config %w", err)
	}

	return &cfg, nil
}

func initLogger() (*zap.SugaredLogger, error) {
	logLevelStr := viper.GetString(config.LogLevel)
	logLevel, err := zapcore.ParseLevel(logLevelStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level from given value: %s", logLevelStr)
	}
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(logLevel)

	// Sugaring the logger by default as this is code is not performance critical
	return zap.Must(loggerConfig.Build()).Sugar(), nil
}
