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
	aliceStart := core.GiftRecord{Inventory: core.Inventory{Snowballs: 2, Coal: 1}, Aura: core.Aura{Bad: 10}}
	store := &giftStoreFixture{records: map[Member]core.GiftRecord{
		alice: aliceStart,
		bob:   {Inventory: core.Inventory{Epic: 2, Coal: 3}, Aura: core.Aura{Good: 20}},
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
	bobStart := store.records[bob]
	if _, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball); !errors.Is(err, failed) ||
		store.records[bob] != bobStart || store.records[alice] != aliceStart {
		t.Fatalf("failed throw changed inventory: %+v, %v", store.records, err)
	}
	store.saveError = nil
	// Each throw takes one snowball from the sender and gives it to the recipient.
	for count := int64(1); count <= 2; count++ {
		got, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball)
		wantBob, wantAlice := bobStart, aliceStart
		wantBob.Inventory.Snowballs = count
		wantAlice.Inventory.Snowballs = 2 - count
		if err != nil || got != count || store.records[bob] != wantBob || store.records[alice] != wantAlice {
			t.Fatalf("throw = %d, %v, records %+v", got, err, store.records)
		}
	}
	// With none left, the throw is refused and nothing moves.
	bobNow, aliceNow := store.records[bob], store.records[alice]
	if _, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball); !errors.Is(err, core.ErrInsufficientSnowball) ||
		store.records[bob] != bobNow || store.records[alice] != aliceNow {
		t.Fatalf("throw without snowballs = %+v, %v", store.records, err)
	}
	// A full recipient inventory refuses the throw and keeps the sender's snowball.
	aliceNow.Inventory.Snowballs = 1
	bobNow.Inventory.Snowballs = math.MaxInt64
	store.records[alice], store.records[bob] = aliceNow, bobNow
	if _, err := service.ThrowSnowball(t.Context(), alice, Recipient{Member: bob}, core.NormalSnowball); !errors.Is(err, core.ErrSnowballCapacity) ||
		store.records[bob] != bobNow || store.records[alice] != aliceNow {
		t.Fatalf("full inventory = %+v, %v", store.records, err)
	}
}
