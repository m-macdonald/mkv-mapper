package mkvmappertest

import (
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/makemkv"
)

func NewMakeMkvTitle(segmentMap string) makemkv.Title {
	return makemkv.Title{
		Segments:       segmentMap,
		TitleId:        1,
		OutputFilename: "file1-out.mkv",
		SourceFilename: "file1.mkv",
		OutputFileSize: 1000,
	}
}

func NewDiscTitle(segmentMap string) discdb.Title {
	return discdb.Title{
		SegmentMap: segmentMap,
	}
}

func NewDiscRecord(titles ...discdb.Title) discdb.DiscRecord {
	return discdb.DiscRecord{
		Disc: discdb.Disc{
			Titles: titles,
		},
	}
}

