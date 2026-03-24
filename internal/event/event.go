package event

import (
	"fmt"
	"io"
	"strings"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

type Event interface {
	isEvent()
}

type MessageEvent struct {
	Message string
}

func (MessageEvent) isEvent() {}

type ProgressPercentEvent struct {
	CurrentPercent float64
	TotalPercent   float64
}

func (ProgressPercentEvent) isEvent() {}

type ProgressCurrentEvent struct {
	Message string
}

func (ProgressCurrentEvent) isEvent() {}

type ProgressTotalEvent struct {
	Message string
}

func (ProgressTotalEvent) isEvent() {}

func ParsedLineToEvent(line lines.ParsedLine) (Event, bool) {
	switch l := line.(type) {
	case lines.ProgressValue:
		return ProgressPercentEvent{
			TotalPercent:   l.TotalPercent(),
			CurrentPercent: l.CurrentPercent(),
		}, true
	case lines.ProgressCurrent:
		return ProgressCurrentEvent{
			Message: l.Name,
		}, true
	case lines.ProgressTitle:
		return ProgressTotalEvent{
			Message: l.Name,
		}, true
	case lines.Message:
		return MessageEvent{
			Message: l.Message,
		}, true
	}

	return nil, false
}

type Renderer struct {
	out         io.Writer
	interactive bool

	currentPercent float64
	totalPercent   float64

	currentMessage string
	totalMessage   string

	statusVisible   bool
	statusLineCount int
}

func NewRenderer(out io.Writer, interactive bool) *Renderer {
	return &Renderer{
		out:         out,
		interactive: interactive,
	}
}

func (r *Renderer) HandleEvent(ev Event) error {
	switch e := ev.(type) {
	case MessageEvent:
		return r.handleMessage(e)
	case ProgressPercentEvent:
		return r.handleProgressPercent(e)
	case ProgressCurrentEvent:
		return r.handleProgressCurrent(e)
	case ProgressTotalEvent:
		return r.handleProgressTotal(e)
	default:
		return nil
	}
}

func (r *Renderer) handleMessage(ev MessageEvent) error {
	if r.interactive && r.statusVisible {
		if err := r.clearStatus(); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(r.out, ev.Message+"\n"); err != nil {
		return err
	}

	if r.interactive && r.hasStatusContent() {
		return r.redrawStatus()
	}

	return nil
}

func (r *Renderer) handleProgressPercent(ev ProgressPercentEvent) error {
	if !r.interactive {
		return nil
	}

	r.currentPercent = ev.CurrentPercent
	r.totalPercent = ev.TotalPercent

	return r.redrawStatus()
}

func (r *Renderer) handleProgressCurrent(ev ProgressCurrentEvent) error {
	if !r.interactive {
		return nil
	}

	r.currentMessage = ev.Message

	return r.redrawStatus()
}

func (r *Renderer) handleProgressTotal(ev ProgressTotalEvent) error {
	if !r.interactive {
		return nil
	}

	r.totalMessage = ev.Message

	return r.redrawStatus()
}

func (r *Renderer) Close() error {
	if r.interactive && r.statusVisible {
		return r.clearStatus()
	}

	return nil
}

func (r *Renderer) hasStatusContent() bool {
	return r.currentPercent > 0 || r.totalPercent > 0
}

func (r *Renderer) redrawStatus() error {
	statusLines := r.statusLines()

	if r.statusVisible {
		if err := r.clearStatus(); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(r.out, strings.Join(statusLines, "\n")); err != nil {
		return err
	}

	r.statusVisible = true
	r.statusLineCount = len(statusLines)

	return nil
}

func (r *Renderer) clearStatus() error {
	if !r.statusVisible {
		return nil
	}

	// Move the cursor to the top of the status block
	if _, err := fmt.Fprintf(r.out, "\033[%dA", r.statusLineCount); err != nil {
		return err
	}

	for i := range r.statusLineCount {
		// Clear the cursor's current line
		if _, err := fmt.Fprintf(r.out, "\r\033[K"); err != nil {
			return err
		}

		if i < r.statusLineCount-1 {
			// Move cursor down a line
			if _, err := fmt.Fprint(r.out, "\033[1B"); err != nil {
				return err
			}
		}
	}

	// Move the cursor back to top of cleared lines
	if _, err := fmt.Fprintf(r.out, "\033[%dA", r.statusLineCount-1); err != nil {
		return err
	}

	r.statusVisible = false
	r.statusLineCount = 0

	return nil
}

func (r *Renderer) statusLines() []string {
	return []string{
		fmt.Sprintf("Task:		%s", r.currentMessage),
		fmt.Sprintf("Current:	%5.1f%%", r.currentPercent),
		fmt.Sprintf("Total:		%5.1f%%", r.totalPercent),
	}
}
