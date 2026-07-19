package util

import "github.com/spf13/pflag"

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
