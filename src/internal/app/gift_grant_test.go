package app

import (
	"errors"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGiveGift(t *testing.T) {
	store := &dropStoreFixture{records: make(map[Member]core.GiftRecord)}
	service := NewService(store, time.Now, func(int) int { return 0 })
	mod := Member{ServerID: "s", UserID: "mod"}
	player := Recipient{Member: Member{ServerID: "s", UserID: "p"}}
	if err := service.GiveGift(t.Context(), mod, player, core.RareGift, 3); err != nil {
		t.Fatal(err)
	}
	if err := service.GiveGift(t.Context(), mod, player, core.RareGift, 2); err != nil {
		t.Fatal(err)
	}
	items, _ := service.ViewInventory(t.Context(), player.Member)
	aura, _ := service.CheckAura(t.Context(), player.Member)
	if items.Rare != 5 || aura != (core.Aura{}) {
		t.Fatalf("recipient = %+v %+v", items, aura)
	}
	if modItems, _ := service.ViewInventory(t.Context(), mod); modItems != (core.Inventory{}) {
		t.Fatalf("moderator inventory changed: %+v", modItems)
	}
	for name, test := range map[string]struct {
		to       Recipient
		rarity   core.GiftRarity
		quantity int64
		err      error
	}{
		"self":     {Recipient{Member: mod}, core.NormalGift, 1, ErrInvalidRecipient},
		"bot":      {Recipient{Member: player.Member, IsBot: true}, core.NormalGift, 1, ErrInvalidRecipient},
		"rarity":   {player, "mythic", 1, core.ErrGiftRarity},
		"quantity": {player, core.NormalGift, 0, core.ErrGiftQuantity},
	} {
		if err := service.GiveGift(t.Context(), mod, test.to, test.rarity, test.quantity); !errors.Is(err, test.err) {
			t.Errorf("%s: err = %v, want %v", name, err, test.err)
		}
	}
	if items, _ := service.ViewInventory(t.Context(), player.Member); items.Rare != 5 || items.Normal != 0 {
		t.Fatalf("rejected grants changed inventory: %+v", items)
	}
}
