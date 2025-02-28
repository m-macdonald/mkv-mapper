package discdb

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

var summaryRegex = regexp.MustCompile("^(.*): (.*)$")

type TitleSummary struct {
    Name                    string
    SourceFileName          string
    Duration                string
    ChaptersCount           string
    Size                    string
    SegmentCount            string
    SegmentMap              string
    Type                    string
    Season                  string
    Episode                 string
    FileName                string
}

func LoadDef(logger *zap.SugaredLogger, defDir string, discNum int, slug string) (map[string]TitleSummary, error) {
    // TODO: Check that discNum is greater than 0
    var summaryFileName string
    if discNum > 9 {
        summaryFileName = fmt.Sprintf("disc%d-summary.txt", discNum)
    } else {
        summaryFileName = fmt.Sprintf("disc0%d-summary.txt", discNum)
    }

    path := filepath.Join(defDir, slug, summaryFileName)
    logger.Debugf("Looking for summary file at %s\n", path)
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    titles := make(map[string]TitleSummary)

    scanner := bufio.NewScanner(file)
    title := TitleSummary {}
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Println(line)
        matches := summaryRegex.FindStringSubmatch(line)

        if matches == nil || len(matches) == 0  {
            titles[strings.TrimSpace(title.SourceFileName)] = title
            title = TitleSummary {}
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

    titles[title.SourceFileName] = title

    return titles, nil
}
