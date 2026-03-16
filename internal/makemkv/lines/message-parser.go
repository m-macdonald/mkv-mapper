package lines

import (
	"strconv"
	"strings"
)

type Message struct {
	parsedLineBase
	Code                 string
	Flags                int
	Count                int
	Message              string
	ParameterizedMessage string
	Params               []string
}

func (Message) isParsedLine() {}

type MessageParser struct{}

func (m *MessageParser) Parse(raw string, payload string) (ParsedLine, error) {
	message := Message{}
	message.raw = raw

	params := strings.Split(payload, COMMA)

	message.Code = params[0]
	if flags, err := strconv.Atoi(params[1]); err == nil {
		message.Flags = flags
	} else {
		return nil, err
	}
	if count, err := strconv.Atoi(params[2]); err == nil {
		message.Count = count
	} else {
		return nil, err
	}
	message.Message = params[3]
	message.ParameterizedMessage = params[4]
	copy(message.Params, params[5:])

	return message, nil
}
