package lines

import (
    "strings"
    "strconv"
)

type MessageParser struct {}

func (m MessageParser) Parse(raw string, payload string) (ParsedLine, error) {
    message := Message {}
	message.raw = raw

    params := strings.Split(payload, COMMA)

    message.Code = params[0]
    if flags, err := strconv.Atoi(params[1]); err == nil {
        message.Flags = flags
    } else {
        // TODO : Handle errors
    }
    if count, err := strconv.Atoi(params[2]); err == nil {
        message.Count = count
    } else {

    }
    message.Message = params[3]
    message.Format = params[4]
	copy(message.Params, params[5:])

    return message, nil
}
