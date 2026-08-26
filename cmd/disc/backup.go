package disc

import (
	"os"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backs up the cuurrently inserted disc to local storage",
	Long:  "Copies the contents of the currently inserted optical disc to a local directory, without ripping or renaming titles.",
	RunE:  runBackup,
}

func init() {
	Cmd.AddCommand(backupCmd)
	backupCmd.Flags().String(discRoot, "", "Disc root")
	backupCmd.Flags().String(outputDir, "", "Directory to back up the disc into")
	viper.BindPFlag(config.DiscRoot, backupCmd.Flags().Lookup(discRoot))
	viper.BindPFlag(config.DiscBackupOutputDirTemplate, backupCmd.Flags().Lookup(outputDir))
}

func runBackup(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	services, err := app.BuildServices(cfg)
	if err != nil {
		return err
	}
	eng := services.NewEngine(terminal.NewSelector())

	identity, discInfo, err := eng.ScanDisc(ctx, cfg.Disc.Root)
	if err != nil {
		return err
	}

	validatedBackupPlan, err := planBackup(ctx, eng, cfg, identity, discInfo)
	if err != nil {
		return err
	}

	out := os.Stdout
	progressRenderer := terminal.NewProgressRenderer(out, terminal.DetectInteractiveOutput(out))
	defer progressRenderer.Close()
	backupRenderer := terminal.NewBackupSummaryRenderer(out)

	if err := backupRenderer.Render(validatedBackupPlan); err != nil {
		return err
	}

	return eng.RunBackupPlan(
		ctx,
		validatedBackupPlan,
		func(e event.Event) {
			err := progressRenderer.HandleEvent(e)
			if err != nil {
				services.Logger.Warnw("renderer failed", "error", err)
			}
		})
}
