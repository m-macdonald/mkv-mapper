package lines

import (
    "strings"
    "strconv"
)

type DriveScanParser struct {}

func (d DriveScanParser) prefix() string {
    return "DRV:"
}

func (d DriveScanParser) Parse(lineText string) ParsedLine {
    driveScan := DriveScan {}
    params := strings.Split(lineText, COMMA)

    if index, err := strconv.Atoi(params[1]); err == nil {
        driveScan.Index = index
    } else {
        // error handling
    }

    if visible, err := strconv.Atoi(params[2]); err == nil {
        if (visible == 1) {
            driveScan.Visible = true
        } else {
            driveScan.Visible = false
        }
    } else {
        // error handling
    }

    if enabled, err := strconv.Atoi(params[3]); err == nil {
        if enabled == 1 {
            driveScan.Enabled = true
        } else {
            driveScan.Enabled = false
        }
    } else {
        // error handling
    }

    



    return driveScan
}

func (d DriveScanParser) CanParse(lineText string) bool {
    return strings.HasPrefix(lineText, d.prefix())
}
