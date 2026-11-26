package app

import (
	"context"
	"errors"
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

// ErrGoodDeedNotSaved reports that a watering counted but the player's aura did not change.
var ErrGoodDeedNotSaved = errors.New("the watering counted but the good deed was not saved")

// Tree owns the shared Christmas tree use cases.
type Tree struct {
	store TreeStore
	now   func() time.Time
	rules core.TreeRules
	// players, when set, holds the aura a watering improves.
	players GiftStore
	aura    core.AuraRules
}

// RewardTree returns a tree use case whose waterings count as good deeds on the players this
// service stores.
func (service Service) RewardTree(tree *Tree) *Tree {
	copied := *tree
	copied.players = service.gifts
	copied.aura = service.tuning.Aura
	return &copied
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

// Watering is what one watering did.
type Watering struct {
	Added      int // Percent of water added, including ListChange.
	ListChange int // What the player's list added to or took off the usual water.
	GoodDeeds  int
}

// Water adds the member's watering and returns the tree with what the watering did.
// On a refusal it returns the current tree with the error. With ErrGoodDeedNotSaved the watering
// still counted.
func (tree *Tree) Water(ctx context.Context, member Member) (core.Tree, Watering, error) {
	if member.ServerID == "" || member.UserID == "" {
		return core.Tree{}, Watering{}, ErrInvalidRecipient
	}
	rules := tree.rules
	if tree.players != nil {
		// Without the aura the watering still counts, and the failure shows up in the reward below.
		if record, err := tree.players.Load(ctx, member); err == nil {
			rules.WaterPoints = tree.aura.WaterPoints(record.Aura, rules.WaterPoints)
		}
	}
	next, err := tree.save(ctx, member, rules)
	if err != nil {
		return next, Watering{}, err
	}
	watering := Watering{Added: rules.WaterPoints, ListChange: rules.WaterPoints - tree.rules.WaterPoints}
	watering.GoodDeeds, err = tree.reward(ctx, member)
	if err != nil {
		return next, watering, fmt.Errorf("%w: %w", ErrGoodDeedNotSaved, err)
	}
	return next, watering, nil
}

func (tree *Tree) save(ctx context.Context, member Member, rules core.TreeRules) (core.Tree, error) {
	now := tree.now().UTC()
	for ctx.Err() == nil {
		stored, revision, err := tree.store.LoadTree(ctx, member.ServerID)
		if err != nil {
			return core.Tree{}, fmt.Errorf("load tree: %w", err)
		}
		next, err := stored.Water(rules, now, member.UserID)
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

// reward moves the member's aura toward good deeds and returns how far it moved.
func (tree *Tree) reward(ctx context.Context, member Member) (int, error) {
	if tree.players == nil || tree.rules.WaterAura == 0 {
		return 0, nil
	}
	for ctx.Err() == nil {
		record, err := tree.players.Load(ctx, member)
		if err != nil {
			return 0, fmt.Errorf("load aura: %w", err)
		}
		if !tree.aura.WaterCounts(record.Aura) {
			return 0, nil
		}
		updated := record
		updated.Aura = record.Aura.Move(1, tree.rules.WaterAura)
		updated = tree.aura.LimitDaily(record, updated, tree.now())
		if updated.Aura == record.Aura {
			return 0, nil
		}
		saved, err := tree.players.Save(ctx, GiftUpdate{Member: member, Previous: record, Next: updated, Reason: ReasonWater})
		if err != nil {
			return 0, fmt.Errorf("save aura: %w", err)
		}
		if saved {
			return int(updated.Aura - record.Aura), nil
		}
	}
	return 0, ctx.Err()
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
