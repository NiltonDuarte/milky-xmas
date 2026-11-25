package giftstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func (s *SQLite) Load(ctx context.Context, member app.Member) (core.GiftRecord, error) {
	return loadPlayer(ctx, s.db, member)
}

type playerReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadPlayer(ctx context.Context, reader playerReader, member app.Member) (core.GiftRecord, error) {
	var record core.GiftRecord
	var lastClaim string
	err := reader.QueryRowContext(ctx, `
		SELECT normal, rare, epic, snowballs, yellow_snowballs, coal, aura, last_claim
		FROM gift_players WHERE server_id = ? AND user_id = ?
	`, member.ServerID, member.UserID).Scan(
		&record.Inventory.Normal, &record.Inventory.Rare, &record.Inventory.Epic,
		&record.Inventory.Snowballs, &record.Inventory.YellowSnowballs, &record.Inventory.Coal, &record.Aura, &lastClaim)
	if errors.Is(err, sql.ErrNoRows) {
		return core.GiftRecord{}, nil
	}
	if err != nil {
		return core.GiftRecord{}, err
	}
	if lastClaim != "" {
		record.LastClaim, err = time.Parse(time.RFC3339Nano, lastClaim)
	}
	return record, err
}

func (s *SQLite) Save(ctx context.Context, update app.GiftUpdate) (bool, error) {
	return savePlayer(ctx, s.db, update)
}

func (s *SQLite) SaveTransfer(ctx context.Context, sender, recipient app.GiftUpdate) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	for _, update := range []app.GiftUpdate{sender, recipient} {
		saved, err := savePlayer(ctx, tx, update)
		if err != nil || !saved {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

type playerWriter interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func savePlayer(ctx context.Context, writer playerWriter, update app.GiftUpdate) (bool, error) {
	next, previous := update.Next, update.Previous
	result, err := writer.ExecContext(ctx, `
		INSERT INTO gift_players (server_id, user_id, normal, rare, epic, snowballs, yellow_snowballs, coal, aura, last_claim)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (server_id, user_id) DO UPDATE SET
			normal = excluded.normal, rare = excluded.rare, epic = excluded.epic,
			snowballs = excluded.snowballs, yellow_snowballs = excluded.yellow_snowballs, coal = excluded.coal,
			aura = excluded.aura, last_claim = excluded.last_claim
		WHERE gift_players.normal = ? AND gift_players.rare = ? AND gift_players.epic = ?
			AND gift_players.snowballs = ? AND gift_players.yellow_snowballs = ? AND gift_players.coal = ?
			AND gift_players.aura = ? AND gift_players.last_claim = ?
	`, update.Member.ServerID, update.Member.UserID,
		next.Inventory.Normal, next.Inventory.Rare, next.Inventory.Epic, next.Inventory.Snowballs, next.Inventory.YellowSnowballs, next.Inventory.Coal,
		next.Aura, claimTime(next.LastClaim),
		previous.Inventory.Normal, previous.Inventory.Rare, previous.Inventory.Epic, previous.Inventory.Snowballs, previous.Inventory.YellowSnowballs, previous.Inventory.Coal,
		previous.Aura, claimTime(previous.LastClaim))
	if err != nil {
		return false, err
	}
	changed, err := result.RowsAffected()
	return changed == 1, err
}

func claimTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
