package giftfixture

import (
	"context"
	"math/bits"
	"sync"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type Store struct {
	mu      sync.Mutex
	records map[app.Member]core.GiftRecord
	advent  map[adventMember]uint32
}

func (store *Store) Load(_ context.Context, member app.Member) (core.GiftRecord, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.records[member], nil
}

func (store *Store) Save(ctx context.Context, member app.Member, total int64, record core.GiftRecord) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if store.records[member].Total != total {
		return false, nil
	}
	if store.records == nil {
		store.records = make(map[app.Member]core.GiftRecord)
	}
	store.records[member] = record
	return true, nil
}

type adventMember struct {
	member app.Member
	year   int
}

func (store *Store) AdventDoors(_ context.Context, member app.Member, year int) (uint32, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.advent[adventMember{member, year}], nil
}

func (store *Store) OpenAdventDoor(ctx context.Context, member app.Member, year, day int) (bool, int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, 0, err
	}
	if store.advent == nil {
		store.advent = make(map[adventMember]uint32)
	}
	key := adventMember{member, year}
	mask := uint32(1) << (day - 1)
	opened := store.advent[key]
	store.advent[key] = opened | mask
	return opened&mask == 0, bits.OnesCount32(opened | mask), nil
}
