package discdb

import (
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/mattn/go-sqlite3"
	// "go.uber.org/zap"
)

type Client struct {
	// logger zap.SugaredLogger
}

func NewClient() *Client {
	return &Client{}
}

func (d *Disc) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

func (c *Client) GetDisc(discHash string) (*Disc, error) {
	db, err := sql.Open("sqlite3", "testing/disc.db")
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(`
		SELECT metadata
		FROM disc
		WHERE hash = ?
		`, discHash)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var disc Disc
	err = row.Scan(&disc)
	if err != nil {
		return nil, err
	}

	return &disc, nil
}
