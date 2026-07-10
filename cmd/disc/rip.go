/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package disc

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
			if o, ok := backupFlag.Value.(*optionalString); ok && !o.wasEmpty {
				viper.Set(config.DiscRipBackupDir, backupFlag.Value.String())
			}
		}
		viper.BindPFlag(config.DiscRipKeepBackup, cmd.Flags().Lookup(keepBackup))
	},
	RunE: runRip,
}

func init() {
	Cmd.AddCommand(ripCmd)
	registerOptionalStringFlag(ripCmd.Flags(), backup, fmt.Sprintf("Backup disc before ripping (Optionally, specify a different destination for the backup: --%s={target dir})", backup))
	ripCmd.Flags().Bool(keepBackup, false, "Retain the backup after the rip completes. Will be deleted otherwise. Does nothing if backup is not specified.")
}

type optionalString struct {
	value    string
	wasEmpty bool
}

const emptyMarker = "\x00"

func (o *optionalString) String() string {
	return o.value
}

func (o *optionalString) Set(s string) error {
	o.wasEmpty = s == emptyMarker
	if !o.wasEmpty {
		o.value = s
	}
	return nil
}

func (o *optionalString) Type() string {
	return "string"
}

func registerOptionalStringFlag(flagSet *pflag.FlagSet, name, usage string) {
	o := &optionalString{}
	flagSet.Var(o, name, usage)
	flagSet.Lookup(name).NoOptDefVal = emptyMarker
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
	validatedPlan := engineService.ValidatePlan(selectedPlan)

	previewRenderer := terminal.NewPreviewRenderer(os.Stdout)
	previewRenderer.Render(validatedPlan)

	interactive := detectInteractiveOutput(os.Stdout)
	renderer := terminal.NewProgressRenderer(os.Stdout, interactive)
	defer renderer.Close()

	err = engineService.RunPlan(
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
