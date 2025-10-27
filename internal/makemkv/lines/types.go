package lines

type ParsedLine interface {
    // OriginalText() string
}

type Message struct {
    Code            string
    Flags           int
    Count           int
    Message         string
    Format          string
    Params          []string
    originalText    string
}

func (m Message) OriginalText() string {
    return m.originalText
}

type ProgressTitle struct {
    Code            string
    Id              int
    Name            string
    originalText    string
}

func (p ProgressTitle) OriginalText() string {
    return p.originalText
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
type DiscInfo struct {
    // Attribute id
    Id          int
    Code        int
    Value       string
}

type TitleInfo struct {
    // Attribute id
    Id          int
    Code        int
    Value       string
}

// Messages in the format 
type StreamInfo struct {
    // Attribute id
    Id          int
    Code        int
    Value       string
}
