package lines

type ParsedLine interface {
	isParsedLine()
	Raw() string
}

type parsedLineBase struct {
	raw string
}

func (r parsedLineBase) Raw() string {
	return r.raw
}

type Message struct {
	parsedLineBase
	Code         string
	Flags        int
	Count        int
	Message      string
	Format       string
	Params       []string
	originalText string
}

func (Message) isParsedLine() {}

type ProgressTitle struct {
	parsedLineBase
	Code         string
	Id           int
	Name         string
	originalText string
}

func (ProgressTitle) isParsedLine() {}

type ProgressCurrent struct {
	parsedLineBase
	Code string
	Id   int
	Name string
}

func (ProgressCurrent) isParsedLine() {}

type ProgressValue struct {
	parsedLineBase
	Current int
	Total   int
	Max     int
}

func (ProgressValue) isParsedLine() {}

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

// Messages in the format TCOUT:count
type TitleCount struct {
	parsedLineBase
	// Title count
	Count int
}

func (TitleCount) isParsedLine() {}

// Messages in the format
type DiscInfo struct {
	parsedLineBase
	// Attribute id
	Id    int
	Code  int
	Value string
}

func (DiscInfo) isParsedLine() {}

type TitleInfoCode uint

const (
	TitleInfoCodeSize           TitleInfoCode = 11
	TitleInfoCodeSourceFileName TitleInfoCode = 16
	TitleInfoCodeOutputFileName TitleInfoCode = 27
	TitleInfoCodeSegmentsMap    TitleInfoCode = 26
)

type TitleInfo struct {
	parsedLineBase
	TitleId     uint
	AttributeId TitleInfoCode
	Code        uint
	Value       string
}

func (TitleInfo) isParsedLine() {}

// Messages in the format
type StreamInfo struct {
	parsedLineBase
	// Attribute id
	Id    int
	Code  int
	Value string
}

func (StreamInfo) isParsedLine() {}
