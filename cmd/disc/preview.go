/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
	"context"
	"os"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/terminal"
	"m-macdonald/mkv-mapper/internal/validation"

	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "View the planned result of a disc rip",
	Long:  ``,
	RunE:  runPreview,
}

func init() {
	Cmd.AddCommand(previewCmd)
	registerRipFlags(previewCmd)
}

func runPreview(cmd *cobra.Command, args []string) error {
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

	ctx := cmd.Context()
	if cfg.Disc.Rip.Backup {
		return runPreviewWithBackup(ctx, cfg, eng, renderers)
	}

	return runPreviewNoBackup(ctx, cfg, eng, renderers)
}

func runPreviewNoBackup(
	ctx context.Context,
	cfg config.Config,
	eng *engine.Engine,
	renderers renderers,
) error {
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

	return renderers.preview.Render(validatedPlan)
}

func runPreviewWithBackup(
	ctx context.Context,
	cfg config.Config,
	eng *engine.Engine,
	renderers renderers,
) error {
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

	return renderers.preview.Render(validatedPlan)
}
