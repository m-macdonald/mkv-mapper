package lines

import (
    "strconv"
)

type ProgressCurrent struct {
	parsedLineBase
	Code string
	Id   int
	Name string
}

func (ProgressCurrent) isParsedLine() {}

type ProgressCurrentParser struct {}

func (p *ProgressCurrentParser) Parse(raw string, params []string) (ParsedLine, error) {
    progressCurrent := ProgressCurrent {}
	progressCurrent.raw = raw

    progressCurrent.Code = params[0]
    if id, err := strconv.Atoi(params[1]); err == nil {
        progressCurrent.Id = id
    } else {

    }
    progressCurrent.Name = params[2]

    return progressCurrent, nil
}
