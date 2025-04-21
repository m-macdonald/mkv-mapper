package lines

import (
    "fmt"
    "strconv"
    "strings"
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
            ProgressTotalParser {},
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

type MessageParser struct {}

func (m MessageParser) prefix() string {
    return "MSG:"
}

func (m MessageParser) Parse(lineText string) ParsedLine {
    message := Message {}
    params := strings.Split(lineText, COMMA)

    message.Code = params[1]
    if flags, err := strconv.Atoi(params[2]); err == nil {
        message.Flags = flags
    } else {
        // TODO : Handle errors
    }
    if count, err := strconv.Atoi(params[3]); err == nil {
        message.Count = count
    } else {

    }
    message.Message = params[4]
    message.Format = params[5]
    for i, param := range params[6:] {
        message.Params[i] = param
    }

    return &message
}

func (m MessageParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, m.prefix())
}

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
