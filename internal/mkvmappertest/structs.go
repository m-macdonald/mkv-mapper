package mkvmappertest

import (
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
)

func NewMakeMkvTitle(opts ...func(*makemkv.Title)) makemkv.Title {
	title := makemkv.Title{
		TitleId: 1,
		OutputFilename: "title.mkv",
		SourceFilename: "title.mpls",
		OutputFileSize: 1000,
	}
	for _, opt := range opts {
		opt(&title)
	}
	return title
}

func WithSegments(segmentMap string) func(*makemkv.Title) {
	return func(title *makemkv.Title) {
		title.Segments = segmentMap
	}
}

func WithOutputFilename(outputFilename string) func(*makemkv.Title) {
	return func(title *makemkv.Title) {
		title.OutputFilename = outputFilename
	}
}

func WithTitleId(titleId lines.TitleId) func(*makemkv.Title) {
	return func(title *makemkv.Title) {
		title.TitleId = titleId
	}
}

func NewDiscTitle(opts ...func(*discdb.Title)) discdb.Title {
	title := discdb.Title{}
	for _, opt := range opts {
		opt(&title)
	}
	return title
}

func WithSegmentMap(segmentMap string) func(*discdb.Title) {
	return func(title *discdb.Title) {
		title.SegmentMap = segmentMap
	}
}

func WithItem(item *discdb.Item) func(*discdb.Title) {
	return func(title *discdb.Title) {
		title.Item = item
	}
}

func NewDiscItem() *discdb.Item {
	return &discdb.Item{
		Title:   "Test Title",
		Season:  "1",
		Episode: "1",
		Type:    "Episode",
	}
}

func NewDiscRecord(opts ...func(*discdb.DiscRecord)) discdb.DiscRecord {
	record := discdb.DiscRecord{}
	for _, opt := range opts {
		opt(&record)
	}
	return record
}

func WithTitles(titles ...discdb.Title) func(*discdb.DiscRecord) {
	return func(record *discdb.DiscRecord) {
		record.Disc.Titles = titles
	}
}

func WithMedia(media discdb.Media) func(*discdb.DiscRecord) {
	return func(record *discdb.DiscRecord) {
		record.Media = media
	}
}
