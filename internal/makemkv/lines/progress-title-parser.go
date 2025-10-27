package lines

import (
    "strings"
    "strconv"
)

type ProgressTitleParser struct {}

func (p ProgressTitleParser) prefix() string {
    return "PRGT:"
}

func (p ProgressTitleParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, p.prefix())
}

func (p ProgressTitleParser) Parse(lineText string) ParsedLine {
    progressTitle := ProgressTitle {}
    params := strings.Split(lineText, COMMA)

    progressTitle.Code = params[1]
    if id, err := strconv.Atoi(params[2]); err == nil {
        progressTitle.Id = id
    } else {

    }
    progressTitle.Name = params[3]

    return progressTitle
}
