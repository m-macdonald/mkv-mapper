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
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/terminal"
	"m-macdonald/mkv-mapper/internal/util"
	"m-macdonald/mkv-mapper/internal/validation"

	"github.com/spf13/cobra"
)

const (
	backup     = "backup"
	keepBackup = "keep-backup"
)

var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "Rips the current disc to .mkv and renames the output files",
	Long:  `The currently inserted disc is ripped to .mkv files and the resulting files are renamed in accordance with the naming pattern using values from TheDiscDB`,
	RunE: runRip,
}

func init() {
	Cmd.AddCommand(ripCmd)
	registerRipFlags(ripCmd)
	util.RegisterBoolFlag(
		ripCmd.Flags(),
		keepBackup,
		config.DiscRipBackupKeep,
		false,
		"Retain the backup after the rip completes. Will be deleted otherwise. Does nothing if backup is not specified.")
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

	plan, err := buildRipPlan(ctx, eng, cfg, cfg.Disc.Root)
	if err != nil {
		return err
	}

	selectedPlan, err := eng.SelectPlan(cfg.Disc.Rip.Mode, plan)
	if err != nil {
		return err
	}

	needed := selectedPlan.SumTitleSizes()
	checks := []validation.CheckGroup{engine.RipChecks(selectedPlan, needed)}
	validatedPlan := eng.ValidateRipPlan(ctx, selectedPlan, checks)

	if err = renderers.preview.Render(validatedPlan); err != nil {
		return err
	}

	return eng.RunRipPlan(ctx, validatedPlan, onEvent)
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
	identity, discInfo, err := eng.ScanDisc(ctx, cfg.Disc.Root)
	if err != nil {
		return err
	}

	validatedBackupPlan, err := planBackup(ctx, eng, cfg, identity, discInfo)
	if err != nil {
		return err
	}

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

	if err := eng.RunBackupPlan(ctx, validatedBackupPlan, onEvent); err != nil {
		return fmt.Errorf("backing up disc: %w", err)
	}

	validatedRipPlan, err := merge(ctx, eng, cfg, validatedPlan, validatedBackupPlan)
	if err != nil {
		return err
	}

	if err := eng.RunRipPlan(ctx, validatedRipPlan, onEvent); err != nil {
		return err
	}

	return cleanupBackup(cfg, validatedPlan, validatedBackupPlan)
}

func cleanupBackup(
	cfg config.Config,
	ripPlan model.ValidatedRipPlan,
	backupPlan model.ValidatedBackupPlan,
) error {
	if cfg.Disc.Rip.KeepBackup {
		return nil
	}

	unsafe, err := files.IsAncestorOrEqual(backupPlan.OutputDir, ripPlan.OutputDir)
	if err != nil {
		return err
	}
	if unsafe {
		return fmt.Errorf("refusing to delete backup directory %q, it is the same as, or a parent of, the output directory %q", backupPlan.OutputDir, ripPlan.OutputDir)
	}

	return os.RemoveAll(backupPlan.OutputDir)
}

func buildRipPlan(
	ctx context.Context,
	eng *engine.Engine,
	cfg config.Config,
	discRoot string,
) (model.RipPlan, error) {
	return eng.BuildRipPlan(ctx, engine.BuildRipPlanConfig{
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
) (model.ValidatedRipPlan, error) {
	planCfg := engine.BuildRipPlanConfig{
		DiscRoot:  cfg.Disc.Root,
		Templates: cfg.Templates,
		Rip:       cfg.Disc.Rip,
	}
	plan, err := eng.CompleteRipPlan(ctx, identity, discInfo, planCfg)
	if err != nil {
		return model.ValidatedRipPlan{}, err
	}

	selected, err := eng.SelectPlan(cfg.Disc.Rip.Mode, plan)
	if err != nil {
		return model.ValidatedRipPlan{}, err
	}

	needed := selected.SumTitleSizes()
	checks := []validation.CheckGroup{
		engine.RipChecks(selected, needed),
	}
	return eng.ValidateRipPlan(ctx, selected, checks), nil
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
) (model.ValidatedBackupPlan, error) {
	backupCfg := engine.BuildBackupPlanConfig{
		OutputDirTemplate: cfg.Disc.Backup.OutputDirTemplate,
		KeepBackup:        cfg.Disc.Rip.KeepBackup,
	}
	backupPlan, err := eng.CompleteBackupPlan(identity, discInfo, backupCfg)
	if err != nil {
		return model.ValidatedBackupPlan{}, err
	}
	return validateBackupPlan(ctx, eng, backupPlan), nil
}

func merge(
	ctx context.Context,
	eng *engine.Engine,
	cfg config.Config,
	validatedRipPlan model.ValidatedRipPlan,
	validatedBackupPlan model.ValidatedBackupPlan,
) (model.ValidatedRipPlan, error) {
	plan, err := buildRipPlan(ctx, eng, cfg, validatedBackupPlan.OutputDir)
	if err != nil {
		return model.ValidatedRipPlan{}, err
	}

	merged, err := plan.MergeIntents(validatedRipPlan.Intents())
	if err != nil {
		return model.ValidatedRipPlan{}, err
	}

	// re-executing validation checks again.
	// This shouldn't be a problem, but if the checks ever become more expensive or add side-effects, this will need to be reworked.
	needed := merged.SumTitleSizes()
	mergedValidatedPlan := eng.ValidateRipPlan(ctx, merged, []validation.CheckGroup{
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
