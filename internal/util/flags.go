package util

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type OptionalString struct {
	Value    string
	WasEmpty bool
}

const emptyMarker = "\x00"

func (o *OptionalString) String() string {
	return o.Value
}

func (o *OptionalString) Set(s string) error {
	o.WasEmpty = s == emptyMarker
	if !o.WasEmpty {
		o.Value = s
	}
	return nil
}

func (o *OptionalString) Type() string {
	return "string"
}

func RegisterOptionalStringFlag(flagSet *pflag.FlagSet, name, usage string) {
	o := &OptionalString{}
	flagSet.Var(o, name, usage)
	flagSet.Lookup(name).NoOptDefVal = emptyMarker
}

func bindConfigFlag(flags *pflag.FlagSet, viperKey, flagName string) {
	if err := viper.BindPFlag(viperKey, flags.Lookup(flagName)); err != nil {
		panic(fmt.Sprintf("bindConfigFlag: binding %q to %q: %v", flagName, viperKey, err))
	}
}

func RegisterStringFlag(flags *pflag.FlagSet, flagName string, viperKey string, defaultValue string, usage string) {
	flags.String(flagName, defaultValue, usage)
	bindConfigFlag(flags, viperKey, flagName)
}

func RegisterBoolFlag(flags *pflag.FlagSet, flagName string, viperKey string, defaultValue bool, usage string) {
	flags.Bool(flagName, defaultValue, usage)
	bindConfigFlag(flags, viperKey, flagName)
}

func RegisterVarFlag(flags *pflag.FlagSet, value pflag.Value, flagName string, viperKey string, usage string) {
	flags.Var(value, flagName, usage)
	bindConfigFlag(flags, viperKey, flagName)
}
