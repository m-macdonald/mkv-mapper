package engine

import (
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParsedLineToEvent(t *testing.T) {
	tests := []struct {
		name      string
		line      lines.ParsedLine
		wantEvent event.Event
		wantOk    bool
	}{
		{
			name: "progress value",
			line: lines.ProgressValue{
				Current: 23,
				Total: 60,
				Max: 100,
			},
			wantEvent: event.ProgressPercentEvent{
				TotalPercent: 60,
				CurrentPercent: 23,
			},
			wantOk: true,
		},
		{
			name:      "progress current",
			line:      lines.ProgressCurrent{Name: "Ripping title 1"},
			wantEvent: event.ProgressCurrentEvent{Message: "Ripping title 1"},
			wantOk:    true,
		},
		{
			name:      "progress title",
			line:      lines.ProgressTitle{Name: "Overall progress"},
			wantEvent: event.ProgressTotalEvent{Message: "Overall progress"},
			wantOk:    true,
		},
		{
			name:      "message",
			line:      lines.Message{Message: "Some message"},
			wantEvent: event.MessageEvent{Message: "Some message"},
			wantOk:    true,
		},
		{
			name:   "unrecognized line",
			line:   lines.StreamInfo{},
			wantOk: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parsedLineToEvent(test.line)
			if ok != test.wantOk {
				t.Fatalf("expected ok %v, got %v", test.wantOk, ok)
			}
			if !test.wantOk {
				return
			}
			if diff := cmp.Diff(test.wantEvent, got); diff != "" {
				t.Errorf("event mismatch: %s", diff)
			}
		})
	}
}
