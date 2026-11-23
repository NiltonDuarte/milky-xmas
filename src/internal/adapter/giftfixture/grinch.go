package giftfixture

import (
	"context"
	"maps"
	"sync"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type grinchRecord struct {
	state   core.Grinch
	version int64
}

type GrinchStore struct {
	mu      sync.Mutex
	records map[string]grinchRecord
}

func (store *GrinchStore) LoadGrinch(_ context.Context, scope string) (core.Grinch, int64, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	record := store.records[scope]
	return cloneGrinch(record.state), record.version, nil
}

func (store *GrinchStore) SaveGrinch(ctx context.Context, scope string, version int64, state core.Grinch) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if store.records[scope].version != version {
		return false, nil
	}
	if store.records == nil {
		store.records = make(map[string]grinchRecord)
	}
	store.records[scope] = grinchRecord{state: cloneGrinch(state), version: version + 1}
	return true, nil
}

func cloneGrinch(state core.Grinch) core.Grinch {
	state.Joined = maps.Clone(state.Joined)
	state.Participants = maps.Clone(state.Participants)
	state.PlayerDamage = maps.Clone(state.PlayerDamage)
	state.DamageDealt = maps.Clone(state.DamageDealt)
	state.Gifts = maps.Clone(state.Gifts)
	state.Accepted = maps.Clone(state.Accepted)
	return state
}
