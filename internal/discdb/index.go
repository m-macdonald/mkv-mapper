package discdb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	_ "github.com/mattn/go-sqlite3"
)

const DiscDbRepo = "https://github.com/TheDiscDb/data.git"

type Disc struct {
	Index       int     `json:"Index"`
	Slug        string  `json:"Slug"`
	Name        string  `json:"Name"`
	Format      string  `json:"Format"`
	ContentHash string  `json:"ContentHash"`
	Titles      []Title `json:"Titles"`
}

type Title struct {
	Index      int    `json:"Index"`
	Comment    string `json:"Comment"`
	SourceFile string `json:"SourceFile"`
	SegmentMap string `json:"SegmentMap"`
	Duration   string `json:"Duration"`
	Item       Item   `json:"Item"`
}

type Item struct {
	Title   string `json:"Title"`
	Type    string `json:"Type"`
	Season  string `json:"Season"`
	Episode string `json:"Episode"`
}

func Index() error {
	db, err := sql.Open("sqlite3", "testing/disc.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO: Magic Number
	discRecordChan := make(chan Disc, 32)
	err = cloneRepo()
	if err != nil {
		return err
	}
	go collectDiscs(discRecordChan)
	// if err != nil {
	// 	logger.Errorln("Well, something blew up", err)
	//
	// 	return
	// }
	// _, err = db.Query(`CREATE TABLE IF NOT EXISTS disc (
	// 		hash  varchar(255) PRIMARY KEY,
	// 		metadata TEXT
	// 	)`)
	// if err != nil {
	// 	logger.Warnln("Unable to create table", err)
	//
	// 	return
	// }

	for disc := range discRecordChan {
		discJson, err := json.Marshal(disc)
		if err != nil {
			return fmt.Errorf("unable to marshal disc to JSON %w", err)
		}
		_, err = db.Query(`
			INSERT INTO disc (hash, metadata)
			VALUES (?, ?)
			`, disc.ContentHash, discJson)
		if err != nil {
			return fmt.Errorf("failed to insert disc to db %w", err)
		}
	}

	return nil
}

func cloneRepo() error {
	// TODO: Add an option to set the temp dir?
	repo, err := git.PlainOpen(os.TempDir())
	if errors.Is(err, git.ErrRepositoryNotExists) {
		_, err := git.PlainClone(os.TempDir(), false, &git.CloneOptions{
			URL:      DiscDbRepo,
			Progress: os.Stdout,
			Depth:    1,
		})

		return err
	} else if err != nil {
		return err
	} else {
		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		return worktree.Pull(&git.PullOptions{RemoteName: "origin"})
	}
}

func collectDiscs(discRecordChan chan Disc) error {
	defer close(discRecordChan)

	return filepath.WalkDir(
		os.TempDir()+"/data",
		func(path string, d os.DirEntry, err error) error {
			if d.IsDir() ||
				!(strings.HasPrefix(d.Name(), "disc") &&
					strings.HasSuffix(d.Name(), ".json")) {
				return nil
			}

			fileContents, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var disc Disc
			err = json.Unmarshal(fileContents, &disc)
			if err != nil {
				return err
			}

			discRecordChan <- disc

			return nil
		})
}
