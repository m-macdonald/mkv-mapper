package naming

import (
	"testing"

	"m-macdonald/mkv-mapper/internal/discdb"
)

func TestTemplateFromItemType(t *testing.T) {
	tests := []struct {
		name     string
		itemType discdb.ItemType
		want     templateType
	}{
		{
			name: "ItemTypeMovie -> templateTypeMovie",
			itemType: discdb.ItemTypeMovie,
			want: templateTypeMovie,
		},
		{
			name: "ItemTypeEpisode -> templateTypeEpisode",
			itemType: discdb.ItemTypeEpisode,
			want: templateTypeEpisode,
		},
		{
			name: "ItemTypeDeletedScene -> templateTypeExtra",
			itemType: discdb.ItemTypeExtra,
			want: templateTypeExtra,
		},
		{
			name: "ItemTypeExtra -> templateTypeExtra",
			itemType: discdb.ItemTypeTrailer,
			want: templateTypeExtra,
		},
		{
			name: "ItemTypeTrailer -> templateTypeExtra",
			itemType: discdb.ItemTypeTrailer,
			want: templateTypeExtra,
		},
		{
			name: "default -> templateTypeUnknown",
			itemType: discdb.ItemType("unknown"),
			want: templateTypeUnknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := templateTypeFromItemType(test.itemType)

			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}
