package model

import (
	"fmt"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/validation"
)

type BackupTitle struct {
	TitleId       lines.TitleId
	EstimatedSize uint64
}

type BackupPlan struct {
	DiscIdentity
	OutputDir  string
	KeepBackup bool
	Titles     []BackupTitle
}

func NewBackupPlan(identity DiscIdentity, outputDir string, keepBackup bool) BackupPlan {
	return BackupPlan{
		DiscIdentity: identity,
		OutputDir:    outputDir,
		KeepBackup:   keepBackup,
	}
}

func (b BackupPlan) SumTitleSizes() uint64 {
	var total uint64
	for _, title := range b.Titles {
		total += title.EstimatedSize
	}
	return total
}

type ValidatedBackupPlan struct {
	BackupPlan
	Report validation.Report
}

func (v ValidatedBackupPlan) Err() error {
	if v.Report.HasErrors() {
		return fmt.Errorf("plan has validation errors")
	}
	return nil
}
