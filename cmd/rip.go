/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
    "os"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
    "github.com/alexeyco/simpletable"
)

// ripCmd represents the rip command
var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "Rips the current disc to .mkv and renames the output files",
	Long: `The currently inserted disc is ripped to .mkv files and the resulting files are renamed in accordance with the naming pattern using values from TheDiscDB`,
	Run: runRip, 
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
    ripCmd.Flags().String("slug", "", "The slug of the disc as defined in TheDiscDB")
    // TODO: Consolidate Viper configuration
    viper.BindPFlag("drive", ripCmd.Flags().Lookup("drive"))
    viper.BindPFlag("slug", ripCmd.Flags().Lookup("slug"))
}

func runRip(cmd *cobra.Command, args []string) {
    var cfg config.Config
    if val, ok := cmd.Context().Value("GLOBAL").(config.Config); !ok {
        // TODO: This needs proper error messaging
        fmt.Printf("%v", cmd.Context())
        os.Exit(1)
    } else {
        cfg = val
    }

    titles, err := makemkv.ReadTitles(cfg.MakeMkvPath, cfg.DriveNum)
    if err != nil {
        // TODO: Do Something with the error
        fmt.Sprintln("%s", err)
    }
    fmt.Printf("%v", titles)
    for mplsFile, _ := range titles {
        fmt.Println("Makemkv mplsFile names")
        fmt.Printf("%s: [% x] %U\n", mplsFile, []byte(mplsFile), []rune(mplsFile))
    }
    // Second param is the disc number
    // Third is the slug of the title
    discDef, err := discdb.LoadDef(cfg.DiscDbDefs, 1, "/series/Black Sails (2014)/2018-complete-collection-blu-ray")
    if err != nil {
        fmt.Sprintln("%s", err)
    }
    for mplsFile, _ := range discDef {
        fmt.Println("DiscDB mplsFile names")
        fmt.Printf("%s: [% x] %U\n", mplsFile, []byte(mplsFile), []rune(mplsFile))
    }

    fmt.Printf("%v", discDef)
    mappings := make(map[string]discdb.SummaryTitle)
    for mplsFile, summary := range discDef {
        fmt.Printf("Mpls: %s", mplsFile)
        if mapped, ok := titles[mplsFile]; !ok {
            // TODO: Inform the user things are not ok
            fmt.Printf("Mapping failed for: %s\n", mplsFile)
        } else {
            fmt.Printf("mapping: %v\n", mapped)
            // Need a check to make sure that we were actually able to retrieve "mapped" and it's not nil
            mappings[mapped] = summary
        }
    }

    table := simpletable.New() 

    table.Header = &simpletable.Header{
        Cells: []*simpletable.Cell{
            {Align: simpletable.AlignCenter, Text: "MakeMkv Output File"},
            {Align: simpletable.AlignCenter, Text: ".mpls File"},
            {Align: simpletable.AlignCenter, Text: "TheDiscDB File Name"},
        },
    }

    for outputFile, mapping := range mappings {
        r := []*simpletable.Cell{
            {Text: fmt.Sprintf("%s", outputFile)},
            {Text: fmt.Sprintf("%s", mapping.SourceFileName)},
            {Text: fmt.Sprintf("%s", mapping.FileName)},
        }

        table.Body.Cells = append(table.Body.Cells, r)
    }

    table.SetStyle(simpletable.StyleCompactClassic)
    fmt.Println(table.String())
}
