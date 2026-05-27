package signature

import (
	"errors"
	"strings"
)

type SegmentSignature string

func NormalizeSegments(segmentString string) (SegmentSignature, error) {
	segmentString = strings.TrimSpace(segmentString)
	if len(segmentString) == 0 {
		return "", errors.New("segment string must not be empty")
	}
	segments := strings.Split(segmentString, ",")
	for i, s := range segments {
		segments[i] = strings.TrimSpace(s)
	}
	return SegmentSignature(strings.Join(segments, ",")), nil
}
