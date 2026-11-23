package giftstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func (s *SQLite) LoadGuildSettings(ctx context.Context, guild string) (app.GuildSettings, error) {
	var settings app.GuildSettings
	var payload []byte
	err := s.db.QueryRowContext(ctx, "SELECT payload FROM guild_settings WHERE guild_id = ?", guild).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return settings, nil
	}
	if err != nil {
		return settings, err
	}
	err = json.Unmarshal(payload, &settings)
	return settings, err
}

func (s *SQLite) SaveGuildSettings(ctx context.Context, guild string, settings app.GuildSettings) error {
	payload, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO guild_settings (guild_id, payload) VALUES (?, ?)
		ON CONFLICT (guild_id) DO UPDATE SET payload = excluded.payload
	`, guild, payload)
	return err
}
