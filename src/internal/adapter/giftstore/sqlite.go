package giftstore

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func OpenSQLite(ctx context.Context, path string) (*SQLite, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	file := filepath.ToSlash(absolute)
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	dsn := url.URL{Scheme: "file", Path: file, RawQuery: "_busy_timeout=2000"}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gift_claims (
			server_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			total INTEGER NOT NULL CHECK (total > 0),
			last_claim TEXT NOT NULL,
			PRIMARY KEY (server_id, user_id)
		);
		CREATE TABLE IF NOT EXISTS advent_openings (
			server_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			year INTEGER NOT NULL,
			day INTEGER NOT NULL CHECK (day BETWEEN 1 AND 24),
			PRIMARY KEY (server_id, user_id, year, day)
		);
		CREATE TABLE IF NOT EXISTS grinch_encounters (scope TEXT PRIMARY KEY, revision INTEGER NOT NULL, payload BLOB NOT NULL)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Load(ctx context.Context, member app.Member) (core.GiftRecord, error) {
	var record core.GiftRecord
	var lastClaim string
	err := s.db.QueryRowContext(ctx, `
		SELECT total, last_claim FROM gift_claims WHERE server_id = ? AND user_id = ?
	`, member.ServerID, member.UserID).Scan(&record.Total, &lastClaim)
	if errors.Is(err, sql.ErrNoRows) {
		return core.GiftRecord{}, nil
	}
	if err != nil {
		return core.GiftRecord{}, err
	}
	record.LastClaim, err = time.Parse(time.RFC3339Nano, lastClaim)
	return record, err
}

func (s *SQLite) Save(ctx context.Context, member app.Member, total int64, record core.GiftRecord) (bool, error) {
	lastClaim := record.LastClaim.Format(time.RFC3339Nano)
	var result sql.Result
	var err error
	if total == 0 {
		result, err = s.db.ExecContext(ctx, `
			INSERT INTO gift_claims (server_id, user_id, total, last_claim)
			VALUES (?, ?, ?, ?) ON CONFLICT (server_id, user_id) DO NOTHING
		`, member.ServerID, member.UserID, record.Total, lastClaim)
	} else {
		result, err = s.db.ExecContext(ctx, `
			UPDATE gift_claims SET total = ?, last_claim = ?
			WHERE server_id = ? AND user_id = ? AND total = ?
		`, record.Total, lastClaim, member.ServerID, member.UserID, total)
	}
	if err != nil {
		return false, err
	}
	changed, err := result.RowsAffected()
	return changed == 1, err
}
