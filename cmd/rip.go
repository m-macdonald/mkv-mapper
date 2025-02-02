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

// ripCmd represents the rip command
var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "Rips the current disc to .mkv and renames the output files",
	Long: `The currently inserted disc is ripped to .mkv files and the resulting files are renamed in accordance with the naming pattern using values from TheDiscDB`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("rip called")
	},
}

func init() {
	rootCmd.AddCommand(ripCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ripCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ripCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	ripCmd.Flags().Int("drive", 0, "The number of your optical drive as defined by makemkv (default is 0)")
    // TODO: Consolidate Viper configuration
    viper.BindPFlag("drive", rootCmd.Flags().Lookup("drive"))
}

func runRip(cmd *cobra.Command, args []string) {
    config, err := config.Load()
    if err != nil {
        fmt.Printf("%s", err)
    }

    titles, err := makemkv.ReadTitles(config.MakeMkvPath, config.DriveNum)
    // Second param is the disc number
    // Third is the slug of the title
    discDef, err := discdb.LoadDef(config.DiscDbDefs, 0, "slug")
}
