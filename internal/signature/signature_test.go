package signature

import "testing"

func TestNormalizeSegments(t *testing.T) {
	tests := []struct {
		name          string
		segmentString string
		want          SegmentSignature
	}{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeSegments(test.segmentString)
			if err != nil {
			}

			if got != test.want {
				t.Fatalf("got %s, want %s", got, test.want)
			}
		})
	}
}
