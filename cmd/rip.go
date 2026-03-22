/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/event"

	"github.com/spf13/cobra"
)

var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "Rips the current disc to .mkv and renames the output files",
	Long:  `The currently inserted disc is ripped to .mkv files and the resulting files are renamed in accordance with the naming pattern using values from TheDiscDB`,
	RunE:  runRip,
}

func init() {
	rootCmd.AddCommand(ripCmd)
}

func runRip(cmd *cobra.Command, args []string) error {
	ctx, ok := cmd.Context().Value(app.AppContextKey).(app.AppContext)
	if !ok {
		panic(fmt.Errorf("failed to retrieve app context, unable to continue"))
	}
	services, err := app.BuildServices(ctx)
	if err != nil {
		return err
	}
	defer services.Close()

	ripPreview, err := services.Ripper.PreviewRip(
		cmd.Context(),
		ctx.Config.DiscRoot,
		ctx.Config.OutputDir,
		ctx.Config.Templates)
	if err != nil {
		return err
	}

	if len(ripPreview.ValidationReport.Errors) > 0 {
		// TODO: Handle the potentially multiple errors within ValidationReport
		for _, err := range ripPreview.ValidationReport.Errors {
			ctx.Logger.Error(err)
		}
		return fmt.Errorf("validation failed")
	}

	renderer := event.NewRenderer(os.Stdout)
	defer renderer.Close()

	// TODO: Log the intended plan steps and any warnings from the ValidationReport
	err = services.Ripper.ExecuteRip(
		cmd.Context(),
		ripPreview.Plan,
		func(e event.Event) {
			err := renderer.HandleEvent(e)
			if err != nil {
				ctx.Logger.Warnw("renderer failed", "error", err)
			}
		})
	if err != nil {
		return err
	}

	return nil
}
