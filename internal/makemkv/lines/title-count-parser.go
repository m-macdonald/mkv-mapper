package lines

import (
    "strconv"
)

// Messages in the format TCOUT:count
type TitleCount struct {
	parsedLineBase
	// Title count
	Count int
}

func (TitleCount) isParsedLine() {}

type TitleCountParser struct {}

func (t *TitleCountParser) Parse(raw string, params []string) (ParsedLine, error) {
    titleCount := TitleCount {}
	titleCount.raw = raw

    if count, err := strconv.Atoi(params[0]); err == nil {
        titleCount.Count = count
    } else {
		return nil, err
    }

    return titleCount, nil
}


