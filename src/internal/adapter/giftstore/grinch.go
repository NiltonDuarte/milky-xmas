package giftstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
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

func (s *SQLite) SaveGrinch(ctx context.Context, scope string, revision int64, state core.Grinch, server string, rewards map[string]core.GrinchReward) (bool, error) {
	if len(rewards) == 0 {
		return saveGrinch(ctx, s.db, scope, revision, state)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	saved, err := saveGrinch(ctx, tx, scope, revision, state)
	if err != nil || !saved {
		return false, err
	}
	for user, reward := range rewards {
		member := app.Member{ServerID: server, UserID: user}
		previous, err := loadPlayer(ctx, tx, member)
		if err != nil {
			return false, err
		}
		next := previous
		next.Inventory = previous.Inventory.AddGrinchReward(reward)
		saved, err := savePlayer(ctx, tx, app.GiftUpdate{Member: member, Previous: previous, Next: next})
		if err != nil || !saved {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func saveGrinch(ctx context.Context, writer playerWriter, scope string, revision int64, state core.Grinch) (bool, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return false, err
	}
	var result sql.Result
	if revision == 0 {
		result, err = writer.ExecContext(ctx, "INSERT INTO grinch_encounters VALUES (?, 1, ?) ON CONFLICT(scope) DO NOTHING", scope, payload)
	} else {
		result, err = writer.ExecContext(ctx, "UPDATE grinch_encounters SET revision = revision + 1, payload = ? WHERE scope = ? AND revision = ?", payload, scope, revision)
	}
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}
