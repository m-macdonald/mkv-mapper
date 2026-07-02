package naming

import "testing"

func TestSanitizeSegment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "Mission Impossible",
			want:  "Mission Impossible",
		},
		{
			name:  "forward slash",
			input: "Fast and Furious 7/8",
			want:  "Fast and Furious 7-8",
		},
		{
			name:  "backslash",
			input: `Title\Subtitle`,
			want:  "Title-Subtitle",
		},
		{
			name:  "colon",
			input: "Mission: Impossible",
			want:  "Mission - Impossible",
		},
		{
			name:  "double quote",
			input: `The "Best" Movie`,
			want:  "The 'Best' Movie",
		},
		{
			name:  "pipe",
			input: "Good|Bad|Ugly",
			want:  "Good-Bad-Ugly",
		},
		{
			name:  "question mark",
			input: "What Lies Beneath?",
			want:  "What Lies Beneath",
		},
		{
			name:  "asterisk",
			input: "Movie*Title",
			want:  "MovieTitle",
		},
		{
			name:  "angle brackets",
			input: "<Title>",
			want:  "Title",
		},
		{
			name:  "multiple invalid characters",
			input: `Who: What/Where\When?`,
			want:  "Who - What-Where-When",
		},
		{
			name:  "trailing space",
			input: "Title ",
			want:  "Title",
		},
		{
			name:  "collapses whitespace left by removed characters",
			input: "Title * Subtitle?",
			want:  "Title Subtitle",
		},
		{
			name:  "empty string",
			input: "",
			want:  "_",
		},
		{
			name:  "only invalid characters",
			input: "???",
			want:  "_",
		},
		{
			name:  "leading and trailing whitespace preserved internally but trimmed at edges",
			input: "  Title  ",
			want:  "Title",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sanitizeSegment(test.input)
			if got != test.want {
				t.Errorf("want %q, got %q", test.want, got)
			}
		})
	}
}
