package lines

import (
    "strings"
    "strconv"
)

type StreamInfoParser struct {}

func (s StreamInfoParser) prefix() string {
    return "SINFO:"
}

func (s StreamInfoParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, s.prefix())
}

func (s StreamInfoParser) Parse(lineText string) ParsedLine {
    streamInfo := StreamInfo {}
    params := strings.Split(lineText, COMMA)

    if id, err := strconv.Atoi(params[0]); err == nil {
        streamInfo.Id = id
    } else {
        //error handling
    }

    if code, err := strconv.Atoi(params[1]); err == nil {
        streamInfo.Code = code
    } else {
        //error handling
    }

    streamInfo.Value = params[2]

    return streamInfo
}
