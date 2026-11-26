package giftstore

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"
	"strings"

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
	dsn := url.URL{Scheme: "file", Path: file, RawQuery: "_busy_timeout=2000&_txlock=immediate"}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := initSchema(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func initSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gift_players (
			server_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			normal INTEGER NOT NULL DEFAULT 0 CHECK (normal >= 0),
			rare INTEGER NOT NULL DEFAULT 0 CHECK (rare >= 0),
			epic INTEGER NOT NULL DEFAULT 0 CHECK (epic >= 0),
			aura INTEGER NOT NULL DEFAULT 0 CHECK (aura BETWEEN -50 AND 50),
			last_claim TEXT NOT NULL DEFAULT '',
			PRIMARY KEY (server_id, user_id)
		);
		CREATE TABLE IF NOT EXISTS advent_openings (
			server_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			year INTEGER NOT NULL,
			day INTEGER NOT NULL CHECK (day BETWEEN 1 AND 24),
			PRIMARY KEY (server_id, user_id, year, day)
		);
		CREATE TABLE IF NOT EXISTS grinch_encounters (scope TEXT PRIMARY KEY, revision INTEGER NOT NULL, payload BLOB NOT NULL);
		CREATE TABLE IF NOT EXISTS guild_settings (guild_id TEXT PRIMARY KEY, payload BLOB NOT NULL);
		CREATE TABLE IF NOT EXISTS christmas_trees (server_id TEXT PRIMARY KEY, revision INTEGER NOT NULL, payload BLOB NOT NULL);
		CREATE TABLE IF NOT EXISTS quote_posts (
			server_id TEXT NOT NULL, channel_id TEXT NOT NULL, at TEXT NOT NULL,
			PRIMARY KEY (server_id, channel_id)
		);
		CREATE TABLE IF NOT EXISTS aura_sends (
			server_id TEXT NOT NULL, user_id TEXT NOT NULL, peer_id TEXT NOT NULL, good INTEGER NOT NULL, at TEXT NOT NULL,
			PRIMARY KEY (server_id, user_id, peer_id, good)
		);
		CREATE TABLE IF NOT EXISTS grinch_history (
			scope TEXT NOT NULL, fight_id TEXT NOT NULL, server_id TEXT NOT NULL, payload BLOB NOT NULL,
			PRIMARY KEY (scope, fight_id)
		);
		CREATE TABLE IF NOT EXISTS player_history (
			id INTEGER PRIMARY KEY,
			server_id TEXT NOT NULL, user_id TEXT NOT NULL, peer_id TEXT NOT NULL, reason TEXT NOT NULL,
			normal INTEGER NOT NULL, rare INTEGER NOT NULL, epic INTEGER NOT NULL,
			snowballs INTEGER NOT NULL, yellow_snowballs INTEGER NOT NULL, coal INTEGER NOT NULL,
			aura INTEGER NOT NULL, at TEXT NOT NULL
		)`)
	if err != nil {
		return err
	}

	var legacy bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'gift_claims')
	`).Scan(&legacy)
	if err != nil {
		return err
	}
	if legacy {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO gift_players (server_id, user_id, normal, last_claim)
			SELECT server_id, user_id, total, last_claim FROM gift_claims;
			DROP TABLE gift_claims;
		`)
		if err != nil {
			return err
		}
	}
	var hasItems bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM pragma_table_info('gift_players') WHERE name = 'snowballs')
	`).Scan(&hasItems)
	if err != nil {
		return err
	}
	if !hasItems {
		_, err = tx.ExecContext(ctx, `
			ALTER TABLE gift_players ADD COLUMN snowballs INTEGER NOT NULL DEFAULT 0 CHECK (snowballs >= 0);
			ALTER TABLE gift_players ADD COLUMN coal INTEGER NOT NULL DEFAULT 0 CHECK (coal >= 0);
		`)
		if err != nil {
			return err
		}
	}
	var hasYellowSnowballs bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM pragma_table_info('gift_players') WHERE name = 'yellow_snowballs')
	`).Scan(&hasYellowSnowballs)
	if err != nil {
		return err
	}
	if !hasYellowSnowballs {
		_, err = tx.ExecContext(ctx, `
			ALTER TABLE gift_players ADD COLUMN yellow_snowballs INTEGER NOT NULL DEFAULT 0 CHECK (yellow_snowballs >= 0)
		`)
		if err != nil {
			return err
		}
	}
	// The single aura scale replaced the separate Bad and Good columns. Everyone starts neutral;
	// databases created before it keep the old columns unused.
	var hasAura bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM pragma_table_info('gift_players') WHERE name = 'aura')
	`).Scan(&hasAura)
	if err != nil {
		return err
	}
	if !hasAura {
		_, err = tx.ExecContext(ctx, `
			ALTER TABLE gift_players ADD COLUMN aura INTEGER NOT NULL DEFAULT 0 CHECK (aura BETWEEN -50 AND 50)
		`)
		if err != nil {
			return err
		}
	}
	var hasAuraDay bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM pragma_table_info('gift_players') WHERE name = 'aura_day')
	`).Scan(&hasAuraDay)
	if err != nil {
		return err
	}
	if !hasAuraDay {
		_, err = tx.ExecContext(ctx, `
			ALTER TABLE gift_players ADD COLUMN aura_day TEXT NOT NULL DEFAULT '';
			ALTER TABLE gift_players ADD COLUMN aura_raised INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE gift_players ADD COLUMN aura_lowered INTEGER NOT NULL DEFAULT 0
		`)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) Close() error {
	return s.db.Close()
}
