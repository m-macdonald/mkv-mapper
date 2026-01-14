/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: initConfig,
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

var debug bool
var cfgFile string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfgFile, fmt.Sprintf("config file (default is %s)", defaultCfgFile))
    rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Pass this flag to output more logging as mkv-mapper works")
        
    viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}

func initConfig(cmd *cobra.Command, args []string) {
    // var logLevel zapcore.Level;
    // if debug {
    //     logLevel = zap.DebugLevel
    // } else {
    //     logLevel = zap.InfoLevel
    // }

    // loggerConfig := zap.Config {
    //     Level: zap.NewAtomicLevelAt(logLevel),
    //     Development: false,
    //     DisableCaller: false,
    //     Encoding: "console",
    //     OutputPaths: []string { "stdout" },
    //     ErrorOutputPaths: []string { "stderr" },
    // }

    // Sugaring the logger by default as this is code is not performance critical
    logger := zap.Must(zap.NewDevelopment()).Sugar()

    logger.Infoln("Initializing global config")

    viper.SetConfigFile(cfgFile)
    viper.ReadInConfig()
    var cfg config.Config
    viper.Unmarshal(&cfg)


    ctx := context.Background()
    ctx = context.WithValue(ctx, "GLOBAL", cfg)
    ctx = context.WithValue(ctx, "LOGGER", logger)
    logger.Infof("%s", ctx)

    cmd.SetContext(ctx)
}
