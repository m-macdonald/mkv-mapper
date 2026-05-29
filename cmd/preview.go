/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"m-macdonald/mkv-mapper/internal/app"
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
	rootCmd.AddCommand(previewCmd)
}

func runPreview(cmd *cobra.Command, args []string) error {
	ctx, ok := cmd.Context().Value(app.AppContextKey).(app.AppContext)
	if !ok {
		panic(fmt.Errorf("failed to retrieve app context, unable to continue"))
	}
	services, err := app.BuildServices(ctx)
	if err != nil {
		return err
	}
	defer services.Close()

	plan, err := services.Engine.BuildPlan(
		cmd.Context(),
		ctx.Config.DiscRoot,
		ctx.Config.OutputDir,
		ctx.Config.Templates)
	if err != nil {
		return err
	}
	validatedPlan := services.Engine.ValidatePlan(plan)

	previewRenderer := display.NewPreviewRenderer(os.Stdout)
	previewRenderer.Render(validatedPlan)

	return nil
}
