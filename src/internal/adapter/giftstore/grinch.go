package giftstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func (s *SQLite) LoadGrinch(ctx context.Context, scope string) (core.Grinch, int64, error) {
	var state core.Grinch
	var revision int64
	var payload []byte
	err := s.db.QueryRowContext(ctx, "SELECT revision, payload FROM grinch_encounters WHERE scope = ?", scope).Scan(&revision, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return state, 0, nil
	}
	if err != nil {
		return state, 0, err
	}
	err = json.Unmarshal(payload, &state)
	return state, revision, err
}

func (s *SQLite) SaveGrinch(ctx context.Context, scope string, revision int64, state core.Grinch) (bool, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return false, err
	}
	var result sql.Result
	if revision == 0 {
		result, err = s.db.ExecContext(ctx, "INSERT INTO grinch_encounters VALUES (?, 1, ?) ON CONFLICT(scope) DO NOTHING", scope, payload)
	} else {
		result, err = s.db.ExecContext(ctx, "UPDATE grinch_encounters SET revision = revision + 1, payload = ? WHERE scope = ? AND revision = ?", payload, scope, revision)
	}
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}
