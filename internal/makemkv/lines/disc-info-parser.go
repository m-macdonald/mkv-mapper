package lines

import (
	"strconv"
)

// Messages in the format
type DiscInfo struct {
	parsedLineBase
	// Attribute id
	Id    int
	Code  int
	Value string
}

func (DiscInfo) isParsedLine() {}

type DiscInfoParser struct {}

func (d *DiscInfoParser) Parse(raw string, params []string) (ParsedLine, error) {
    discInfo := DiscInfo {}
	discInfo.raw = raw

    if id, err := strconv.Atoi(params[0]); err == nil {
        discInfo.Id = id
    } else {
		return nil, err
    }

    if code, err := strconv.Atoi(params[1]); err == nil {
        discInfo.Code = code
    } else {
		return nil, err
    }

    discInfo.Value = params[2]

    return discInfo, nil
}


