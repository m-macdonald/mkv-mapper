package disc

import (
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/util"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "disc",
	Short: "Commands for working with discs",
}

func init() {
	util.RegisterStringFlag(
		Cmd.PersistentFlags(),
		discRoot,
		config.DiscRoot,
		"",
		"Path to disc root")
	dfltMakemkvPath := config.DefaultConfig().MakeMkvPath
	util.RegisterStringFlag(
		Cmd.PersistentFlags(),
		makemkvPath,
		config.MakeMkvPath,
		dfltMakemkvPath,
		"The location of the makemkvcon binary. Defaults to assuming the binary is already available on the path.")
}
