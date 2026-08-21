package terminal

import (
	"fmt"
	"io"
	"m-macdonald/mkv-mapper/internal/validation"
	"os"

	"golang.org/x/term"
)

func DetectInteractiveOutput(out *os.File) bool {
	return term.IsTerminal(int(out.Fd()))
}

func renderCheckResults(
	out io.Writer,
	label validation.CheckGroupLabel,
	results []validation.Result,
) error {
	var discResults []validation.Result
	for _, result := range results {
		if result.RefID == "" {
			discResults = append(discResults, result)
		}
	}
	if len(discResults) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(out, "%s:\n", label); err != nil {
		return err
	}
	for _, result := range discResults {
		symbol := getValidationSymbol(result.Status)
		if _, err := fmt.Fprintf(out, " %s %s\n", symbol, result.Message); err != nil {
			return err
		}
	}
	return nil
}

func getValidationSymbol(status validation.Status) string {
	switch status {
	case validation.StatusPass:
		return "✓"
	case validation.StatusWarn:
		return "⚠"
	case validation.StatusFail:
		return "✗"
	}
	return ""
}
