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
	SaveGrinch(context.Context, string, int64, core.Grinch) (bool, error)
}

type GrinchDisplay interface {
	Announce(context.Context, core.Grinch) (string, time.Time, error)
	Update(context.Context, core.Grinch) error
}

type GrinchEncounter struct {
	store     GrinchStore
	scope     string
	now       func() time.Time
	random    func(int) int
	displayMu sync.Mutex
}

func NewGrinchEncounter(store GrinchStore, scope string, now func() time.Time, random func(int) int) *GrinchEncounter {
	return &GrinchEncounter{store: store, scope: scope, now: now, random: random}
}

func (e *GrinchEncounter) change(ctx context.Context, update func(*core.Grinch) bool) (core.Grinch, bool, error) {
	for ctx.Err() == nil {
		state, version, err := e.store.LoadGrinch(ctx, e.scope)
		if err != nil {
			return core.Grinch{}, false, err
		}
		if !update(&state) {
			return state, false, nil
		}
		saved, err := e.store.SaveGrinch(ctx, e.scope, version, state)
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
		g.Advance(now)
		// Keep an ended announcement until its final state has been shown.
		if (g.Dirty && g.MessageID != "") || !g.CanSpawn(now, activity, roll, manual) {
			return false
		}
		*g = core.Grinch{ID: fmt.Sprint(now.UnixNano()), Phase: core.GrinchPending, StartedAt: now, HP: 10000}
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
	reply := "This fight is no longer active."
	next := uint8(e.random(254) + 1) // Both emoji types appear.
	if move.Phase == core.GrinchAttack || move.Phase == core.GrinchFinalAttack {
		move.FinalBoard = uint8(e.random(14) + 1) // Both attack and gift tiles.
	}
	if move.Phase == core.GrinchFinalAttack {
		move.GiftPoints = e.random(5) + 1
	}
	state, _, err := e.change(ctx, func(g *core.Grinch) bool {
		if g.ID != id || g.MessageID == "" || g.MessageID != message {
			return false
		}
		advanced := g.Advance(move.At)
		if advanced && move.Phase != core.GrinchLobby {
			reply = "This part of the fight has ended."
			return true
		}
		move.NextBoard = next
		count := len(g.Accepted)
		reply = g.Move(move)
		return advanced || len(g.Accepted) != count
	})
	return state, reply, err
}

func (e *GrinchEncounter) Sync(ctx context.Context, display GrinchDisplay) (core.Grinch, error) {
	e.displayMu.Lock()
	defer e.displayMu.Unlock()
	state, _, err := e.change(ctx, func(g *core.Grinch) bool { return g.Advance(e.now()) })
	if err != nil || !state.Dirty || state.MessageID == "" {
		return state, err
	}
	if err := display.Update(ctx, state); err != nil {
		return state, err
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
