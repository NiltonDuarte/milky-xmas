package core

import (
	"cmp"
	"errors"
	"slices"
	"time"
)

const (
	TreeMaxLevel      = 100
	TreeStartLevel    = 10
	TreeWaterPoints   = 2
	TreeDecayInterval = time.Hour
	TreeWaterCooldown = 10 * time.Minute
)

var (
	ErrTreeFull     = errors.New("the tree is already full of water")
	ErrTreeCooldown = errors.New("the player watered the tree recently")
)

// TreeCooldownError says how long a player has to wait before watering again.
type TreeCooldownError struct{ Remaining time.Duration }

func (e TreeCooldownError) Error() string { return ErrTreeCooldown.Error() }
func (e TreeCooldownError) Unwrap() error { return ErrTreeCooldown }

// Tree is the server's Christmas tree. The zero value is a tree that has not been touched yet.
// Its water level falls by one point every TreeDecayInterval and each watering adds TreeWaterPoints.
type Tree struct {
	Level int
	// Since is when Level was last current. Decay counts whole intervals from here.
	Since     time.Time
	Waterings map[string]time.Time // Each player's last watering, kept for the cooldown.
	Waters    map[string]int       // Each player's total waterings.
}

// At returns the tree as it stands at now, after any decay.
func (t Tree) At(now time.Time) Tree {
	if t.Since.IsZero() {
		t.Level, t.Since = TreeStartLevel, now
		return t
	}
	if hours := int(now.Sub(t.Since) / TreeDecayInterval); hours > 0 {
		t.Level = max(0, t.Level-hours)
		t.Since = t.Since.Add(time.Duration(hours) * TreeDecayInterval)
		if t.Level == 0 {
			t.Since = now // A dry tree does not owe water for the time it spent dry.
		}
	}
	return t
}

// Water adds a watering by the member and returns the next tree. It does not change the receiver's maps.
func (t Tree) Water(now time.Time, member string) (Tree, error) {
	next := t.At(now)
	if next.Level >= TreeMaxLevel {
		return next, ErrTreeFull
	}
	if last, ok := t.Waterings[member]; ok && now.Sub(last) < TreeWaterCooldown {
		return next, TreeCooldownError{Remaining: TreeWaterCooldown - now.Sub(last)}
	}
	waterings := make(map[string]time.Time, len(t.Waterings)+1)
	for other, at := range t.Waterings {
		if now.Sub(at) < TreeWaterCooldown { // Older entries no longer matter.
			waterings[other] = at
		}
	}
	waterings[member] = now
	waters := make(map[string]int, len(t.Waters)+1)
	for other, count := range t.Waters {
		waters[other] = count
	}
	waters[member]++
	next.Level = min(TreeMaxLevel, next.Level+TreeWaterPoints)
	next.Since = now
	next.Waterings, next.Waters = waterings, waters
	return next, nil
}

// TreeStage groups the water level: 0 for 0-25%, 1 for 26-50%, 2 for 51-75%, 3 for 76-99%, 4 for 100%.
func TreeStage(level int) int {
	switch {
	case level >= TreeMaxLevel:
		return 4
	case level > 75:
		return 3
	case level > 50:
		return 2
	case level > 25:
		return 1
	default:
		return 0
	}
}

// TreeWaterer is a player and their number of waterings.
type TreeWaterer struct {
	Member string
	Count  int
}

// TopWaterers lists the players who watered the most, ties broken by ID.
func (t Tree) TopWaterers(limit int) []TreeWaterer {
	waterers := make([]TreeWaterer, 0, len(t.Waters))
	for member, count := range t.Waters {
		waterers = append(waterers, TreeWaterer{member, count})
	}
	slices.SortFunc(waterers, func(a, b TreeWaterer) int {
		if order := cmp.Compare(b.Count, a.Count); order != 0 {
			return order
		}
		return cmp.Compare(a.Member, b.Member)
	})
	return waterers[:min(limit, len(waterers))]
}
