package lines

import (
    "strings"
    "strconv"
)

type DiscInfoParser struct {}

func (d DiscInfoParser) prefix() string {
    return "CINFO:"
}

func (d DiscInfoParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, d.prefix())
}

func (t DiscInfoParser) Parse(lineText string) ParsedLine {
    discInfo := DiscInfo {}
    params := strings.Split(lineText, COMMA)

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

    return discInfo
}


