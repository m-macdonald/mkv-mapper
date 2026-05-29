package disc

import (
	"m-macdonald/mkv-mapper/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	makemkvPath      = "makemkv-path"
	templateOverride = "template-override"
	outputDir        = "output-dir"
	discRoot         = "disc-root"
	hideUnmatched    = "hideUnmatched"
)

var Cmd = &cobra.Command{
	Use:   "disc",
	Short: "Commands for working with discs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlag(config.MakeMkvPath, cmd.Flags().Lookup(makemkvPath))
		viper.BindPFlag(config.TemplateOverride, cmd.Flags().Lookup(templateOverride))
		viper.BindPFlag(config.OutputDir, cmd.Flags().Lookup(outputDir))
		viper.BindPFlag(config.DiscRoot, cmd.Flags().Lookup(discRoot))

		return nil
	},
}

func init() {
	Cmd.PersistentFlags().String(discRoot, "", "Path to disc root")
	Cmd.PersistentFlags().String(outputDir, "", "Output directory")
	Cmd.PersistentFlags().Bool(hideUnmatched, false, "Hide titles with no DiscDB metadata")
	Cmd.PersistentFlags().String(makemkvPath, "makemkvcon", "The location of the makemkvcon binary. Defaults to assuming the binary is already available on the path")
	Cmd.PersistentFlags().String(templateOverride, "", "Provide a file naming template for this rip. This template will be used in place of any config-defined templates")
}
