package disc

import (
	"m-macdonald/mkv-mapper/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	discRoot         = "disc-root"
	makemkvPath      = "makemkv-path"
	mode             = "mode"
	outputDir        = "output-dir"
	templateOverride = "template-override"
)

var dfltMode = config.DefaultConfig().Disc.Mode

var Cmd = &cobra.Command{
	Use:   "disc",
	Short: "Commands for working with discs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlag(config.MakeMkvPath, cmd.Flags().Lookup(makemkvPath))
		viper.BindPFlag(config.TemplateOverride, cmd.Flags().Lookup(templateOverride))
		viper.BindPFlag(config.OutputDir, cmd.Flags().Lookup(outputDir))
		viper.BindPFlag(config.DiscRoot, cmd.Flags().Lookup(discRoot))
		viper.BindPFlag(config.DiscMode, cmd.Flags().Lookup(mode))
		return nil
	},
}

func init() {
	Cmd.PersistentFlags().String(discRoot, "", "Path to disc root")
	Cmd.PersistentFlags().String(outputDir, "", "Output directory")
	Cmd.PersistentFlags().String(makemkvPath, "makemkvcon", "The location of the makemkvcon binary. Defaults to assuming the binary is already available on the path")
	Cmd.PersistentFlags().String(templateOverride, "", "Provide a file naming template for this rip. This template will be used in place of any config-defined templates")
	Cmd.PersistentFlags().Var(&dfltMode, mode, "Mode to execute in")
}
