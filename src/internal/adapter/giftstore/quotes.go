package giftstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

var _ app.QuoteStore = (*SQLite)(nil)

func (s *SQLite) LastQuote(ctx context.Context, server, channel string) (time.Time, error) {
	var at string
	err := s.db.QueryRowContext(ctx, "SELECT at FROM quote_posts WHERE server_id = ? AND channel_id = ?", server, channel).Scan(&at)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, at)
}

func (s *SQLite) SaveLastQuote(ctx context.Context, server, channel string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO quote_posts (server_id, channel_id, at) VALUES (?, ?, ?)
		ON CONFLICT (server_id, channel_id) DO UPDATE SET at = excluded.at
	`, server, channel, at.UTC().Format(time.RFC3339Nano))
	return err
}
