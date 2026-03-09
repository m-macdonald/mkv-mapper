package lines

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

type TitleInfoCode uint

const (
	TitleInfoCodeSize           TitleInfoCode = 11
	TitleInfoCodeSourceFileName TitleInfoCode = 16
	TitleInfoCodeOutputFileName TitleInfoCode = 27
	TitleInfoCodeSegmentsMap    TitleInfoCode = 26
)

type TitleInfo struct {
	parsedLineBase
	TitleId     uint
	AttributeId TitleInfoCode
	Code        uint
	Value       string
}

func (TitleInfo) isParsedLine() {}

type TitleInfoParser struct {}

func (t *TitleInfoParser) Parse(raw string, payload string) (ParsedLine, error) {
    titleInfo := TitleInfo {}
	titleInfo.raw = raw

	r := csv.NewReader(strings.NewReader(payload))
	r.LazyQuotes = true
	params, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to parse line payload %w", err)
	}

    if titleId, err := strconv.ParseUint(params[0], 10, 0); err == nil {
        titleInfo.TitleId = uint(titleId)
    } else {
		return nil, fmt.Errorf("Unable to parse uint for TitleId %w", err)
    }

	if attributeId, err := strconv.ParseUint(params[1], 10, 0); err == nil {
		titleInfo.AttributeId = TitleInfoCode(attributeId) 
	} else {
		return nil, fmt.Errorf("Unable to parse uint for AttributeId %w", err)
	}

    if code, err := strconv.ParseUint(params[2], 10, 0); err == nil {
        titleInfo.Code = uint(code)
    } else {
		return nil, fmt.Errorf("Unable to parse uint for Code %w", err)
    }

    titleInfo.Value = params[3]

    return titleInfo, nil
}
