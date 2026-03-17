package event

import (
	"fmt"
	"io"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"strings"
)

type Event interface {
	isEvent()
}

type MessageEvent struct {
	Message string
}

func (MessageEvent) isEvent() {}

type ProgressEvent struct {
	CurrentPercent float64
	TotalPercent   float64
}

func (ProgressEvent) isEvent() {}

func ParsedLineToEvent(line lines.ParsedLine) (Event, bool) {
	switch l := line.(type) {
	case lines.ProgressValue:
		return ProgressEvent{
			TotalPercent:   l.TotalPercent(),
			CurrentPercent: l.CurrentPercent(),
		}, true
	case lines.Message:
		return MessageEvent{
			Message: l.Message,
		}, true
	}

	return nil, false
}

type Renderer struct {
	out io.Writer
	interactive bool

	currentPercent float64
	totalPercent float64

	// currentMessage string
	totalMessage string

	statusVisible bool
}

func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{
		out: out,
		interactive: true,
	}
}

func (r *Renderer) HandleEvent(ev Event) error {
	switch e := ev.(type) {
	case MessageEvent:
		fmt.Printf("message")
		return r.handleMessage(e)
	case ProgressEvent:
		fmt.Printf("progress")
		return r.handleProgress(e)
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

	if _, err := fmt.Fprintf(r.out, ev.Message); err != nil {
		return err
	}

	if r.interactive && r.hasStatusContent() {
		return r.redrawStatus()
	}

	return nil
}

func (r *Renderer) handleProgress(ev ProgressEvent) error {
	if !r.interactive {
		return nil
	}

	r.currentPercent = ev.CurrentPercent
	r.totalPercent = ev.TotalPercent

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
	block := r.statusBlock()

	if r.statusVisible {
		if err := r.clearStatus(); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(r.out, block); err != nil {
		return err
	}

	r.statusVisible = true

	return nil
}

func (r *Renderer) clearStatus() error {
	if !r.statusVisible {
		return nil
	}

	// Move the cursor up 3 lines
	if _, err := fmt.Fprint(r.out, "\033[3A"); err != nil {
		return err
	}

	// Clear terminal output moving down a total of 4 lines
	for i := range [3]int{} {
		if _, err := fmt.Fprintf(r.out, "\r\033[K"); err != nil {
			return err
		}

		if i < 2 {
			if _, err := fmt.Fprintf(r.out, "\033[1B"); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintf(r.out, "\033[2A"); err != nil {
		return err
	}

	r.statusVisible = false

	return nil
}

func (r *Renderer) statusBlock() string {
	lines := []string{
		fmt.Sprintf("Task:		%s", r.totalMessage),
		fmt.Sprintf("Current:	%5.1f%%", r.currentPercent),
		fmt.Sprintf("Total:		%5.1f%%", r.totalPercent),
	}

	return strings.Join(lines, "\n") + "\n"
}


