package giftstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func (s *SQLite) LoadTree(ctx context.Context, server string) (core.Tree, int64, error) {
	var tree core.Tree
	var revision int64
	var payload []byte
	err := s.db.QueryRowContext(ctx, "SELECT revision, payload FROM christmas_trees WHERE server_id = ?", server).Scan(&revision, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return tree, 0, nil
	}
	if err != nil {
		return tree, 0, err
	}
	err = json.Unmarshal(payload, &tree)
	return tree, revision, err
}

func (s *SQLite) SaveTree(ctx context.Context, server string, revision int64, tree core.Tree) (bool, error) {
	payload, err := json.Marshal(tree)
	if err != nil {
		return false, err
	}
	var result sql.Result
	if revision == 0 {
		result, err = s.db.ExecContext(ctx, "INSERT INTO christmas_trees VALUES (?, 1, ?) ON CONFLICT(server_id) DO NOTHING", server, payload)
	} else {
		result, err = s.db.ExecContext(ctx, "UPDATE christmas_trees SET revision = revision + 1, payload = ? WHERE server_id = ? AND revision = ?", payload, server, revision)
	}
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}
