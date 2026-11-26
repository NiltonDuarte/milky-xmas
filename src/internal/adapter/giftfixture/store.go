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
	grinch  map[string]grinchRecord
	history []serverHistory
	fights  map[string]finishedFight
}

func (store *Store) Load(_ context.Context, member app.Member) (core.GiftRecord, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.records[member], nil
}

func (store *Store) Save(ctx context.Context, update app.GiftUpdate) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if store.records[update.Member] != update.Previous {
		return false, nil
	}
	store.save(update)
	return true, nil
}

func (store *Store) ServerAuras(_ context.Context, server string) ([]core.Aura, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	var auras []core.Aura
	for member, record := range store.records {
		if member.ServerID == server {
			auras = append(auras, record.Aura)
		}
	}
	return auras, nil
}

func (store *Store) SaveTransfer(ctx context.Context, sender, recipient app.GiftUpdate) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if store.records[sender.Member] != sender.Previous || store.records[recipient.Member] != recipient.Previous {
		return false, nil
	}
	store.save(sender)
	store.save(recipient)
	return true, nil
}

func (store *Store) save(update app.GiftUpdate) {
	if store.records == nil {
		store.records = make(map[app.Member]core.GiftRecord)
	}
	if update.Reason != "" {
		next, previous := update.Next.Inventory, update.Previous.Inventory
		store.history = append(store.history, serverHistory{server: update.Member.ServerID, entry: app.HistoryEntry{
			UserID: update.Member.UserID, Reason: update.Reason,
			Delta: core.Inventory{
				Normal: next.Normal - previous.Normal, Rare: next.Rare - previous.Rare, Epic: next.Epic - previous.Epic,
				Snowballs: next.Snowballs - previous.Snowballs, YellowSnowballs: next.YellowSnowballs - previous.YellowSnowballs,
				Coal: next.Coal - previous.Coal,
			},
		}})
	}
	store.records[update.Member] = update.Next
}

type serverHistory struct {
	server string
	entry  app.HistoryEntry
}

type finishedFight struct {
	server string
	phase  core.GrinchPhase
}

func (store *Store) PlayerHistory(_ context.Context, server string) ([]app.HistoryEntry, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	var history []app.HistoryEntry
	for _, kept := range store.history {
		if kept.server == server {
			history = append(history, kept.entry)
		}
	}
	return history, nil
}

func (store *Store) FightResults(_ context.Context, server string) ([]core.GrinchPhase, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	var results []core.GrinchPhase
	for _, fight := range store.fights {
		if fight.server == server {
			results = append(results, fight.phase)
		}
	}
	return results, nil
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
