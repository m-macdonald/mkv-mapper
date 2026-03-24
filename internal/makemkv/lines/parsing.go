package lines

import (
	"encoding/csv"
	"fmt"
	"strings"
)

const (
	COMMA = ","
)

type ParsedLine interface {
	isParsedLine()
	Raw() string
}

type parsedLineBase struct {
	raw string
}

func (r parsedLineBase) Raw() string {
	return r.raw
}

type LineProcessor struct {
	parsers map[string]LineParser
}

type LineParser interface {
	Parse(raw string, params []string) (ParsedLine, error)
}

func NewLineProcessor() *LineProcessor {
	return &LineProcessor{
		parsers: map[string]LineParser{
			"CINFO":  &DiscInfoParser{},
			"DRV":    &DriveScanParser{},
			"MSG":    &MessageParser{},
			"PRGC":   &ProgressCurrentParser{},
			"PRGT":   &ProgressTitleParser{},
			"PRGV":   &ProgressValueParser{},
			"SINFO":  &StreamInfoParser{},
			"TCOUNT": &TitleCountParser{},
			"TINFO":  &TitleInfoParser{},
		},
	}
}

func (p *LineProcessor) ProcessLine(line string) (ParsedLine, error) {
	prefix, payload, ok := strings.Cut(line, ":")
	if !ok {
		return nil, fmt.Errorf("invalid line: %s", line)
	}

	params, err := splitCsv(payload)
	if err != nil {
		return nil, err
	}

	parser, exists := p.parsers[prefix]
	if !exists {
		return nil, fmt.Errorf("no parser for prefix: %s", prefix)
	}

	return parser.Parse(line, params)
}

func splitCsv(payload string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(payload))
	r.LazyQuotes = true
	params, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("split csv: %w", err)
	}

	return params, nil
}
