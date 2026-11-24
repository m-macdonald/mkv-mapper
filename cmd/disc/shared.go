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
	registerBackupFlags(cmd)
	registerRipOutputDirFlag(cmd)
	registerModeFlag(cmd)
	registerTemplateOverrideFlag(cmd)
}

func registerBackupFlags(cmd *cobra.Command) {
	util.RegisterBoolFlag(
		cmd.Flags(),
		backup,
		config.DiscRipBackup,
		false,
		"Include a disc backup in this operation")
	util.RegisterStringFlag(
		cmd.Flags(),
		backupDir,
		config.DiscBackupOutputDirTemplate,
		"",
		fmt.Sprintf("template for the backup directory, e.g. ~/Videos/backup/{{.Disc.Label}}. Requires --%s", backup))
	util.AppendPreRunE(cmd, requireBackup)
}

// Checks the resolved config rather than the --backup flag, so a config-enabled backup also satisfies it.
func requireBackup(cmd *cobra.Command, args []string) error {
	if viper.GetBool(config.DiscRipBackup) {
		return nil
	}
	for _, name := range []string{backupDir, keepBackup} {
		if cmd.Flags().Changed(name) {
			return util.WithUsageHint(cmd, fmt.Errorf("--%s requires --%s", name, backup))
		}
	}
	return nil
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
