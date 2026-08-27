package disc

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/util"

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

var dfltMode = config.DefaultConfig().Disc.Rip.Mode

func registerRipFlags(cmd *cobra.Command) {
	registerBackupFlag(cmd)
	registerRipOutputDirFlag(cmd)
	registerModeFlag(cmd)
	registerTemplateOverrideFlag(cmd)
}

func registerBackupFlag(cmd *cobra.Command) {
	util.RegisterOptionalStringFlag(
		cmd.Flags(),
		backup,
		fmt.Sprintf("Include a disc backup in this operation (optionally, specify a template for the destination directory: --%s=[target dir])", backup))
	util.AppendPreRun(cmd, func(cmd *cobra.Command, args []string) {
		backupFlag := cmd.Flags().Lookup(backup)
		if backupFlag != nil && backupFlag.Changed {
			viper.Set(config.DiscRipBackup, true)
			if o, ok := backupFlag.Value.(*util.OptionalString); ok && !o.WasEmpty {
				viper.Set(config.DiscBackupOutputDirTemplate, backupFlag.Value.String())
			}
		}
	})
}

func registerTemplateOverrideFlag(cmd *cobra.Command) {
	util.RegisterStringFlag(
		cmd.Flags(),
		templateOverride,
		config.TemplateOverride,
		"",
		"Provide a file naming template. This template will be used in place of any config-defined templates")
}

func registerRipOutputDirFlag(cmd *cobra.Command) {
	util.RegisterStringFlag(
		cmd.Flags(),
		outputDir,
		config.DiscRipOutputDirTemplate,
		"",
		"template for the output directory, e.g. ~/Videos/{{.Media.Title}} ({{.Disc.Year}})")
}

func registerModeFlag(cmd *cobra.Command) {
	util.RegisterVarFlag(
		cmd.Flags(),
		&dfltMode,
		mode,
		config.DiscRipMode,
		"Mode to execute in")
}
