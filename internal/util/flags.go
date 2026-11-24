package util

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const viperKeyAnnotation = "viper-key"

func annotateConfigKey(flags *pflag.FlagSet, viperKey, flagName string) {
	// Sets an annotation on the flag containing the Viper key to bind to.
	// This allows us to tie the Viper key to the flag without actually binding it.
	// We delay binding it until BindConfigFlags is called, after we've determined the command that is running.
	// Binding here instead would be wrong. Viper holds only one flag per key, so the last command registered
	// would take the key from every other command whose flag shares it.
	if err := flags.SetAnnotation(flagName, viperKeyAnnotation, []string{viperKey}); err != nil {
		panic(fmt.Sprintf("annotating %q with %q: %v", flagName, viperKey, err))
	}
}

// BindConfigFlags binds each of cmd's flags that was registered with a config key to that key in viper.
//
// Call it once, for the command being executed, before reading config.
// Viper allows only one flag per key, so flags from different commands that share a key will only bind the last encountered flag.
func BindConfigFlags(cmd *cobra.Command) error {
	var err error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		keys := flag.Annotations[viperKeyAnnotation]
		if err != nil || len(keys) == 0 {
			return
		}
		if bindErr := viper.BindPFlag(keys[0], flag); bindErr != nil {
			err = fmt.Errorf("binding flag %q to %q: %w", flag.Name, keys[0], bindErr)
		}
	})
	return err
}

// WithUsageHint wraps err with a pointer to cmd's help text.
// Use it for errors caused by how the command was invoked (bad flags, invalid flag combinations),
// since full usage output is silenced on the root command.
func WithUsageHint(cmd *cobra.Command, err error) error {
	return fmt.Errorf("%w\nRun '%s --help' for usage.", err, cmd.CommandPath())
}

// AppendPreRunE chains fn after any PreRunE already set on cmd, stopping at the first error.
// Use carefully. Cobra ignores a PreRun defined on a command that has a PreRunE,
// so this func could cause PreRun funcs to be unintentionally skipped.
func AppendPreRunE(cmd *cobra.Command, fn func(cmd *cobra.Command, args []string) error) {
	existing := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if existing != nil {
			if err := existing(cmd, args); err != nil {
				return err
			}
		}
		return fn(cmd, args)
	}
}

func RegisterStringFlag(flags *pflag.FlagSet, flagName string, viperKey string, defaultValue string, usage string) {
	flags.String(flagName, defaultValue, usage)
	annotateConfigKey(flags, viperKey, flagName)
}

func RegisterBoolFlag(flags *pflag.FlagSet, flagName string, viperKey string, defaultValue bool, usage string) {
	flags.Bool(flagName, defaultValue, usage)
	annotateConfigKey(flags, viperKey, flagName)
}

func RegisterVarFlag(flags *pflag.FlagSet, value pflag.Value, flagName string, viperKey string, usage string) {
	flags.Var(value, flagName, usage)
	annotateConfigKey(flags, viperKey, flagName)
}
