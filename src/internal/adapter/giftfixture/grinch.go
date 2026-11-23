package giftfixture

import (
	"context"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"maps"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type grinchRecord struct {
	state   core.Grinch
	version int64
}

func (store *Store) LoadGrinch(_ context.Context, scope string) (core.Grinch, int64, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	record := store.grinch[scope]
	return cloneGrinch(record.state), record.version, nil
}

func (store *Store) SaveGrinch(ctx context.Context, scope string, version int64, state core.Grinch, server string, rewards map[string]core.GrinchReward) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if store.grinch[scope].version != version {
		return false, nil
	}
	if store.grinch == nil {
		store.grinch = make(map[string]grinchRecord)
	}
	for user, reward := range rewards {
		member := app.Member{ServerID: server, UserID: user}
		record := store.records[member]
		record.Inventory = record.Inventory.AddGrinchReward(reward)
		store.save(app.GiftUpdate{Member: member, Next: record})
	}
	store.grinch[scope] = grinchRecord{state: cloneGrinch(state), version: version + 1}
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
