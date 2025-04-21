package lines

type ParsedLine interface {}

type Message struct {
    Code        string
    Flags       int
    Count       int
    Message     string
    Format      string
    Params      []string
    OriginalMessage string
}

type ProgressTitle struct {
    Code        string
    Id          int
    Name        string
}

type ProgressCurrent struct {
    Code        string
    Id          int
    Name        string
}

type ProgressValue struct {
    Current     int
    Total       int
    Max         int
}

type DriveScan struct {
    Index       int
    Visible     bool
    Enabled     bool
    Flags       int
    DriveName   string
    DiscName    string
}

// Messages in the format TCOUT:count
type TitleCount struct {
    // Title count
    Count       int
}

// Messages in the format 
type DiscInformation struct {
    // Attribute id
    Id          int
    Code        int
    Value       string
}

type TitleInformation struct {
    // Attribute id
    Id          int
    Code        int
    Value       string
}

// Messages in the format 
type StreamInformation struct {
    // Attribute id
    Id          int
    Code        int
    Value       string
}
