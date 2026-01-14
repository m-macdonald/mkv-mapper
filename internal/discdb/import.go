package discdb

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/blevesearch/bleve"
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
)

const DiscDbRepo = "https://github.com/TheDiscDb/data.git"

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

type Disc struct {
	Index					int 		`json:"Index"`
	Slug 					string 		`json:"Slug"`
	Name					string 		`json:"Name"`
	Format 					string 		`json:"Format"`
	ContentHash				string		`json:"ContentHash"`
	Titles					[]Title		`json:"Titles"`
}

type Title struct {
	Index					int			`json:"Index"`
	Comment					string 		`json:"Comment"`
	SourceFile				string		`json:"SourceFile"`
	SegmentMap				string		`json:"SegmentMap"`
	Duration				string		`json:"Duration"`
	Item					Item		`json:"Item"`
}

type Item struct {
	Title					string 		`json:"Title"`
	Type					string		`json:"Type"`
	Season					string		`json:"Season"`
	Episode					string		`json:"Episode"`
}

// var discJsonFileMatcher regexp.Regexp = *regexp.MustCompile("disc[0-9]*\\.json")

func Index(logger *zap.SugaredLogger) {
	discMapping := bleve.NewIndexMapping()
	// TODO: This path does not work. it creates a "$HOME" directory in the cwd
	index, err := bleve.New("./testing/discs.bleve", discMapping)
	if err != nil {
		logger.Errorln("Unable to establish Bleve index", err)

		return
	}
	indexBatch := index.NewBatch()
	index.Close()

	discRecordChan := make(chan Disc)
	_ = clone(logger)
	go collectDiscs(logger, discRecordChan)
	if err != nil {
		logger.Errorln("Well, something blew up", err)

		return
	}

	for disc := range discRecordChan {
		logger.Infoln("Writing disc to Bleve %v", disc)
		indexBatch.Index(fmt.Sprintf("%s-%d", disc.Slug, disc.Index), disc)
	}

	logger.Infoln("Writing batch")
	index.Batch(indexBatch)
	logger.Infoln("Batch write complete")
}

func clone(logger *zap.SugaredLogger) error {
	// TODO: Add an option to set the temp dir?
	_, err := git.PlainClone(os.TempDir(), false, &git.CloneOptions {
		URL: DiscDbRepo,
		Progress: os.Stdout,
	})

	if err != nil {
		logger.Error("Failed to clone Disc DB repo")
		
		return err
	}

	return nil
}

func collectDiscs(logger *zap.SugaredLogger, discRecordChan chan Disc) error {
	// var wg = sync.WaitGroup {}
	logger.Infoln("Beginning Walk")

	// walkDir(logger, discRecordChan, &wg)
	//
	// for disc := range discRecordChan {
	// 	logger.Infoln(disc.Slug, disc.Name)
	// }
	
	// TODO: Handle error
	_ = filepath.WalkDir(os.TempDir() + "/data", getDbDirWalker(logger, discRecordChan))

	close(discRecordChan)

	return nil
}

func walkDir(logger *zap.SugaredLogger, discRecordChan chan Disc, wg *sync.WaitGroup) {
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
}

func getDbDirWalker(logger *zap.SugaredLogger, discRecordChan chan Disc) fs.WalkDirFunc {
	return func(path string, d os.DirEntry, err error) error {
		// TODO: Clean this up
		// logger.Infoln("DirEntry: ", d)

		if d.IsDir() || 
			!(strings.HasPrefix(d.Name(), "disc") && 
			strings.HasSuffix(d.Name(), ".json")) {
			return nil
		}

		fileContents, err := os.ReadFile(path)
		if err != nil {
			logger.Errorln("Failed to read file contents", err)
		}

		var disc Disc
		err = json.Unmarshal(fileContents, &disc)
		if err != nil {
			logger.Warnln("Failed to unmarshal contents of %s. This is not a terminal error. Continuing...", path)

			return nil
		}

		discRecordChan <- disc

		return nil 
	}
}
