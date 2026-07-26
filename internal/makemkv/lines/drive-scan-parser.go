package lines

import (
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
	Device    string
}

func (DriveScan) isParsedLine() {}

type DriveScanParser struct{}

func (d *DriveScanParser) Parse(raw string, params []string) (ParsedLine, error) {
	driveScan := DriveScan{}
	driveScan.raw = raw

	if index, err := strconv.Atoi(params[0]); err == nil {
		driveScan.Index = index
	} else {
		return nil, err
	}

	if visible, err := strconv.Atoi(params[1]); err == nil {
		if visible == 1 {
			driveScan.Visible = true
		} else {
			driveScan.Visible = false
		}
	} else {
		return nil, err
	}

	if enabled, err := strconv.Atoi(params[2]); err == nil {
		if enabled == 1 {
			driveScan.Enabled = true
		} else {
			driveScan.Enabled = false
		}
	} else {
		return nil, err
	}

	if flags, err := strconv.Atoi(params[3]); err == nil {
		driveScan.Flags = flags
	} else {
		return nil, err
	}

	driveScan.DriveName = params[4]
	driveScan.DiscName = params[5]
	driveScan.Device = params[6]

	return driveScan, nil
}
