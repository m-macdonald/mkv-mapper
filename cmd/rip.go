/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"m-macdonald/mkv-mapper/internal/app"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"

	"github.com/spf13/cobra"
)

// ripCmd represents the rip command
var ripCmd = &cobra.Command{
	Use:   "rip",
	Short: "Rips the current disc to .mkv and renames the output files",
	Long:  `The currently inserted disc is ripped to .mkv files and the resulting files are renamed in accordance with the naming pattern using values from TheDiscDB`,
	Run:   runRip,
}

func init() {
	rootCmd.AddCommand(ripCmd)
}

func runRip(cmd *cobra.Command, args []string) {
	ctx, ok := cmd.Context().Value(appContextKey).(AppContext)
	if !ok {
		panic(fmt.Errorf("failed to retrieve app context, unable to continue"))
	}

	makemkvClient := makemkv.NewClient(
		ctx.Config.MakeMkvPath,
		ctx.Logger.Named("makemkv"),
	)

	discdbClient := discdb.NewClient()

	pipeline := app.New(
		makemkvClient,
		discdbClient,
		ctx.Logger.Named("pipeline"))

	plan, err := pipeline.BuildPlan(ctx.Config.DiscRoot, ctx.Config.OutputDir, ctx.Config.Templates)
	if err != nil {
		ctx.Logger.Panicf("plan construction failed %w", err)
	}

	// TODO: Log the intended plan steps
	// TODO: Execute the plan
	ctx.Logger.Infof("%v", plan)

	err = pipeline.RunPlan(plan)
	if err != nil {
		ctx.Logger.Panicf("plan execution failed %w", err)
	}
}
