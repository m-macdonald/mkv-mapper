package discdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type Cache interface {
	GetDiscRecord(ctx context.Context, discHash string) (*DiscRecord, bool, error)
	PutDiscRecord(ctx context.Context, discHash string, discRecord *DiscRecord) error
}

type SQLiteCache struct {
	db *sql.DB
}

func NewSQLiteCache(cachePath string) (*SQLiteCache, error) {
	db, err := sql.Open("sqlite3", cachePath)
	if err != nil {
		return nil, err
	}

	return &SQLiteCache{
		db: db,
	}, nil
}

func (s *SQLiteCache) GetDiscRecord(ctx context.Context, discHash string) (*DiscRecord, bool, error) {
	// This will need to change if we ever need to support multiple discs matching the same hash.
	row := s.db.QueryRowContext(
		ctx,
		`SELECT 
			record
		FROM disc_record
		WHERE hash = ?
		`, discHash)

	var record DiscRecord
	var data []byte
	err := row.Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}

		return nil, false, err
	}
	err = json.Unmarshal(data, &record)
	if err != nil {
		return nil, false, err
	}

	return &record, true, nil
}

func (s *SQLiteCache) PutDiscRecord(
	ctx context.Context,
	discHash string,
	discRecord *DiscRecord,
) error {
	recordJsonBytes, err := json.Marshal(discRecord)
	if err != nil {
		return err
	}

	slugSignature := discRecord.Release.Slug + "|" + discRecord.Disc.Slug
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO disc_record (slug_signature, hash, record)
		VALUES (?, ?, ?)
		ON CONFLICT(slug_signature) DO UPDATE SET 
			hash = excluded.hash,
			record = excluded.record
		`, slugSignature, discHash, recordJsonBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteCache) Close() error {
	if s.db != nil {
		return s.db.Close()
	}

	return nil
}
