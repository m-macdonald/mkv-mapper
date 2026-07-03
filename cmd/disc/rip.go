/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
	"os"

	"golang.org/x/term"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/terminal"
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
	Cmd.AddCommand(ripCmd)
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

	engine := services.NewEngine(terminal.NewSelector())

	plan, err := engine.BuildPlan(
		cmd.Context(),
		cfg.DiscRoot,
		cfg.OutputDir,
		cfg.Templates)
	if err != nil {
		return err
	}
	selectedPlan, err := engine.SelectPlan(cfg.Disc.Mode, plan)
	validatedPlan := engine.ValidatePlan(selectedPlan)

	previewRenderer := terminal.NewPreviewRenderer(os.Stdout);
	previewRenderer.Render(validatedPlan)

	interactive := detectInteractiveOutput(os.Stdout)
	renderer := terminal.NewProgressRenderer(os.Stdout, interactive)
	defer renderer.Close()

	err = engine.RunPlan(
		cmd.Context(),
		validatedPlan,
		func(e event.Event) {
			err := renderer.HandleEvent(e)
			if err != nil {
				services.Logger.Warnw("renderer failed", "error", err)
			}
		})
	if err != nil {
		return err
	}

	return nil
}

// TODO: Centralize this check when multiple commands are implemented
func detectInteractiveOutput(out *os.File) bool {
	return term.IsTerminal(int(out.Fd()))
}
