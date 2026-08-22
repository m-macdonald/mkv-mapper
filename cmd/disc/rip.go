/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
	"context"
	"fmt"
	"io"
	"os"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/makemkv"
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

	out := os.Stdout
	renderers := newRenderers(out, terminal.DetectInteractiveOutput(out))
	defer renderers.close()

	onEvent := func(e event.Event) {
		err := renderers.progress.HandleEvent(e)
		if err != nil {
			services.Logger.Warnw("renderer failed", "error", err)
		}
	}

	if cfg.Disc.Rip.Backup {
		return runRipWithBackup(cmd, cfg, eng, onEvent, renderers)
	}

	return runRipNoBackup(cmd, cfg, eng, onEvent, renderers)
}

func runRipNoBackup(
	cmd *cobra.Command,
	cfg config.Config,
	eng *engine.Engine,
	onEvent engine.EngineEventSink,
	renderers renderers,
) error {
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

	if err = renderers.preview.Render(validatedPlan); err != nil {
		return err
	}

	return eng.RunPlan(ctx, validatedPlan, onEvent)
}

func runRipWithBackup(
	cmd *cobra.Command,
	cfg config.Config,
	eng *engine.Engine,
	onEvent engine.EngineEventSink,
	renderers renderers,
) error {
	ctx := cmd.Context()

	// Scan disc so that we can reuse the scan for plan and backupPlan construction.
	// Slightly reduces wear on the disc by allowing us to scan only once.
	identity, discInfo, err := eng.ScanDisc(ctx, cfg.DiscRoot)
	if err != nil {
		return err
	}

	validatedBackupPlan := planBackup(ctx, eng, cfg, identity, discInfo)

	if err := renderers.backup.Render(validatedBackupPlan); err != nil {
		return err
	}

	validatedPlan, err := planRip(ctx, cfg, eng, identity, discInfo)
	if err != nil {
		return err
	}

	if err = renderers.preview.Render(validatedPlan); err != nil {
		return err
	}

	if err := eng.BackupPlanDisc(ctx, validatedBackupPlan, onEvent); err != nil {
		return fmt.Errorf("backing up disc: %w", err)
	}

	validatedRipPlan, err := merge(ctx, eng, cfg, validatedPlan)
	if err != nil {
		return err
	}

	if err := eng.RunPlan(ctx, validatedRipPlan, onEvent); err != nil {
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

func buildPlan(
	ctx context.Context,
	eng *engine.Engine,
	cfg config.Config,
	discRoot string,
) (model.Plan, error) {
	return eng.BuildPlan(ctx, engine.BuildPlanConfig{
		OutputDir: cfg.OutputDir,
		DiscRoot:  discRoot,
		Templates: cfg.Templates,
		Rip:       cfg.Disc.Rip,
	})
}

func planRip(
	ctx context.Context,
	cfg config.Config,
	eng *engine.Engine,
	identity model.DiscIdentity,
	discInfo makemkv.DiscInfo,
) (model.ValidatedPlan, error) {
	planCfg := engine.BuildPlanConfig{
		OutputDir: cfg.OutputDir,
		DiscRoot:  cfg.DiscRoot,
		Templates: cfg.Templates,
		Rip:       cfg.Disc.Rip,
	}
	plan, err := eng.CompletePlan(ctx, identity, discInfo, planCfg)
	if err != nil {
		return model.ValidatedPlan{}, err
	}

	selected, err := eng.SelectPlan(cfg.Disc.Mode, plan)
	if err != nil {
		return model.ValidatedPlan{}, err
	}

	needed := selected.SumTitleSizes()
	checks := []validation.CheckGroup{
		engine.RipChecks(selected, needed),
	}
	return eng.ValidatePlan(ctx, selected, checks), nil
}

func validateBackupPlan(
	ctx context.Context,
	eng *engine.Engine,
	plan model.BackupPlan,
) model.ValidatedBackupPlan {
	// This is an intentional overestimation.
	// It's possible (likely) that we are counting the size of some segments more than once.
	// Better to overestimate than underestimate
	backupSize := plan.SumTitleSizes()
	backupChecks := []validation.CheckGroup{
		engine.BackupChecks(plan.OutputDir, backupSize),
	}
	return eng.ValidateBackupPlan(ctx, plan, backupChecks)
}

func planBackup(
	ctx context.Context,
	eng *engine.Engine,
	cfg config.Config,
	identity model.DiscIdentity,
	discInfo makemkv.DiscInfo,
) model.ValidatedBackupPlan {
	backupCfg := engine.BuildBackupPlanConfig{
		OutputDir:  cfg.OutputDir,
		KeepBackup: cfg.Disc.Rip.KeepBackup,
	}
	backupPlan := eng.CompleteBackupPlan(identity, discInfo, backupCfg)
	return validateBackupPlan(ctx, eng, backupPlan)
}

func merge(
	ctx context.Context,
	eng *engine.Engine,
	cfg config.Config,
	validatedPlan model.ValidatedPlan,
) (model.ValidatedPlan, error) {
	plan, err := buildPlan(ctx, eng, cfg, cfg.Disc.Backup.OutputDir)
	if err != nil {
		return model.ValidatedPlan{}, err
	}

	merged, err := plan.MergeIntents(validatedPlan.Intents())
	if err != nil {
		return model.ValidatedPlan{}, err
	}

	// re-executing validation checks again.
	// This shouldn't be a problem, but if the checks ever become more expensive or add side-effects, this will need to be reworked.
	needed := merged.SumTitleSizes()
	mergedValidatedPlan := eng.ValidatePlan(ctx, merged, []validation.CheckGroup{
		engine.RipChecks(merged, needed),
	})
	return mergedValidatedPlan, nil
}

type renderers struct {
	backup   terminal.BackupSummaryRenderer
	preview  terminal.PreviewRenderer
	progress terminal.ProgressRenderer
}

func newRenderers(out io.Writer, interactive bool) renderers {
	return renderers{
		backup:   terminal.NewBackupSummaryRenderer(out),
		preview:  terminal.NewPreviewRenderer(out),
		progress: *terminal.NewProgressRenderer(out, interactive),
	}
}

func (r *renderers) close() error {
	return r.progress.Close()
}
