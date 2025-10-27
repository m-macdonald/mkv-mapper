package lines

import (
    "strings"
    "strconv"
)

type TitleCountParser struct {}

func (t TitleCountParser) prefix() string {
    return "TCOUT:"
}

func (t TitleCountParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, t.prefix())
}

func (t TitleCountParser) Parse(lineText string) ParsedLine {
    titleCount := TitleCount {}
    params := strings.Split(lineText, COMMA)

    if count, err := strconv.Atoi(params[1]); err == nil {
        titleCount.Count = count
    } else {
        //error handling
    }

    return titleCount
}


