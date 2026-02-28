package lines

import (
    "strings"
    "strconv"
)

type TitleCountParser struct {}

func (t TitleCountParser) Parse(raw string, payload string) (ParsedLine, error) {
    titleCount := TitleCount {}
    params := strings.Split(payload, COMMA)

    if count, err := strconv.Atoi(params[0]); err == nil {
        titleCount.Count = count
    } else {
        //error handling
    }

    return titleCount, nil
}


