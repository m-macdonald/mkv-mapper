package config

import "fmt"

type SelectionMode string

const (
	ModeFullAuto    SelectionMode = "full-auto"
	ModeTrimmedAuto SelectionMode = "trimmed-auto"
	ModeManual      SelectionMode = "manual"
)

func (mode SelectionMode) Valid() bool {
	switch mode {
	case ModeFullAuto, ModeTrimmedAuto, ModeManual:
		return true
	default:
		return false
	}
}

func ParseSelectionMode(str string) (SelectionMode, error) {
	mode := SelectionMode(str)
	if !mode.Valid() {
		return "", fmt.Errorf("invalid mode %q (must be one of: %s, %s, %s)", mode, ModeFullAuto, ModeTrimmedAuto, ModeManual)
	}
	return mode, nil
}

// Satisfy Cobra flag interface
func (m *SelectionMode) Set(str string) error {
	parsed, err := ParseSelectionMode(str)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

func (m SelectionMode) String() string {
	return string(m)
}

func (m SelectionMode) Type() string {
	return "SelectionMode"
}
