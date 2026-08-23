package preview

import (
	"fmt"
	"sort"
	"strconv"

	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/util"
	"m-macdonald/mkv-mapper/internal/validation"
)

type Note struct {
	Status  validation.Status
	Message string
}

type TitleView struct {
	Source  string
	Target  string
	Size    uint64
	Matched bool
	Notes   []Note
}

type CheckGroupView struct {
	Label   validation.CheckGroupLabel
	Results []validation.Result
}

type PlanView struct {
	DiscName       string
	Year           int
	Format         string
	Hash           string
	Matched        []TitleView
	Unmatched      []TitleView
	UnmatchedRange string

	CheckGroups []CheckGroupView
}

// TODO:Remove or refactor. This is largely vestigial now that Backup and Rip have different plans.
// The concept of groups may be valuable in the future, so it may not need to be ripped out completely.
var groupOrder = []validation.CheckGroupLabel{validation.BackupLabel, validation.RipLabel}

func BuildPlanView(plan model.ValidatedPlan) PlanView {
	warningsByTitle := indexByTitleId(
		plan.BuildReport.Warnings,
		func(w model.PlanWarning) *string {
			s := strconv.Itoa(int(w.TitleId))
			return &s
		},
	)
	checkResultsByTitle := indexByTitleId(
		plan.ValidationReport.ResultsByGroup[validation.RipLabel],
		func(v validation.Result) *string { return &v.RefID },
	)
	fmt.Printf("ResultsByGroup:\n%v", plan.ValidationReport.ResultsByGroup)

	view := PlanView{
		DiscName: plan.MediaInfo.Title,
		Year:     plan.MediaInfo.Year,
		Format:   plan.DiscInfo.Format,
		Hash:     plan.DiscInfo.Hash,
	}

	for _, label := range groupOrder {
		results, ok := plan.ValidationReport.ResultsByGroup[label]
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
				Label: label,
				Results: discLevel,
			})
		}
	}

	for _, t := range plan.Titles {
		titleId := strconv.Itoa(int(t.TitleId))

		tv := TitleView{
			Source: t.SourcePlaylist,
			Target: t.FinalName,
			Size:   t.EstimatedSize,
		}

		for _, w := range warningsByTitle[titleId] {
			tv.Notes = append(tv.Notes, Note{Status: validation.StatusWarn, Message: w.Message})
		}
		for _, r := range checkResultsByTitle[titleId] {
			tv.Notes = append(tv.Notes, Note{Status: r.Status, Message: r.Message})
		}

		fmt.Printf("%v", tv)
		if t.IsMatched {
			tv.Matched = true
			view.Matched = append(view.Matched, tv)
		} else {
			view.Unmatched = append(view.Unmatched, tv)
		}
	}

	sortBySource(view.Matched)
	sortBySource(view.Unmatched)
	view.UnmatchedRange = sizeRange(view.Unmatched)
	return view
}

func sortBySource(titles []TitleView) {
	sort.Slice(titles, func(i, j int) bool {
		return titles[i].Source < titles[j].Source
	})
}

func sizeRange(titles []TitleView) string {
	if len(titles) == 0 {
		return ""
	}

	min, max := titles[0].Size, titles[0].Size
	for _, t := range titles[1:] {
		if t.Size < min {
			min = t.Size
		}

		if t.Size > max {
			max = t.Size
		}
	}

	return fmt.Sprintf("%s - %s", util.FormatSize(min), util.FormatSize(max))
}

func indexByTitleId[T any](items []T, getId func(T) *string) map[string][]T {
	index := map[string][]T{}
	for _, item := range items {
		id := getId(item)
		if id != nil {
			index[*id] = append(index[*id], item)
		}
	}
	return index
}
