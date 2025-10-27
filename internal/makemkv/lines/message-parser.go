package lines

import (
    "strings"
    "strconv"
)

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
