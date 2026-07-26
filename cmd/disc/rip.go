/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/term"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/terminal"
	"m-macdonald/mkv-mapper/internal/util"
	"m-macdonald/mkv-mapper/internal/validation"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	backup     = "backup"
	keepBackup = "keep-backup"
)

var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "Rips the current disc to .mkv and renames the output files",
	Long:  `The currently inserted disc is ripped to .mkv files and the resulting files are renamed in accordance with the naming pattern using values from TheDiscDB`,
	PreRun: func(cmd *cobra.Command, args []string) {
		backupFlag := cmd.Flags().Lookup(backup)
		if backupFlag != nil && backupFlag.Changed {
			viper.Set(config.DiscRipBackup, true)
			if o, ok := backupFlag.Value.(*util.OptionalString); ok && !o.WasEmpty {
				viper.Set(config.DiscBackupOutputDir, backupFlag.Value.String())
			}
		}
		viper.BindPFlag(config.DiscRipBackupKeep, cmd.Flags().Lookup(keepBackup))
	},
	RunE: runRip,
}

func init() {
	Cmd.AddCommand(ripCmd)
	util.RegisterOptionalStringFlag(ripCmd.Flags(), backup, fmt.Sprintf("Backup disc before ripping (Optionally, specify a different destination for the backup: --%s={target dir})", backup))
	ripCmd.Flags().Bool(keepBackup, false, "Retain the backup after the rip completes. Will be deleted otherwise. Does nothing if backup is not specified.")
}

func runRip(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	services, err := app.BuildServices(cfg)
	if err != nil {
		return err
	}
	defer services.Close()

	eng := services.NewEngine(terminal.NewSelector())

	interactive := detectInteractiveOutput(os.Stdout)
	progressRenderer := terminal.NewProgressRenderer(os.Stdout, interactive)
	defer progressRenderer.Close()
	previewRenderer := terminal.NewPreviewRenderer(os.Stdout)

	onEvent := func(e event.Event) {
		err := progressRenderer.HandleEvent(e)
		if err != nil {
			services.Logger.Warnw("renderer failed", "error", err)
		}
	}

	if cfg.Disc.Rip.Backup {
		return runRipWithBackup(cmd, cfg, eng, onEvent, &previewRenderer)
	}

	return runRipNoBackup(cmd, cfg, eng, onEvent, &previewRenderer)
}

func runRipNoBackup(cmd *cobra.Command, cfg config.Config, eng *engine.Engine, onEvent engine.EngineEventSink, previewRenderer *terminal.PreviewRenderer) error {
	ctx := cmd.Context()

	plan, err := buildPlan(ctx, eng, cfg, cfg.DiscRoot)
	if err != nil {
		return err
	}

	selectedPlan, err := eng.SelectPlan(cfg.Disc.Mode, plan)
	if err != nil {
		return err
	}

	needed := selectedPlan.SumTitleSizes()
	checks := []validation.CheckGroup{engine.RipChecks(selectedPlan, needed)}
	validatedPlan := eng.ValidatePlan(ctx, selectedPlan, checks)

	if err = previewRenderer.Render(validatedPlan); err != nil {
		return err
	}

	return eng.RunPlan(ctx, validatedPlan, onEvent)
}

func runRipWithBackup(cmd *cobra.Command, cfg config.Config, eng *engine.Engine, onEvent engine.EngineEventSink, previewRenderer *terminal.PreviewRenderer) error {
	ctx := cmd.Context()

	plan, err := buildPlan(ctx, eng, cfg, cfg.DiscRoot)
	if err != nil {
		return err
	}

	selected, err := eng.SelectPlan(cfg.Disc.Mode, plan)
	if err != nil {
		return err
	}

	needed := selected.SumTitleSizes()
	// This is an intentional overestimation.
	// It's possible (likely) that we are counting the size of some segments more than once.
	// Better to overestimate than underestimate
	backupSize := plan.SumTitleSizes()
	checks := []validation.CheckGroup{
		engine.RipChecks(selected, needed),
		engine.BackupChecks(cfg.Disc.Backup.OutputDir, backupSize),
	}
	validatedPlan := eng.ValidatePlan(ctx, selected, checks)

	if err = previewRenderer.Render(validatedPlan); err != nil {
		return err
	}

	if err := eng.BackupPlanDisc(ctx, validatedPlan, cfg.Disc.Backup.OutputDir, onEvent); err != nil {
		return fmt.Errorf("backing up disc: %w", err)
	}

	ripPlan, err := buildPlan(ctx, eng, cfg, cfg.Disc.Backup.OutputDir)
	if err != nil {
		return err
	}

	merged, err := ripPlan.MergeIntents(selected.Intents())
	if err != nil {
		return err
	}

	// re-executing validation checks again.
	// This shouldn't be a problem, but if the checks ever become more expensive or add side-effects, this will need to be reworked.
	needed = merged.SumTitleSizes()
	mergedValidated := eng.ValidatePlan(ctx, merged, []validation.CheckGroup{
		engine.RipChecks(merged, needed),
	})

	if err := eng.RunPlan(ctx, mergedValidated, onEvent); err != nil {
		return err
	}

	return cleanupBackup(cfg)
}

func cleanupBackup(cfg config.Config) error {
	if cfg.Disc.Rip.KeepBackup {
		return nil
	}
	if cfg.Disc.Backup.OutputDir == cfg.OutputDir {
		return fmt.Errorf("refusing to delete backup directory, it is the same as the output directory")
	}
	return os.RemoveAll(cfg.Disc.Backup.OutputDir)
}

func buildPlan(ctx context.Context, eng *engine.Engine, cfg config.Config, discRoot string) (model.Plan, error) {
	return eng.BuildPlan(ctx, engine.BuildPlanConfig{
		OutputDir: cfg.OutputDir,
		DiscRoot:  discRoot,
		Templates: cfg.Templates,
		Rip:       cfg.Disc.Rip,
	})
}

// TODO: Centralize this check when multiple commands are implemented
func detectInteractiveOutput(out *os.File) bool {
	return term.IsTerminal(int(out.Fd()))
}
