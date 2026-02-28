package lines

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

type TitleInfoParser struct {}

func (t TitleInfoParser) Parse(raw string, payload string) (ParsedLine, error) {
    titleInfo := TitleInfo {}
	titleInfo.raw = raw

	r := csv.NewReader(strings.NewReader(payload))
	r.LazyQuotes = true
	params, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to parse line payload %w", err)
	}
    // params := strings.Split(payload, COMMA)

    if titleId, err := strconv.ParseUint(params[0], 10, 0); err == nil {
        titleInfo.TitleId = uint(titleId)
    } else {
        //error handling
    }

	if attributeId, err := strconv.ParseUint(params[1], 10, 0); err == nil {
		titleInfo.AttributeId = TitleInfoCode(attributeId) 
	} else {
		// error handling
	}

    if code, err := strconv.ParseUint(params[2], 10, 0); err == nil {
        titleInfo.Code = uint(code)
    } else {
        //error handling
    }

    titleInfo.Value = params[3]

    return titleInfo, nil
}
