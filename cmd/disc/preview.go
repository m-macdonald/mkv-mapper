/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
	"os"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/display"

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

	plan, err := services.Engine.BuildPlan(
		cmd.Context(),
		cfg.DiscRoot,
		cfg.OutputDir,
		cfg.Templates)
	if err != nil {
		return err
	}
	validatedPlan := services.Engine.ValidatePlan(plan)

	previewRenderer := display.NewPreviewRenderer(os.Stdout)
	previewRenderer.Render(validatedPlan)

	return nil
}
