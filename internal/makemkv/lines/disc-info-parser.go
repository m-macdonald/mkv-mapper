package lines

import (
    "strings"
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

func (d *DiscInfoParser) Parse(raw string, payload string) (ParsedLine, error) {
    discInfo := DiscInfo {}
	discInfo.raw = raw

    params := strings.Split(payload, COMMA)

    if id, err := strconv.Atoi(params[0]); err == nil {
        discInfo.Id = id
    } else {
        //error handling
    }

    if code, err := strconv.Atoi(params[1]); err == nil {
        discInfo.Code = code
    } else {
        //error handling
    }

    discInfo.Value = params[2]

    return discInfo, nil
}


