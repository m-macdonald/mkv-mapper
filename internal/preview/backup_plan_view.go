package preview

import (
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/util"
	"m-macdonald/mkv-mapper/internal/validation"
)

type BackupPlanView struct {
	Label     string
	OutputDir string
	Size      string

	CheckGroups []CheckGroupView
}

func BuildBackupPlanView(plan model.ValidatedBackupPlan) BackupPlanView {
	view := BackupPlanView{
		Label:     plan.Label,
		OutputDir: plan.OutputDir,
		Size:      util.FormatSize(plan.SumTitleSizes()),
	}

	for _, label := range groupOrder {
		results, ok := plan.Report.ResultsByGroup[label]
		if !ok {
			continue
		}
		var discLevel []validation.Result
		for _, r := range results {
			if r.RefID == "" {
				discLevel = append(discLevel, r)
			}
		}
		if len(discLevel) > 0 {
			view.CheckGroups = append(view.CheckGroups, CheckGroupView{
				Label:   label,
				Results: discLevel,
			})
		}
	}

	return view
}
