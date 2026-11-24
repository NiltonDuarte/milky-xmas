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

func dropRandom(coal, rarity int, active bool) func(int) int {
	rarityRolls := []int{coal, rarity}
	return func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			if active {
				return core.GiftRainMessageRange - 1
			}
			return 0
		case core.RandomActiveDropRolls:
			if active {
				return 0
			}
			return 1
		default:
			if active {
				return rarity
			}
			next := rarityRolls[0]
			rarityRolls = append(rarityRolls[1:], next)
			return next
		}
	}
}

func TestActiveDropGoodDeedsBoostSlotsAndSummary(t *testing.T) {
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	alice, bob, carol := Member{ServerID: "g", UserID: "alice"}, Member{ServerID: "g", UserID: "bob"}, Member{ServerID: "g", UserID: "carol"}
	store.records[alice] = core.GiftRecord{Aura: core.Aura{Good: 50}}
	store.records[bob] = core.GiftRecord{Aura: core.Aura{Good: 49}}
	drops := NewActivityDrops(NewService(store, time.Now, dropRandom(0, 80, true)))
	drop := drops.Observe("g", "c")
	if drop == nil || drop.Kind != core.RandomActiveDrop {
		t.Fatalf("drop = %+v", drop)
	}
	award, err := drops.Claim(t.Context(), GiftDropClaim{Member: alice, DropID: drop.ID, Slot: 3})
	if err != nil || award.Rarity != core.RareGift || award.Taken != 0b100 || award.Finished != nil {
		t.Fatalf("boosted claim = %+v, %v", award, err)
	}
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: bob, DropID: drop.ID, Slot: 3}); !errors.Is(err, ErrDropSlotTaken) {
		t.Fatalf("taken slot error = %v", err)
	}
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: alice, DropID: drop.ID, Slot: 4}); !errors.Is(err, ErrDropCollected) {
		t.Fatalf("second claim by the same player = %v", err)
	}
	if award, err := drops.Claim(t.Context(), GiftDropClaim{Member: bob, DropID: drop.ID, Slot: 4}); err != nil || award.Taken != 0b1100 {
		t.Fatalf("second slot claim = %+v, %v", award, err)
	}
	summary, ok := drops.Close(drop.ID)
	want := []DropClaimRecord{{UserID: "alice", Rarity: core.RareGift}, {UserID: "bob", Rarity: core.RareGift}}
	if !ok || len(summary.Claims) != 2 || summary.Claims[0] != want[0] || summary.Claims[1] != want[1] {
		t.Fatalf("summary = %+v, %v", summary, ok)
	}
	if _, ok := drops.Close(drop.ID); ok {
		t.Fatal("a drop closed twice")
	}
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: carol, DropID: drop.ID, Slot: 5}); !errors.Is(err, ErrDropUnavailable) {
		t.Fatalf("claim after close = %v", err)
	}
}

func TestActiveDropBoostChangesRarity(t *testing.T) {
	for _, tt := range []struct {
		good int
		roll int
		want core.GiftRarity
	}{{49, 50, core.NormalGift}, {50, 50, core.RareGift}, {50, 90, core.EpicGift}, {49, 90, core.RareGift}} {
		store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
		player := Member{ServerID: "g", UserID: "p"}
		store.records[player] = core.GiftRecord{Aura: core.Aura{Good: tt.good}}
		drops := NewActivityDrops(NewService(store, time.Now, dropRandom(0, tt.roll, true)))
		drop := drops.Observe("g", "c")
		award, err := drops.Claim(t.Context(), GiftDropClaim{Member: player, DropID: drop.ID})
		if err != nil || award.Rarity != tt.want {
			t.Errorf("good %d roll %d = %+v, %v; want %s", tt.good, tt.roll, award, err, tt.want)
		}
	}
}

func TestGiftRainEndsAfterTenClaimsWithCoal(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	// Rolls alternate: a coal roll, then (when it is not coal) the rarity roll.
	sequence := []int{0, 99, 60, 10, 5}
	index := 0
	random := func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			value := sequence[index%len(sequence)]
			index++
			return value
		}
	}
	drops := NewActivityDrops(NewService(store, func() time.Time { return now }, random))
	var rain *GiftDrop
	for range core.GiftRainMinimumMessages {
		rain = drops.Observe("g", "c")
	}
	if rain == nil || rain.Kind != core.GiftRainDrop {
		t.Fatalf("rain = %+v", rain)
	}
	var last GiftDropAward
	coal, gifts := 0, 0
	for i := range core.GiftRainClaimLimit {
		member := Member{ServerID: "g", UserID: string(rune('a' + i))}
		award, err := drops.Claim(t.Context(), GiftDropClaim{Member: member, DropID: rain.ID})
		if err != nil {
			t.Fatalf("claim %d: %v", i+1, err)
		}
		if i < core.GiftRainClaimLimit-1 && award.Finished != nil {
			t.Fatalf("rain finished after %d claims", i+1)
		}
		if award.Coal {
			coal++
		} else {
			gifts++
		}
		last = award
	}
	if last.Finished == nil || len(last.Finished.Claims) != core.GiftRainClaimLimit || coal == 0 || gifts == 0 {
		t.Fatalf("last = %+v, coal %d, gifts %d", last, coal, gifts)
	}
	if _, err := drops.Claim(t.Context(), GiftDropClaim{Member: Member{ServerID: "g", UserID: "late"}, DropID: rain.ID}); !errors.Is(err, ErrDropUnavailable) {
		t.Fatalf("eleventh claim = %v", err)
	}
	if _, ok := drops.Close(rain.ID); ok {
		t.Fatal("a finished rain closed again")
	}
	total := int64(0)
	for _, record := range store.records {
		total += record.Inventory.Coal
	}
	if total != int64(coal) {
		t.Fatalf("coal saved = %d, awarded %d", total, coal)
	}
}
