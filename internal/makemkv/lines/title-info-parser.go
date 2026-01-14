package lines

import (
    "strings"
    "strconv"
)

type TitleInfoParser struct {}

func (t TitleInfoParser) prefix() string {
    return "TINFO:"
}

func (d TitleInfoParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, d.prefix())
}

func (t TitleInfoParser) Parse(lineText string) ParsedLine {
    titleInfo := TitleInfo {}
    params := strings.Split(lineText, COMMA)

    if id, err := strconv.Atoi(params[0]); err == nil {
        titleInfo.Id = id
    } else {
        //error handling
    }

 //    if code, err := strconv.Atoi(params[1]); err == nil {
 //        titleInfo.Code = code
 //    } else {
 //        //error handling
 //    }
	//
	// if code == TitleInfoCodeEnum.Size {}

    titleInfo.Value = params[2]

    return titleInfo
}

