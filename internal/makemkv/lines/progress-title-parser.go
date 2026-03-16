package lines

import (
    "strings"
    "strconv"
)

type ProgressTitle struct {
	parsedLineBase
	Code         string
	Id           int
	Name         string
}

func (ProgressTitle) isParsedLine() {}

type ProgressTitleParser struct {}

func (p *ProgressTitleParser) Parse(raw string, payload string) (ParsedLine, error) {
    progressTitle := ProgressTitle {}
	progressTitle.raw = raw

    params := strings.Split(payload, COMMA)

    progressTitle.Code = params[0]
    if id, err := strconv.Atoi(params[1]); err == nil {
        progressTitle.Id = id
    } else {
		return nil, err
    }
    progressTitle.Name = params[2]

    return progressTitle, nil
}
