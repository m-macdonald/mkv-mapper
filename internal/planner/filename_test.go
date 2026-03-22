package planner

import "testing"

func TestEnsureUniqueFilename(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		ext     string
		titleId int
		used    map[string]struct{}
		want    string
	}{
		{
			name:    "no collisions",
			base:    "movie",
			ext:     ".mkv",
			titleId: 1,
			used:    map[string]struct{}{},
			want:    "movie.mkv",
		},
		{
			name:    "collision",
			base:    "movie",
			ext:     ".mkv",
			titleId: 1,
			used: map[string]struct{}{
				"movie.mkv": {},
			},
			want: "movie_t1.mkv",
		},
		{
			name: "multiple collisions",
			base: "movie",
			ext:  ".mkv",
			titleId: 1,
			used: map[string]struct{}{
				"movie.mkv": {},
				"movie_t1.mkv": {},
			},
			want: "movie_t1_1.mkv",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _, err := ensureUniqueFilename(test.base, test.ext, test.titleId, test.used)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != test.want {
				t.Fatalf("got %s, want %s", got, test.want)
			}
		})
	}
}
