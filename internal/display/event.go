package display

import (
	"fmt"
	"io"
	"strings"

	"m-macdonald/mkv-mapper/internal/event"
)

type ProgressRenderer struct {
	out         io.Writer
	interactive bool

	currentPercent float64
	totalPercent   float64

	currentMessage string
	totalMessage   string

	statusVisible   bool
	statusLineCount int
}

func NewProgressRenderer(out io.Writer, interactive bool) *ProgressRenderer {
	return &ProgressRenderer{
		out:         out,
		interactive: interactive,
	}
}

func (p *ProgressRenderer) HandleEvent(ev event.Event) error {
	switch e := ev.(type) {
	case event.MessageEvent:
		return p.handleMessage(e)
	case event.ProgressPercentEvent:
		return p.handleProgressPercent(e)
	case event.ProgressCurrentEvent:
		return p.handleProgressCurrent(e)
	case event.ProgressTotalEvent:
		return p.handleProgressTotal(e)
	default:
		return nil
	}
}

func (p *ProgressRenderer) handleMessage(ev event.MessageEvent) error {
	if p.interactive && p.statusVisible {
		if err := p.clearStatus(); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(p.out, ev.Message+"\n"); err != nil {
		return err
	}

	if p.interactive && p.hasStatusContent() {
		return p.redrawStatus()
	}

	return nil
}

func (p *ProgressRenderer) handleProgressPercent(ev event.ProgressPercentEvent) error {
	if !p.interactive {
		return nil
	}

	p.currentPercent = ev.CurrentPercent
	p.totalPercent = ev.TotalPercent

	return p.redrawStatus()
}

func (p *ProgressRenderer) handleProgressCurrent(ev event.ProgressCurrentEvent) error {
	if !p.interactive {
		return nil
	}

	p.currentMessage = ev.Message

	return p.redrawStatus()
}

func (p *ProgressRenderer) handleProgressTotal(ev event.ProgressTotalEvent) error {
	if !p.interactive {
		return nil
	}

	p.totalMessage = ev.Message

	return p.redrawStatus()
}

func (p *ProgressRenderer) Close() error {
	if p.interactive && p.statusVisible {
		return p.clearStatus()
	}

	return nil
}

func (r *ProgressRenderer) hasStatusContent() bool {
	return r.currentPercent > 0 || r.totalPercent > 0
}

func (p *ProgressRenderer) redrawStatus() error {
	statusLines := p.statusLines()

	if p.statusVisible {
		if err := p.clearStatus(); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(p.out, strings.Join(statusLines, "\n")); err != nil {
		return err
	}

	p.statusVisible = true
	p.statusLineCount = len(statusLines)

	return nil
}

func (p *ProgressRenderer) clearStatus() error {
	if !p.statusVisible {
		return nil
	}

	// Move the cursor to the top of the status block
	if _, err := fmt.Fprintf(p.out, "\033[%dA", p.statusLineCount); err != nil {
		return err
	}

	for i := range p.statusLineCount {
		// Clear the cursor's current line
		if _, err := fmt.Fprintf(p.out, "\r\033[K"); err != nil {
			return err
		}

		if i < p.statusLineCount-1 {
			// Move cursor down a line
			if _, err := fmt.Fprint(p.out, "\033[1B"); err != nil {
				return err
			}
		}
	}

	// Move the cursor back to top of cleared lines
	if _, err := fmt.Fprintf(p.out, "\033[%dA", p.statusLineCount-1); err != nil {
		return err
	}

	p.statusVisible = false
	p.statusLineCount = 0

	return nil
}

func (p *ProgressRenderer) statusLines() []string {
	return []string{
		fmt.Sprintf("Task:		%s", p.totalMessage),
		fmt.Sprintf("Current:	%5.1f%% %s", p.currentPercent, p.currentMessage),
		fmt.Sprintf("Total:		%5.1f%%", p.totalPercent),
	}
}
