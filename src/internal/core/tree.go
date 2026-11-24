package core

import (
	"cmp"
	"errors"
	"slices"
	"time"
)

// TreeMaxLevel is a full tree, in percent of water. It is structural and not tunable.
const TreeMaxLevel = 100

// The default numbers of the tree, documented in the README.
const (
	TreeStartLevel    = 10
	TreeWaterPoints   = 2
	TreeDecayInterval = time.Hour
	TreeWaterCooldown = time.Hour
)

// TreeEventEnd is when the event ends by default: January 5, 2027 at 01:00 UTC. The tree stops
// decaying and cannot be watered from then on.
var TreeEventEnd = time.Date(2027, time.January, 5, 1, 0, 0, 0, time.UTC)

// TreeRules holds the adjustable numbers of the Christmas tree.
type TreeRules struct {
	// StartLevel is the water level, in percent, of a tree nobody has touched yet.
	StartLevel int
	// WaterPoints is the percent of water a watering adds.
	WaterPoints int
	// DecayInterval is how long the tree takes to lose one point of water.
	DecayInterval time.Duration
	// WaterCooldown is how long a player waits between waterings.
	WaterCooldown time.Duration
	// EventEnd is when the tree stops decaying and cannot be watered any more.
	EventEnd time.Time
}

// DefaultTreeRules returns the numbers the tree uses without a tuning file.
func DefaultTreeRules() TreeRules {
	return TreeRules{
		StartLevel: TreeStartLevel, WaterPoints: TreeWaterPoints,
		DecayInterval: TreeDecayInterval, WaterCooldown: TreeWaterCooldown, EventEnd: TreeEventEnd,
	}
}

// EventOver reports whether the event has ended.
func (rules TreeRules) EventOver(now time.Time) bool { return !now.Before(rules.EventEnd) }

var (
	ErrTreeFull      = errors.New("the tree is already full of water")
	ErrTreeCooldown  = errors.New("the player watered the tree recently")
	ErrTreeEventOver = errors.New("the event is over")
)

// TreeCooldownError says how long a player has to wait before watering again.
type TreeCooldownError struct{ Remaining time.Duration }

func (e TreeCooldownError) Error() string { return ErrTreeCooldown.Error() }
func (e TreeCooldownError) Unwrap() error { return ErrTreeCooldown }

// Tree is the server's Christmas tree. The zero value is a tree that has not been touched yet.
// Its water level falls by one point every DecayInterval and each watering adds WaterPoints.
type Tree struct {
	Level int
	// Since is when Level was last current. Decay counts whole intervals from here.
	Since     time.Time
	Waterings map[string]time.Time // Each player's last watering, kept for the cooldown.
	Waters    map[string]int       // Each player's total waterings.
}

// At returns the tree as it stands at now, after any decay.
func (t Tree) At(rules TreeRules, now time.Time) Tree {
	if rules.EventOver(now) {
		now = rules.EventEnd // The tree keeps the level it had when the event ended.
	}
	if t.Since.IsZero() {
		t.Level, t.Since = rules.StartLevel, now
		return t
	}
	if intervals := int(now.Sub(t.Since) / rules.DecayInterval); intervals > 0 {
		t.Level = max(0, t.Level-intervals)
		t.Since = t.Since.Add(time.Duration(intervals) * rules.DecayInterval)
		if t.Level == 0 {
			t.Since = now // A dry tree does not owe water for the time it spent dry.
		}
	}
	return t
}

// Water adds a watering by the member and returns the next tree. It does not change the receiver's maps.
func (t Tree) Water(rules TreeRules, now time.Time, member string) (Tree, error) {
	next := t.At(rules, now)
	if rules.EventOver(now) {
		return next, ErrTreeEventOver
	}
	if next.Level >= TreeMaxLevel {
		return next, ErrTreeFull
	}
	if last, ok := t.Waterings[member]; ok && now.Sub(last) < rules.WaterCooldown {
		return next, TreeCooldownError{Remaining: rules.WaterCooldown - now.Sub(last)}
	}
	waterings := make(map[string]time.Time, len(t.Waterings)+1)
	for other, at := range t.Waterings {
		if now.Sub(at) < rules.WaterCooldown { // Older entries no longer matter.
			waterings[other] = at
		}
	}
	waterings[member] = now
	waters := make(map[string]int, len(t.Waters)+1)
	for other, count := range t.Waters {
		waters[other] = count
	}
	waters[member]++
	next.Level = min(TreeMaxLevel, next.Level+rules.WaterPoints)
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
