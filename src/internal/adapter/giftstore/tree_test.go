package giftstore

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestTreePersistsAndChecksRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tree.sqlite")
	store, err := OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	if tree, revision, err := store.LoadTree(t.Context(), "s"); err != nil || revision != 0 || !tree.Since.IsZero() {
		t.Fatalf("missing tree = %+v, %d, %v", tree, revision, err)
	}
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	first, err := core.Tree{}.Water(core.DefaultTreeRules(), now, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if saved, err := store.SaveTree(t.Context(), "s", 0, first); !saved || err != nil {
		t.Fatalf("first save = %v, %v", saved, err)
	}
	if saved, err := store.SaveTree(t.Context(), "s", 0, first); saved || err != nil {
		t.Fatalf("stale first save = %v, %v", saved, err)
	}
	second, err := first.Water(core.DefaultTreeRules(), now.Add(time.Hour), "bob")
	if err != nil {
		t.Fatal(err)
	}
	if saved, err := store.SaveTree(t.Context(), "s", 1, second); !saved || err != nil {
		t.Fatalf("second save = %v, %v", saved, err)
	}
	if saved, err := store.SaveTree(t.Context(), "s", 1, second); saved || err != nil {
		t.Fatalf("stale second save = %v, %v", saved, err)
	}
	store.Close()

	store, err = OpenSQLite(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	got, revision, err := store.LoadTree(t.Context(), "s")
	if err != nil || revision != 2 || got.Level != second.Level || got.Waters["alice"] != 1 || got.Waters["bob"] != 1 ||
		!got.Waterings["bob"].Equal(now.Add(time.Hour)) {
		t.Fatalf("reloaded tree = %+v, %d, %v", got, revision, err)
	}
	if other, _, _ := store.LoadTree(t.Context(), "other"); !other.Since.IsZero() {
		t.Fatalf("another server's tree = %+v", other)
	}
}
