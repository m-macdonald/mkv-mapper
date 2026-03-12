package signature

import (
	"fmt"
	"strconv"
	"strings"
)

type SegmentSignature string

func NormalizeSegments(segmentString string) (SegmentSignature, error) {
	segments := strings.Split(segmentString, ",")
	parts := make([]string, 0, len(segments))

	for _, segment := range segments {
		segment = strings.TrimSpace(segment)

		i, err := strconv.Atoi(segment)
		if err != nil {
			return "", err
		}
		a := fmt.Sprintf("%05d", i)

		parts = append(parts, a)
	}

	return SegmentSignature(strings.Join(parts, "|")), nil
}
