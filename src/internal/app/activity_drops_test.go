package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type dropStoreFixture struct {
	mu      sync.Mutex
	records map[Member]core.GiftRecord
}

func (store *dropStoreFixture) Load(_ context.Context, member Member) (core.GiftRecord, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.records[member], nil
}

func (store *dropStoreFixture) Save(_ context.Context, update GiftUpdate) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.records[update.Member] != update.Previous {
		return false, nil
	}
	store.records[update.Member] = update.Next
	return true, nil
}

func (store *dropStoreFixture) SaveTransfer(context.Context, GiftUpdate, GiftUpdate) (bool, error) {
	return false, errors.New("unused")
}

func (store *dropStoreFixture) AdventDoors(context.Context, Member, int) (uint32, error) {
	return 0, nil
}

func (store *dropStoreFixture) OpenAdventDoor(context.Context, Member, int, int) (bool, int, error) {
	return false, 0, errors.New("unused")
}

func TestActivityDropsObserveAndClaim(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	drops := NewActivityDrops(NewService(store, func() time.Time { return now }, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		case core.GiftRarityRolls:
			return 70
		default:
			t.Fatalf("random bound = %d", bound)
			return 0
		}
	}))
	for message := 1; message <= core.GiftRainMinimumMessages; message++ {
		drop := drops.Observe("guild", "channel")
		if message < core.GiftRainMinimumMessages && drop != nil {
			t.Fatalf("drop after message %d = %+v", message, drop)
		}
		if message == core.GiftRainMinimumMessages && (drop == nil || drop.Kind != core.GiftRainDrop) {
			t.Fatalf("rain drop = %+v", drop)
		}
		if message == core.GiftRainMinimumMessages {
			member := Member{ServerID: "guild", UserID: "alice"}
			award, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: drop.ID})
			if err != nil || award != (GiftDropAward{Kind: core.GiftRainDrop, Rarity: core.RareGift}) {
				t.Fatalf("first rain claim = %+v, %v", award, err)
			}
			if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: drop.ID}); !errors.Is(err, ErrDropCooldown) {
				t.Fatalf("early rain claim error = %v", err)
			}
			now = now.Add(GiftRainClaimCooldown)
			if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: drop.ID}); err != nil {
				t.Fatalf("rain claim at minute boundary = %v", err)
			}
			got, _ := store.Load(t.Context(), member)
			if got.Inventory.Rare != 2 {
				t.Fatalf("rain gifts = %+v", got.Inventory)
			}
		}
	}
}

func TestActivityDropOnlyAwardsOnceConcurrently(t *testing.T) {
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	drops := NewActivityDrops(NewService(store, time.Now, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return core.GiftRainMessageRange - 1
		case core.RandomActiveDropRolls, core.GiftRarityRolls:
			return 0
		default:
			return 0
		}
	}))
	drop := drops.Observe("guild", "channel")
	if drop == nil || drop.Kind != core.RandomActiveDrop {
		t.Fatalf("active drop = %+v", drop)
	}
	claim := GiftDropClaim{Member: Member{ServerID: "guild", UserID: "alice"}, DropID: drop.ID}
	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := drops.Claim(t.Context(), claim)
			results <- err
		}()
	}
	awards := 0
	for range 2 {
		if err := <-results; err == nil {
			awards++
		} else if !errors.Is(err, ErrDropCollected) {
			t.Fatalf("concurrent claim = %v", err)
		}
	}
	if awards != 1 {
		t.Fatalf("awards = %d", awards)
	}
	got, _ := store.Load(t.Context(), claim.Member)
	if got.Inventory.Normal != 1 {
		t.Fatalf("inventory = %+v", got.Inventory)
	}
}

func TestActivityDropExpires(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	drops := NewActivityDrops(NewService(store, func() time.Time { return now }, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return core.GiftRainMessageRange - 1
		case core.RandomActiveDropRolls, core.GiftRarityRolls:
			return 0
		default:
			return 0
		}
	}))
	drop := drops.Observe("guild", "channel")
	now = now.Add(GiftDropLifetime)
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: Member{ServerID: "guild", UserID: "alice"}, DropID: drop.ID}); !errors.Is(err, ErrDropUnavailable) {
		t.Fatalf("expired claim error = %v", err)
	}
}
