package giftstore

import (
	"context"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func (s *SQLite) AdventDoors(ctx context.Context, member app.Member, year int) (uint32, error) {
	var opened uint32
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(1 << (day - 1)), 0) FROM advent_openings
		WHERE server_id = ? AND user_id = ? AND year = ?
	`, member.ServerID, member.UserID, year).Scan(&opened)
	return opened, err
}

func (s *SQLite) OpenAdventDoor(ctx context.Context, member app.Member, year, day int) (bool, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO advent_openings (server_id, user_id, year, day) VALUES (?, ?, ?, ?)
		ON CONFLICT (server_id, user_id, year, day) DO NOTHING
	`, member.ServerID, member.UserID, year, day)
	if err != nil {
		return false, 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, 0, err
	}
	var total int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM advent_openings WHERE server_id = ? AND user_id = ? AND year = ?
	`, member.ServerID, member.UserID, year).Scan(&total)
	if err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return rows == 1, total, nil
}
