/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mapCmd represents the mp command
var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Renames existing .mkv files using TheDiscDB definitions.",
	Long: `map skips the ripping step and moves right to mapping existing .mkv files.
This command expects that your files have already been ripped and that you provide a path to a mapping file that tells it the names of the output files and the names of their original .mpls files.`,
	Run: runMap,
}

func init() {
	rootCmd.AddCommand(mapCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func runMap(cmd *cobra.Command, args []string) {
    fmt.Printf("Running map")
}
