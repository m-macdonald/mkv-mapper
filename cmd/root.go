/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
    "fmt"
	"os"

	"github.com/spf13/cobra"
    "github.com/spf13/viper"
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
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var defaultCfgFile = "$HOME/.config/mkv-mapper.json"

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
    var debug, cfgFile string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfgFile, fmt.Sprintf("config file (default is %s)", defaultCfgFile))
    rootCmd.PersistentFlags().StringVar(&debug, "debug", "", "")
        
    configureViper()
}

func configureViper() {
    viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}
