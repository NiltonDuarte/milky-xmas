package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type TreeStore interface {
	// LoadTree returns the server's tree and its revision. A missing tree is the zero value at revision 0.
	LoadTree(context.Context, string) (core.Tree, int64, error)
	// SaveTree stores the tree only if the stored revision still matches.
	SaveTree(context.Context, string, int64, core.Tree) (bool, error)
}

// Tree owns the shared Christmas tree use cases.
type Tree struct {
	store TreeStore
	now   func() time.Time
	rules core.TreeRules
}

// NewTree uses the default tree numbers. Use WithRules to apply a tuning file.
func NewTree(store TreeStore, now func() time.Time) *Tree {
	return &Tree{store: store, now: now, rules: core.DefaultTreeRules()}
}

// WithRules returns a tree use case that applies the given numbers instead of the defaults.
func (tree *Tree) WithRules(rules core.TreeRules) *Tree {
	copied := *tree
	copied.rules = rules
	return &copied
}

// Rules are the tree numbers this use case applies, for adapters that show them.
func (tree *Tree) Rules() core.TreeRules { return tree.rules }

// Now is the use case's current time, for adapters that show the tree.
func (tree *Tree) Now() time.Time { return tree.now().UTC() }

// View returns the server's tree as it stands now, without changing it.
func (tree *Tree) View(ctx context.Context, server string) (core.Tree, error) {
	stored, _, err := tree.store.LoadTree(ctx, server)
	return stored.At(tree.rules, tree.now().UTC()), err
}

// Water adds the member's watering. On a refusal it returns the current tree with the error.
func (tree *Tree) Water(ctx context.Context, member Member) (core.Tree, error) {
	if member.ServerID == "" || member.UserID == "" {
		return core.Tree{}, ErrInvalidRecipient
	}
	now := tree.now().UTC()
	for ctx.Err() == nil {
		stored, revision, err := tree.store.LoadTree(ctx, member.ServerID)
		if err != nil {
			return core.Tree{}, fmt.Errorf("load tree: %w", err)
		}
		next, err := stored.Water(tree.rules, now, member.UserID)
		if err != nil {
			return next, err
		}
		saved, err := tree.store.SaveTree(ctx, member.ServerID, revision, next)
		if err != nil {
			return core.Tree{}, fmt.Errorf("save tree: %w", err)
		}
		if saved {
			return next, nil
		}
	}
	return core.Tree{}, ctx.Err()
}

// MemoryTreeStore keeps trees in memory, for the local chat and tests.
type MemoryTreeStore struct {
	mu    sync.Mutex
	trees map[string]memoryTree
}

type memoryTree struct {
	tree     core.Tree
	revision int64
}

func (store *MemoryTreeStore) LoadTree(_ context.Context, server string) (core.Tree, int64, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	stored := store.trees[server]
	return stored.tree, stored.revision, nil
}

func (store *MemoryTreeStore) SaveTree(_ context.Context, server string, revision int64, tree core.Tree) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.trees[server].revision != revision {
		return false, nil
	}
	if store.trees == nil {
		store.trees = make(map[string]memoryTree)
	}
	store.trees[server] = memoryTree{tree: tree, revision: revision + 1}
	return true, nil
}
