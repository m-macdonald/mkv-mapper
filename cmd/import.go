package cmd

import (
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var importCmd = &cobra.Command {
	Use: "import",
	Short: "Imports The Disc DB database",
	Run: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().String("repo", "", "The DiscDB repo to clone")
}

func runImport(cmd *cobra.Command, args []string) {
	logger, ok := cmd.Context().Value("LOGGER").(*zap.SugaredLogger)
	if !ok {
		panic("Failed to retrieve logger from context. Unable to continue.")
	}

	cfg, ok := cmd.Context().Value("GLOBAL").(config.Config)

	logger.Infoln(cfg)

	err := discdb.Index()
	if err != nil {
		logger.Panicln("Import failed", err)
	} else {
		logger.Infoln("Import Complete")
	}
}
