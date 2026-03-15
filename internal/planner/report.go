package planner

type BuildReport struct {
	Warnings []PlanWarning
}

type PlanWarning struct {
	TitleId uint
	Code    WarningCode
	Message string
	Cause   error
}

type WarningCode string

const (
	WarningNamingFallback   WarningCode = "naming_fallback"
	WarningFilenameSuffixed WarningCode = "filename_suffixed"
)
