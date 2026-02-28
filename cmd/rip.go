/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/files"

	"github.com/alexeyco/simpletable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	// TODO: Option to rip a backup that was taken previously
	ripCmd.Flags().String("disc-root", "", "")
	ripCmd.Flags().Int("drive", 1, "The number of your optical drive as defined by makemkv (default is 0)")
    // ripCmd.Flags().String("slug", "", "The slug of the disc as defined in TheDiscDB")
    // TODO: Consolidate Viper configuration
    viper.BindPFlag("drive", ripCmd.Flags().Lookup("drive"))
}

func runRip(cmd *cobra.Command, args []string) {
    logger, ok := cmd.Context().Value("LOGGER").(*zap.SugaredLogger)
    if !ok {
        panic("Failed to retrieve logger from context. Unable to continue.")
    }
    
    cfg, ok := cmd.Context().Value("GLOBAL").(config.Config)
    if !ok {
        logger.Panicln("Failed to retrieve global config from context. Unable to continue.", "context", cmd.Context())
    }

	discRoot, err := cmd.Flags().GetString("disc-root")
	if err != nil {
		logger.Panicln("Failed to parse disc-root flag", err)
	}

	root, err := files.ResolveDiscRoot(discRoot)
	if err != nil {
		logger.Panicln("Unable to find disc root", err)
	}
	hash, err := files.Hash(root)
	if err != nil {
		logger.Panicln("Unable to hash disc", err)
	}

	logger.Debugln("Calculated hash", hash)
	disc, err := discdb.GetDisc(hash)
    if err != nil {
        logger.Panicln("Failed to retrieve disc definitions from TheDiscDB", err)
    }

	logger.Infoln(disc)
	titles, err := makemkv.ReadTitles(logger, cfg.MakeMkvPath)
    if err != nil {
        logger.Panicln("Unable to read disc titles using MakeMkv", err)
    }
    // For now just going to write the titles to log. Maybe format this a bit better in the future
    logger.Debug("MakeMkv Titles", titles)


    mappings := make(map[string]discdb.Title)
    for playlistFile, outputName := range titles {
		for _, title := range disc.Titles {
			if title.SourceFile == playlistFile {
				mappings[outputName] = title
			}
		}
        // if mapped, ok := discDef[mplsFile]; !ok {
        //     logger.Warnf("Failed to map %s to a DiscDB definition\n", mplsFile)
        // } else {
        //     logger.Debugf("Mapped %s to DiscDB definition %v\n", outputName, mapped)
        //     mappings[outputName] = mapped 
        // }
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
            {Text: fmt.Sprintf("%s", mapping.SourceFile)},
        }

        table.Body.Cells = append(table.Body.Cells, r)
    }

    table.SetStyle(simpletable.StyleCompactClassic)
    logger.Infoln("Mappings")
    logger.Infoln(table.String())

    logger.Infoln("Beginning disc rip...")
    // makemkv.RipDisc(logger, cfg.MakeMkvPath, cfg.DriveNum, cfg.MkvDest)
    // mapErrors := mapper.Map(cfg.MkvDest, mappings)
    // if mapErrors != nil {
    //     logger.Errorln("Error(s) while mapping .mpls files")
    // } else {
    //     logger.Infoln("Mapping complete")
    // }
}
