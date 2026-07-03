package planner

type PlanBase struct {
	Disc      Disc
	DiscRoot  string
	MediaInfo MediaInfo
	OutputDir string
	Titles    []TitlePlan
}

type Disc struct {
	Format string
	Hash   string
}

type MediaInfo struct {
	Title string
	Year  int
}
