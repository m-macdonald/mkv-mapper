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

// func FindDiscByHash(logger *zap.SugaredLogger, defDir string, hash string) {
// 	filepath.WalkDir(defDir, getDirFunc(logger, hash))
// }

// func IndexDb(logger *zap.SugaredLogger, dbDir string) error {
// 	// var wg sync.WaitGroup
// 	discRecordChan := make(chan Disc)
// 	defer close(discRecordChan)
//
//
// 	for disc := range discRecordChan {
// 		logger.Infoln(disc)
// 	}
//
// 	return nil
// }

// func walkDir(logger *zap.SugaredLogger, dbDir string, discRecordChan chan Disc, wg *sync.WaitGroup) {
	// defer wg.Done()
	// err := filepath.WalkDir(dbDir, func(path string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	if d.IsDir() {
	// 		return fs.SkipDir
	// 	}
	//
	// 	if !(filepath.Base(path)[:4] == "disc" || filepath.Ext(path) == ".json") {
	// 		return fs.SkipDir
	// 	}
	//
	// 	rel, err := filepath.Rel(dbDir, path)
	// 	parts := filepath.SplitList(rel)
	// 	if len(parts) == 0 {
	// 		parts = filepath.SplitList
	// 	}
	//
	// 	var parentName, slug string
	// 	pathParts := filepath.SplitList(filepath.Dir(rel))
	// 	if len() {
	//
	// 	}
	//
	// 	dirPath := filepath.Dir(rel)
	// 	parentName = filepath.Base(filepath.Dir(dirPath))
	// 	slug = filepath.Base(dirPath)
	//
	// 	fileContents,  err := os.ReadFile(path)
	// 	if err != nil {
	// 		discRecordChan <- nil
	// 		return nil
	// 	}
	//
	// 	var disc Disc
	// 	err = json.Unmarshal(fileContents, &disc)
	// 	if err != nil {
	// 		discRecordChan <- nil
	// 		return nil
	// 	}
	//
	// 	discRecordChan <- nil
	// })
	//
	// if err != nil {
	// 	discRecordChan <- nil
	// }
// }


// func getDirFunc(logger *zap.SugaredLogger, hash string) fs.WalkDirFunc {
// 	return func(path string, d os.DirEntry, err error) error {
// 		if !d.Type().IsRegular() {
// 			return nil
// 		}
//
// 		if fileInfo, err := d.Info(); err == nil {
// 			logger.Infoln(fileInfo.Name())
// 			if strings.HasPrefix(fileInfo.Name(), "disc") && strings.HasSuffix(fileInfo.Name(), ".json") {
// 				logger.Infoln("File name matches")
// 				content, err := os.ReadFile(path)
// 				var disc Disc
// 				err = json.Unmarshal(content, &disc)
// 				if err != nil {
// 					// TODO Add more detail to this log
// 					logger.Errorln("Failed to unmarshal json contained in file", err)
// 				}
// 				
// 				if disc.ContentHash == hash {
// 					logger.Infoln("Found disc by hash: %s", path)
// 					return filepath.SkipAll
// 				}
// 			}
// 		}
//
// 		return nil
// 	}
// }

func LoadDef(logger *zap.SugaredLogger, defDir string, discNum int, slug string) (map[string]TitleSummary, error) {
    if discNum < 1 {
        return nil, fmt.Errorf("Disc Number must be greater than 0, but was %d", discNum)
    }

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
