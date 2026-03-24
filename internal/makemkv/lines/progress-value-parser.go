package lines

import (
	"strconv"
)

type ProgressValue struct {
	parsedLineBase
	Current      uint64
	Total        uint64
	Max          uint64
}

func (ProgressValue) isParsedLine() {}

func (p *ProgressValue) CurrentPercent() float64 {
	if p.Max == 0 {
		return 0
	}

	return float64(p.Current) / float64(p.Max) * 100
}

func (p *ProgressValue) TotalPercent() float64 {
	if p.Max == 0 {
		return 0
	}

	return float64(p.Total) / float64(p.Max) * 100
}

type ProgressValueParser struct{}

func (p *ProgressValueParser) Parse(raw string, params []string) (ParsedLine, error) {
	progressValue := ProgressValue{}
	progressValue.raw = raw

	if current, err := strconv.ParseUint(params[0], 10, 64); err == nil {
		progressValue.Current = current
	} else {
		return nil, err
	}

	if total, err := strconv.ParseUint(params[1], 10, 64); err == nil {
		progressValue.Total = total	
	} else {
		return nil, err
	}
	
	// Avoiding collision with "max" built-in
	if maxValue, err :=  strconv.ParseUint(params[2], 10, 64); err == nil {
		progressValue.Max = maxValue
	} else {
		return nil, err
	}

	return progressValue, nil
}
