package lines

import (
    "strconv"
)

// Messages in the format
type StreamInfo struct {
	parsedLineBase
	// Attribute id
	Id    int
	Code  int
	Value string
}

func (StreamInfo) isParsedLine() {}

type StreamInfoParser struct {}

func (s *StreamInfoParser) Parse(raw string, params []string) (ParsedLine, error) {
    streamInfo := StreamInfo {}
	streamInfo.raw = raw

    if id, err := strconv.Atoi(params[0]); err == nil {
        streamInfo.Id = id
    } else {
		return nil, err
    }

    if code, err := strconv.Atoi(params[1]); err == nil {
        streamInfo.Code = code
    } else {
		return nil, err
    }

    streamInfo.Value = params[2]

    return streamInfo, nil
}
