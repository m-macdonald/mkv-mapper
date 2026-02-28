package lines

import (
	"fmt"
	"strings"
)

const (
	COMMA = ","
)

type LineProcessor struct {
	parsers map[string]LineParser
}

func NewLineProcessor() *LineProcessor {
	return &LineProcessor{
		parsers: map[string]LineParser{
			"MSG":    MessageParser{},
			"PRGC":   ProgressCurrentParser{},
			"PRGT":   ProgressTitleParser{},
			"DRV":    DriveScanParser{},
			"SINFO":  StreamInfoParser{},
			"TCOUNT": TitleCountParser{},
			"TINFO":  TitleInfoParser{},
			"CINFO":  DiscInfoParser{},
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

type LineParser interface {
	Parse(raw string, payload string) (ParsedLine, error)
}
