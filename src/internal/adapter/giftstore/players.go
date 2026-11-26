package giftstore

import (
	"context"
	"database/sql"
	"encoding/json"
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
		SELECT normal, rare, epic, snowballs, yellow_snowballs, coal, aura, last_claim, aura_day, aura_raised, aura_lowered
		FROM gift_players WHERE server_id = ? AND user_id = ?
	`, member.ServerID, member.UserID).Scan(
		&record.Inventory.Normal, &record.Inventory.Rare, &record.Inventory.Epic,
		&record.Inventory.Snowballs, &record.Inventory.YellowSnowballs, &record.Inventory.Coal, &record.Aura, &lastClaim,
		&record.AuraDay, &record.AuraRaised, &record.AuraLowered)
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

// LastAuraSend returns when a send to the recipient last moved the sender's aura that way.
func (s *SQLite) LastAuraSend(ctx context.Context, sender app.Member, recipientID string, good bool) (time.Time, error) {
	var at string
	err := s.db.QueryRowContext(ctx, `
		SELECT at FROM aura_sends WHERE server_id = ? AND user_id = ? AND peer_id = ? AND good = ?
	`, sender.ServerID, sender.UserID, recipientID, good).Scan(&at)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, at)
}

// PlayerHistory returns every kept inventory change of a server.
func (s *SQLite) PlayerHistory(ctx context.Context, server string) ([]app.HistoryEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id, reason, normal, rare, epic, snowballs, yellow_snowballs, coal
		FROM player_history WHERE server_id = ?
	`, server)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var history []app.HistoryEntry
	for rows.Next() {
		var entry app.HistoryEntry
		if err := rows.Scan(&entry.UserID, &entry.Reason, &entry.Delta.Normal, &entry.Delta.Rare, &entry.Delta.Epic,
			&entry.Delta.Snowballs, &entry.Delta.YellowSnowballs, &entry.Delta.Coal); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	return history, rows.Err()
}

// FightResults returns how every finished fight of a server ended.
func (s *SQLite) FightResults(ctx context.Context, server string) ([]core.GrinchPhase, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT payload FROM grinch_history WHERE server_id = ?", server)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []core.GrinchPhase
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var fight core.Grinch
		if err := json.Unmarshal(payload, &fight); err != nil {
			return nil, err
		}
		results = append(results, fight.Phase)
	}
	return results, rows.Err()
}

// ServerAuras returns the aura of every player the server has stored.
func (s *SQLite) ServerAuras(ctx context.Context, server string) ([]core.Aura, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT aura FROM gift_players WHERE server_id = ?", server)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var auras []core.Aura
	for rows.Next() {
		var aura core.Aura
		if err := rows.Scan(&aura); err != nil {
			return nil, err
		}
		auras = append(auras, aura)
	}
	return auras, rows.Err()
}

func (s *SQLite) Save(ctx context.Context, update app.GiftUpdate) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	saved, err := savePlayer(ctx, tx, update)
	if err != nil || !saved {
		return false, err
	}
	if err := recordHistory(ctx, tx, update, ""); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *SQLite) SaveTransfer(ctx context.Context, sender, recipient app.GiftUpdate) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	for _, side := range []struct {
		update app.GiftUpdate
		peer   string
	}{{sender, recipient.Member.UserID}, {recipient, sender.Member.UserID}} {
		saved, err := savePlayer(ctx, tx, side.update)
		if err != nil || !saved {
			return false, err
		}
		if err := recordHistory(ctx, tx, side.update, side.peer); err != nil {
			return false, err
		}
		if send := side.update.AuraSend; send != nil {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO aura_sends (server_id, user_id, peer_id, good, at) VALUES (?, ?, ?, ?, ?)
				ON CONFLICT (server_id, user_id, peer_id, good) DO UPDATE SET at = excluded.at
			`, side.update.Member.ServerID, side.update.Member.UserID, send.RecipientID, send.Good, claimTime(send.At)); err != nil {
				return false, err
			}
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
		INSERT INTO gift_players (server_id, user_id, normal, rare, epic, snowballs, yellow_snowballs, coal, aura, last_claim,
			aura_day, aura_raised, aura_lowered)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (server_id, user_id) DO UPDATE SET
			normal = excluded.normal, rare = excluded.rare, epic = excluded.epic,
			snowballs = excluded.snowballs, yellow_snowballs = excluded.yellow_snowballs, coal = excluded.coal,
			aura = excluded.aura, last_claim = excluded.last_claim,
			aura_day = excluded.aura_day, aura_raised = excluded.aura_raised, aura_lowered = excluded.aura_lowered
		WHERE gift_players.normal = ? AND gift_players.rare = ? AND gift_players.epic = ?
			AND gift_players.snowballs = ? AND gift_players.yellow_snowballs = ? AND gift_players.coal = ?
			AND gift_players.aura = ? AND gift_players.last_claim = ?
			AND gift_players.aura_day = ? AND gift_players.aura_raised = ? AND gift_players.aura_lowered = ?
	`, update.Member.ServerID, update.Member.UserID,
		next.Inventory.Normal, next.Inventory.Rare, next.Inventory.Epic, next.Inventory.Snowballs, next.Inventory.YellowSnowballs, next.Inventory.Coal,
		next.Aura, claimTime(next.LastClaim), next.AuraDay, next.AuraRaised, next.AuraLowered,
		previous.Inventory.Normal, previous.Inventory.Rare, previous.Inventory.Epic, previous.Inventory.Snowballs, previous.Inventory.YellowSnowballs, previous.Inventory.Coal,
		previous.Aura, claimTime(previous.LastClaim), previous.AuraDay, previous.AuraRaised, previous.AuraLowered)
	if err != nil {
		return false, err
	}
	changed, err := result.RowsAffected()
	return changed == 1, err
}

// recordHistory keeps what one saved change added to or took from a player, for the event's
// scoreboards. peer is the other player of a transfer. A change without a reason is not kept.
func recordHistory(ctx context.Context, writer playerWriter, update app.GiftUpdate, peer string) error {
	if update.Reason == "" {
		return nil
	}
	next, previous := update.Next.Inventory, update.Previous.Inventory
	_, err := writer.ExecContext(ctx, `
		INSERT INTO player_history (server_id, user_id, peer_id, reason, normal, rare, epic, snowballs, yellow_snowballs, coal, aura, at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	`, update.Member.ServerID, update.Member.UserID, peer, string(update.Reason),
		next.Normal-previous.Normal, next.Rare-previous.Rare, next.Epic-previous.Epic,
		next.Snowballs-previous.Snowballs, next.YellowSnowballs-previous.YellowSnowballs, next.Coal-previous.Coal,
		update.Next.Aura-update.Previous.Aura)
	return err
}

func claimTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
