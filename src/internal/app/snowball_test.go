package app

import (
	"errors"
	"math"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestThrowSnowball(t *testing.T) {
	alice := Member{ServerID: "guild", UserID: "alice"}
	bob := Member{ServerID: "guild", UserID: "bob"}
	store := &giftStoreFixture{records: map[Member]core.GiftRecord{
		bob: {Inventory: core.Inventory{Epic: 2, Coal: 3}, Aura: core.Aura{Good: 20}},
	}}
	service := Service{gifts: store}
	for _, target := range []Recipient{
		{Member: alice}, {Member: bob, IsBot: true}, {Member: Member{ServerID: "another", UserID: "bob"}},
	} {
		if _, err := service.ThrowSnowball(t.Context(), alice, target, core.NormalSnowball); !errors.Is(err, ErrInvalidRecipient) {
			t.Errorf("invalid recipient %+v = %v", target, err)
		}
	}
	failed := errors.New("disk unavailable")
	store.saveError = failed
	before := store.records[bob]
	if _, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball); !errors.Is(err, failed) || store.records[bob] != before {
		t.Fatalf("failed throw changed inventory: %+v, %v", store.records[bob], err)
	}
	store.saveError = nil
	for count := int64(1); count <= 2; count++ {
		got, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball)
		want := before
		want.Inventory.Snowballs = count
		if err != nil || got != count || store.records[bob] != want || store.records[alice] != (core.GiftRecord{}) {
			t.Fatalf("throw = %d, %v, records %+v", got, err, store.records)
		}
	}
	before.Inventory.Snowballs = math.MaxInt64
	store.records[bob] = before
	if _, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball); !errors.Is(err, core.ErrSnowballCapacity) || store.records[bob] != before {
		t.Fatalf("full inventory = %+v, %v", store.records[bob], err)
	}
}
