/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
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

	plan, err := eng.BuildRipPlan(
		cmd.Context(),
		engine.BuildRipPlanConfig{
			DiscRoot:          cfg.Disc.Root,
			Templates:         cfg.Templates,
			Rip:               cfg.Disc.Rip,
		},
	)
	if err != nil {
		return err
	}

	selectedPlan, err := eng.SelectPlan(cfg.Disc.Rip.Mode, plan)
	if err != nil {
		return err
	}

	checkGroups := []validation.CheckGroup{
		engine.RipChecks(selectedPlan, selectedPlan.SumTitleSizes()),
	}
	validatedPlan := eng.ValidateRipPlan(cmd.Context(), selectedPlan, checkGroups)

	previewRenderer := terminal.NewPreviewRenderer(os.Stdout)
	if err := previewRenderer.Render(validatedPlan); err != nil {
		return err
	}

	return nil
}
