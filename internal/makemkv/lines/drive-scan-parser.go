package lines

import (
    "strings"
    "strconv"
)

type DriveScan struct {
	parsedLineBase
	Index     int
	Visible   bool
	Enabled   bool
	Flags     int
	DriveName string
	DiscName  string
}

func (DriveScan) isParsedLine() {}

type DriveScanParser struct {}

func (d *DriveScanParser) Parse(raw string, payload string) (ParsedLine, error) {
    driveScan := DriveScan {}
	driveScan.raw = raw

    params := strings.Split(payload, COMMA)

    if index, err := strconv.Atoi(params[0]); err == nil {
        driveScan.Index = index
    } else {
        // error handling
    }

    if visible, err := strconv.Atoi(params[1]); err == nil {
        if (visible == 1) {
            driveScan.Visible = true
        } else {
            driveScan.Visible = false
        }
    } else {
        // error handling
    }

    if enabled, err := strconv.Atoi(params[2]); err == nil {
        if enabled == 1 {
            driveScan.Enabled = true
        } else {
            driveScan.Enabled = false
        }
    } else {
        // error handling
    }

    return driveScan, nil
}
