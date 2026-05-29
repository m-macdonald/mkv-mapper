package signature

import "testing"

func TestNormalizeSegments(t *testing.T) {
	tests := []struct {
		name          string
		segmentString string
		want          SegmentSignature
		wantErr       bool
	}{
		{
			name:          "creates SegmentSignature",
			segmentString: "05,08",
			want:          SegmentSignature("05,08"),
			wantErr:       false,
		},
		{
			name:          "handles range-style segment maps",
			segmentString: "1-6, 7-11",
			want:          SegmentSignature("1-6,7-11"),
		},
		{
			name:          "error when segment string is empty",
			segmentString: "  ",
			want:          SegmentSignature(""),
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeSegments(test.segmentString)
			if (err != nil) != test.wantErr {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}

			if got != test.want {
				t.Fatalf("got %s, want %s", got, test.want)
			}
		})
	}
}
