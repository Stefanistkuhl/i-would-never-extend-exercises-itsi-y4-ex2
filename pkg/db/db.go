package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"

	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var Schema string

type Store struct {
	*sqlc.Queries
	db *sql.DB
}

func openDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func InitIfNeeded() (*Store, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get dir: %w", err)
	}
	dbPath := filepath.Join(dir, "pcapStore.db")

	_, statErr := os.Stat(dbPath)
	if os.IsNotExist(statErr) {
		db, err := openDB(dbPath)
		if err != nil {
			return nil, err
		}
		tx, err := db.Begin()
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		if _, err := tx.Exec(Schema); err != nil {
			_ = tx.Rollback()
			_ = db.Close()
			return nil, fmt.Errorf("apply schema: %w", err)
		}
		if err := tx.Commit(); err != nil {
			_ = db.Close()
			return nil, err
		}
		return &Store{
			Queries: sqlc.New(db),
			db:      db,
		}, nil
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &Store{
		Queries: sqlc.New(db),
		db:      db,
	}, nil
}

func (s *Store) InsertCaptureWithStats(ctx context.Context,
	captureParams sqlc.InsertCaptureParams,
	statsParams sqlc.InsertCaptureStatsParams) (int64, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	q := sqlc.New(tx)

	captureID, err := q.InsertCapture(ctx, captureParams)
	if err != nil {
		return 0, err
	}

	statsParams.CaptureID = captureID
	err = q.InsertCaptureStats(ctx, statsParams)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return captureID, nil
}
