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
	finished := state.Phase == core.GrinchWon || state.Phase == core.GrinchDefeat || state.Phase == core.GrinchEscaped
	if len(rewards) == 0 && !finished {
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
	if finished {
		if err := recordFight(ctx, tx, scope, server, state); err != nil {
			return false, err
		}
	}
	for user, reward := range rewards {
		member := app.Member{ServerID: server, UserID: user}
		previous, err := loadPlayer(ctx, tx, member)
		if err != nil {
			return false, err
		}
		if reward.SpentGift != "" && previous.Inventory.Gifts(reward.SpentGift) < 1 {
			return false, core.ErrInsufficientGift
		}
		if reward.UsedShield && previous.Inventory.Shields < 1 {
			return false, errors.New("the shield was already used")
		}
		next := previous
		next.Inventory = previous.Inventory.AddGrinchReward(reward)
		update := app.GiftUpdate{Member: member, Previous: previous, Next: next, Reason: app.ReasonGrinchReward}
		saved, err := savePlayer(ctx, tx, update)
		if err != nil || !saved {
			return false, err
		}
		if err := recordHistory(ctx, tx, update, ""); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// recordFight keeps every finished fight, because grinch_encounters only holds the latest one of
// each channel. A finished fight that is saved again replaces its own entry.
func recordFight(ctx context.Context, writer playerWriter, scope, server string, state core.Grinch) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = writer.ExecContext(ctx, `
		INSERT INTO grinch_history (scope, fight_id, server_id, payload) VALUES (?, ?, ?, ?)
		ON CONFLICT (scope, fight_id) DO UPDATE SET payload = excluded.payload
	`, scope, state.ID, server, payload)
	return err
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
