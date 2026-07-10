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

	engineService := services.NewEngine(terminal.NewSelector())

	plan, err := engineService.BuildPlan(
		cmd.Context(),
		engine.BuildPlanConfig{
			OutputDir: cfg.OutputDir,
			DiscRoot:  cfg.DiscRoot,
			Templates: cfg.Templates,
			Rip:       cfg.Disc.Rip,
		},
	)
	if err != nil {
		return err
	}
	selectedPlan, err := engineService.SelectPlan(cfg.Disc.Mode, plan)
	if err != nil {
		return err
	}
	validatedPlan := engineService.ValidatePlan(selectedPlan)

	previewRenderer := terminal.NewPreviewRenderer(os.Stdout)
	previewRenderer.Render(validatedPlan)

	return nil
}
