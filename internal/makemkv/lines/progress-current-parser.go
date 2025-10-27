package lines

import (
    "strings"
    "strconv"
)

type ProgressCurrentParser struct {}

func (p ProgressCurrentParser) prefix() string {
    return "PRGC:"
}

func (p ProgressCurrentParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, p.prefix())
}

func (p ProgressCurrentParser) Parse(lineText string) ParsedLine {
    progressCurrent := ProgressCurrent {}
    params := strings.Split(lineText, COMMA)

    progressCurrent.Code = params[1]
    if id, err := strconv.Atoi(params[2]); err == nil {
        progressCurrent.Id = id
    } else {

    }
    progressCurrent.Name = params[3]

    return progressCurrent
}
