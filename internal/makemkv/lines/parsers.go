package lines

import (
    "fmt"
)

const (
    COMMA = ","
)

type LineProcessor struct {
    parsers []LineParser
}

func NewLineProcessor() *LineProcessor {
    return &LineProcessor {
        parsers: []LineParser {
            MessageParser {},
            ProgressCurrentParser {},
            ProgressTitleParser {},
            DriveScanParser {},
            ProgressTitleParser {},
            StreamInfoParser {},
            TitleCountParser {},
            TitleInfoParser {},
        },
    }
}

func (p *LineProcessor) ProcessLine(line string) (ParsedLine, error) {
    for _, parser := range p.parsers {
        if parser.CanParse(line) {
            return parser.Parse(line), nil
        }
    }

    return nil, fmt.Errorf("No parser available for line: %s", line)
}

type LineParser interface {
    Parse(lineText string) ParsedLine
    CanParse(lineText string) bool
}
