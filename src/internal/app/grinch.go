package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type GrinchStore interface {
	LoadGrinch(context.Context, string) (core.Grinch, int64, error)
	// SaveGrinch commits the new encounter revision and its inventory grants together.
	SaveGrinch(context.Context, string, int64, core.Grinch, string, map[string]core.GrinchReward) (bool, error)
}

type GrinchDisplay interface {
	Announce(context.Context, core.Grinch) (string, time.Time, error)
	Update(context.Context, core.Grinch) error
}

type GrinchEncounter struct {
	store     GrinchStore
	scope     string
	server    string
	rules     core.GrinchRules
	now       func() time.Time
	random    func(int) int
	displayMu sync.Mutex
}

func NewGrinchEncounter(store GrinchStore, server, scope string, rules core.GrinchRules, now func() time.Time, random func(int) int) *GrinchEncounter {
	return &GrinchEncounter{store: store, server: server, scope: scope, rules: rules, now: now, random: random}
}

// Rules returns the fight numbers this encounter applies, for adapters that render them.
func (e *GrinchEncounter) Rules() core.GrinchRules {
	return e.rules
}

func (e *GrinchEncounter) change(ctx context.Context, update func(*core.Grinch) bool) (core.Grinch, bool, error) {
	return e.changeWithRewards(ctx, func(state *core.Grinch) (bool, map[string]core.GrinchReward) {
		return update(state), nil
	})
}

func (e *GrinchEncounter) changeWithRewards(ctx context.Context, update func(*core.Grinch) (bool, map[string]core.GrinchReward)) (core.Grinch, bool, error) {
	for ctx.Err() == nil {
		state, version, err := e.store.LoadGrinch(ctx, e.scope)
		if err != nil {
			return core.Grinch{}, false, err
		}
		wasActive := state.EndedAt.IsZero()
		changed, rewards := update(&state)
		if !changed {
			return state, false, nil
		}
		if wasActive && !state.EndedAt.IsZero() {
			rewards = state.LossRewards(e.rules)
		}
		saved, err := e.store.SaveGrinch(ctx, e.scope, version, state, e.server, rewards)
		if err != nil || saved {
			return state, saved, err
		}
	}
	return core.Grinch{}, false, ctx.Err()
}

func (e *GrinchEncounter) Spawn(ctx context.Context, activity time.Time, manual bool, display GrinchDisplay) (bool, error) {
	now := e.now()
	roll := e.random(100)
	e.displayMu.Lock()
	defer e.displayMu.Unlock()
	state, reserved, err := e.change(ctx, func(g *core.Grinch) bool {
		g.Advance(e.rules, now)
		// Keep an ended announcement until its final state has been shown.
		if (g.Dirty && g.MessageID != "") || !g.CanSpawn(e.rules, now, activity, roll, manual) {
			return false
		}
		*g = core.Grinch{ID: fmt.Sprint(now.UnixNano()), Phase: core.GrinchPending, StartedAt: now, HP: core.GrinchBossHP}
		return true
	})
	if err != nil || !reserved {
		return false, err
	}
	message, started, err := display.Announce(ctx, state)
	if err != nil {
		return false, err // Retain the reservation during an uncertain send.
	}
	_, saved, err := e.change(ctx, func(g *core.Grinch) bool {
		if g.ID != state.ID || g.Phase != core.GrinchPending {
			return false
		}
		g.MessageID, g.StartedAt, g.Phase = message, started, core.GrinchLobby
		g.Joined = make(map[string]bool)
		g.Participants = make(map[string]bool)
		g.PlayerDamage = make(map[string]int)
		g.DamageDealt = make(map[string]core.GrinchDamage)
		g.Accepted = make(map[string]bool)
		g.Dirty = true
		return true
	})
	return saved, err
}

func (e *GrinchEncounter) Play(ctx context.Context, id, message string, move core.GrinchMove) (core.Grinch, string, error) {
	var reply string
	next := uint8(e.random(254) + 1) // Both emoji types appear.
	if move.Phase == core.GrinchAttack || move.Phase == core.GrinchFinalAttack {
		move.FinalBoard = uint8(e.random(14) + 1) // Both attack and gift tiles.
	}
	if move.Phase == core.GrinchFinalAttack {
		move.GiftPoints = e.random(e.rules.GiftPointsMax-e.rules.GiftPointsMin+1) + e.rules.GiftPointsMin
	}
	state, _, err := e.changeWithRewards(ctx, func(g *core.Grinch) (bool, map[string]core.GrinchReward) {
		if g.ID != id || g.MessageID == "" || g.MessageID != message {
			reply = "This fight is no longer active."
			return false, nil
		}
		advanced := g.Advance(e.rules, move.At)
		if advanced && move.Phase != core.GrinchLobby {
			reply = "This part of the fight has ended."
			return true, nil
		}
		move.NextBoard = next
		count := len(g.Accepted)
		reply = g.Move(e.rules, move)
		accepted := len(g.Accepted) != count
		if accepted && move.Phase == core.GrinchFinalAttack && move.Slot == core.GrinchSnowballSlot {
			return true, map[string]core.GrinchReward{move.Member: {Snowballs: e.rules.FinalSnowballReward}}
		}
		return advanced || accepted, nil
	})
	return state, reply, err
}

func (e *GrinchEncounter) CanJoin(ctx context.Context, id, message, member string, at time.Time) (bool, error) {
	state, _, err := e.store.LoadGrinch(ctx, e.scope)
	return err == nil && state.ID == id && state.MessageID == message && state.CanJoin(e.rules, member, at), err
}

func (e *GrinchEncounter) RefreshBoard(ctx context.Context, display GrinchDisplay) (core.Grinch, error) {
	return e.sync(ctx, display, true)
}

func (e *GrinchEncounter) Sync(ctx context.Context, display GrinchDisplay) (core.Grinch, error) {
	return e.sync(ctx, display, false)
}

func (e *GrinchEncounter) sync(ctx context.Context, display GrinchDisplay, refreshActive bool) (core.Grinch, error) {
	e.displayMu.Lock()
	defer e.displayMu.Unlock()
	state, _, err := e.change(ctx, func(g *core.Grinch) bool { return g.Advance(e.rules, e.now()) })
	if err != nil || state.MessageID == "" || (!state.Dirty && (!refreshActive || !state.EndedAt.IsZero())) {
		return state, err
	}
	if err := display.Update(ctx, state); err != nil {
		return state, err
	}
	if !state.Dirty {
		return state, nil
	}
	_, _, err = e.change(ctx, func(g *core.Grinch) bool {
		if g.ID != state.ID || g.Phase != state.Phase || len(g.Accepted) != len(state.Accepted) {
			return false
		}
		g.Dirty = false
		return true
	})
	return state, err
}
