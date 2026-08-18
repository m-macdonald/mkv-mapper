package disc

import (
	"os"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/files"
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
	viper.BindPFlag(config.DiscBackupOutputDir, backupCmd.Flags().Lookup(outputDir))
}

func runBackup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	services, err := app.BuildServices(cfg)
	if err != nil {
		return err
	}
	eng := services.NewEngine(terminal.NewSelector())
	renderer := terminal.NewProgressRenderer(os.Stdout, true)

	return eng.Backup(
		cmd.Context(),
		files.DiscSource(cfg.DiscRoot),
		cfg.Disc.Backup.OutputDir,
		func(e event.Event) {
			err := renderer.HandleEvent(e)
			if err != nil {
				services.Logger.Warnw("renderer failed", "error", err)
			}
		})
}
