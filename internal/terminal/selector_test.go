package terminal

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
)

func TestSelectEmptyPlan(t *testing.T) {
	selector := NewSelector()
	_, err := selector.Select(model.Plan{})
	if err == nil {
		t.Error("Select() with empty plan should return error")
	}
	if err.Error() != "no titles available to select from" {
		t.Errorf("Select() error = %q, want %q", err.Error(), "no titles available to select from")
	}
}

func TestFormatTitleLabel(t *testing.T) {
	tests := []struct {
		name      string
		title     model.TitlePlan
		wantLabel string
	}{
		{
			name: "matched title",
			title: model.TitlePlan{
				TitleId:       1,
				FinalName:     "The Movie.mkv",
				EstimatedSize: 5368709120, // 5 GB
				Duration:      "1:04:24",
				IsMatched:     true,
			},
			wantLabel: "1: The Movie.mkv (1:04:24 / 5.0 GB) [matched]",
		},
		{
			name: "unmatched title",
			title: model.TitlePlan{
				TitleId:       2,
				FinalName:     "Unknown Title.mkv",
				EstimatedSize: 1073741824, // 1 GB
				Duration:      "2:56:31",
				IsMatched:     false,
			},
			wantLabel: "2: Unknown Title.mkv (2:56:31 / 1.0 GB)",
		},
		{
			name: "large size",
			title: model.TitlePlan{
				TitleId:       3,
				FinalName:     "Feature.mkv",
				EstimatedSize: 88236484608, // 82.1 GB
				Duration:      "3:25:35",
				IsMatched:     true,
			},
			wantLabel: "3: Feature.mkv (3:25:35 / 82.2 GB) [matched]",
		},
		{
			name: "small size",
			title: model.TitlePlan{
				TitleId:       4,
				FinalName:     "Trailer.mkv",
				EstimatedSize: 104857600, // 100 MB
				Duration:      "4:23:54",
				IsMatched:     false,
			},
			wantLabel: "4: Trailer.mkv (4:23:54 / 100.0 MB)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTitleLabel(tt.title)
			if got != tt.wantLabel {
				t.Errorf("formatTitleLabel() = %q, want %q", got, tt.wantLabel)
			}
		})
	}
}

func TestTitleOptions(t *testing.T) {
	titles := []model.TitlePlan{
		{
			TitleId:       1,
			FinalName:     "Movie.mkv",
			EstimatedSize: 5368709120,
			Duration:      "1:02:19",
			IsMatched:     true,
		},
		{
			TitleId:       2,
			FinalName:     "Extra.mkv",
			EstimatedSize: 1073741824,
			Duration:      "2:05:06",
			IsMatched:     false,
		},
		{
			TitleId:       3,
			FinalName:     "Trailer.mkv",
			EstimatedSize: 536870912,
			Duration:      "3:34:59",
			IsMatched:     true,
		},
	}

	opts := newTitleOptions(titles)

	if len(opts) != len(titles) {
		t.Errorf("newTitleOptions() returned %d options, want %d", len(opts), len(titles))
	}

	labels := opts.labels()
	if len(labels) != len(opts) {
		t.Errorf("labels() returned %d, want %d", len(labels), len(opts))
	}

	for i, opt := range opts {
		expectedLabel := formatTitleLabel(titles[i])
		if opt.label != expectedLabel {
			t.Errorf("option %d label = %q, want %q", i, opt.label, expectedLabel)
		}
		if opt.titleId != titles[i].TitleId {
			t.Errorf("option %d id = %d, want %d", i, opt.id, titles[i].TitleId)
		}
	}
}

func TestTitleOptionsIdFor(t *testing.T) {
	titles := []model.TitlePlan{
		{TitleId: 10, FinalName: "First.mkv", EstimatedSize: 1000, IsMatched: true},
		{TitleId: 25, FinalName: "Second.mkv", EstimatedSize: 2000, IsMatched: false},
		{TitleId: 42, FinalName: "Third.mkv", EstimatedSize: 3000, IsMatched: true},
	}

	opts := newTitleOptions(titles)

	tests := []struct {
		name      string
		label     string
		wantId    lines.TitleId
		wantFound bool
	}{
		{
			name:      "exact match first",
			label:     formatTitleLabel(titles[0]),
			wantId:    10,
			wantFound: true,
		},
		{
			name:      "exact match middle",
			label:     formatTitleLabel(titles[1]),
			wantId:    25,
			wantFound: true,
		},
		{
			name:      "exact match last",
			label:     formatTitleLabel(titles[2]),
			wantId:    42,
			wantFound: true,
		},
		{
			name:      "nonexistent label",
			label:     "999: Fake Title (999.9 GB)",
			wantId:    0,
			wantFound: false,
		},
		{
			name:      "partial label mismatch",
			label:     "10: Wrong Name (1.0 GB) [matched]",
			wantId:    0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := opts.idFor(tt.label)
			if found != tt.wantFound {
				t.Errorf("idFor() found = %v, want %v", found, tt.wantFound)
			}
			if found && got != tt.wantId {
				t.Errorf("idFor() id = %d, want %d", got, tt.wantId)
			}
		})
	}
}

func TestTitleOptionsRoundtrip(t *testing.T) {
	original := model.TitlePlan{
		TitleId:       99,
		FinalName:     "Roundtrip Test.mkv",
		EstimatedSize: 7516192768,
		IsMatched:     true,
	}

	opts := newTitleOptions([]model.TitlePlan{original})
	label := formatTitleLabel(original)

	recoveredId, found := opts.idFor(label)
	if !found {
		t.Fatal("idFor() failed to find generated label")
	}
	if recoveredId != original.TitleId {
		t.Errorf("roundtrip: original TitleId %d, recovered %d", original.TitleId, recoveredId)
	}
}
