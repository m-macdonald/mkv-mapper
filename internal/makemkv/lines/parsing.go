package lines

import (
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
	Parse(raw string, payload string) (ParsedLine, error)
}

func NewLineProcessor() *LineProcessor {
	return &LineProcessor{
		parsers: map[string]LineParser{
			"MSG":    &MessageParser{},
			"PRGC":   &ProgressCurrentParser{},
			"PRGT":   &ProgressTitleParser{},
			"DRV":    &DriveScanParser{},
			"SINFO":  &StreamInfoParser{},
			"TCOUNT": &TitleCountParser{},
			"TINFO":  &TitleInfoParser{},
			"CINFO":  &DiscInfoParser{},
		},
	}
}

func (p *LineProcessor) ProcessLine(line string) (ParsedLine, error) {
	prefix, payload, ok := strings.Cut(line, ":")
	if !ok {
		return nil, fmt.Errorf("invalid line: %s", line)
	}

	parser, exists := p.parsers[prefix]
	if !exists {
		return nil, fmt.Errorf("no parser for prefix: %s", prefix)
	}

	return parser.Parse(line, payload)
}
