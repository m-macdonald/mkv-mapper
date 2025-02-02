package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var summaryRegex = regexp.MustCompile("^(.*): (.*)$")

func loadDiscDbDef(config Config) (map[string]SummaryTitle, error) {
    
    var summaryFileName string
    if config.Disc > 9 {
        summaryFileName = fmt.Sprintf("disc%d-summary.txt", config.Disc)
    } else {
        summaryFileName = fmt.Sprintf("disc0%d-summary.txt", config.Disc)
    }


    path := filepath.Join(config.DiscDbDefs, config.Slug, summaryFileName)
    
    file, err := os.Open(path)
    if err != nil {
        // Do something
        return nil, err
    }
    defer file.Close()

    titles := make(map[string]SummaryTitle)

    scanner := bufio.NewScanner(file)
    title := SummaryTitle {}
    for scanner.Scan() {
        line := scanner.Text()
        matches := summaryRegex.FindStringSubmatch(line)

        if matches == nil || len(matches) == 0  {
            titles[title.SourceFileName] = title
            title = SummaryTitle {}
            continue
        } 

        switch matches[1] {
        case "Name":
            title.Name = matches[2]
        case "Source file name":
            title.SourceFileName = matches[2]
        case "Duration":
            title.Duration = matches[2]
        case "Chapters count":
            title.ChaptersCount = matches[2]
        case "Size":
            title.Size = matches[2]
        case "Segment count":
            title.SegmentCount = matches[2]
        case "Segment map":
            title.SegmentMap = matches[2]
        case "Type":
            title.Episode = matches[2]
        case "Season":
            title.Season = matches[2]
        case "Episode":
            title.Episode = matches[2]
        case "File name":
            title.FileName = matches[2]
        } 
    }

    return titles, nil
}
