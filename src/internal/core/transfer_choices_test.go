package core

import (
	"errors"
	"math"
	"testing"
)

func TestTransferCoal(t *testing.T) {
	sender := GiftRecord{Inventory: Inventory{Coal: 3, Snowballs: 2}, Aura: Aura{Bad: 95, Good: 25}}
	recipient := GiftRecord{Inventory: Inventory{Coal: 1, YellowSnowballs: 2}, Aura: Aura{Bad: 10, Good: 30}}
	from, to, err := TransferCoal(sender, recipient, 2)
	if err != nil || from.Inventory != (Inventory{Coal: 1, Snowballs: 2}) || from.Aura != (Aura{Bad: 100, Good: 25}) ||
		to.Inventory != (Inventory{Coal: 3, YellowSnowballs: 2}) || to.Aura != recipient.Aura {
		t.Fatalf("send coal = %+v, %+v, %v", from, to, err)
	}
	for _, step := range []struct {
		quantity  int64
		recipient GiftRecord
		err       error
	}{
		{0, recipient, ErrGiftQuantity},
		{4, recipient, ErrInsufficientCoal},
		{1, GiftRecord{Inventory: Inventory{Coal: math.MaxInt64}}, ErrCoalCapacity},
	} {
		from, to, err := TransferCoal(sender, step.recipient, step.quantity)
		if !errors.Is(err, step.err) || from != sender || to != step.recipient {
			t.Errorf("failed send = %+v, %+v, %v", from, to, err)
		}
	}
}

func TestTransferSnowball(t *testing.T) {
	sender := GiftRecord{Inventory: Inventory{Snowballs: 1, YellowSnowballs: 2}, Aura: Aura{Bad: 15}}
	recipient := GiftRecord{Inventory: Inventory{Snowballs: 4}, Aura: Aura{Good: 25}}
	from, to, err := TransferSnowball(sender, recipient)
	if err != nil || from.Inventory != (Inventory{YellowSnowballs: 2}) || to.Inventory != (Inventory{Snowballs: 5}) ||
		from.Aura != sender.Aura || to.Aura != recipient.Aura {
		t.Fatalf("throw snowball = %+v, %+v, %v", from, to, err)
	}
	if _, _, err := TransferSnowball(from, to); !errors.Is(err, ErrInsufficientSnowball) {
		t.Fatalf("empty snowball inventory: %v", err)
	}
	full := GiftRecord{Inventory: Inventory{Snowballs: math.MaxInt64}}
	if _, _, err := TransferSnowball(sender, full); !errors.Is(err, ErrSnowballCapacity) {
		t.Fatalf("full recipient: %v", err)
	}
}

func TestTransferYellowSnowball(t *testing.T) {
	sender := GiftRecord{Inventory: Inventory{YellowSnowballs: 1, Snowballs: 3}, Aura: Aura{Bad: 15}}
	recipient := GiftRecord{Inventory: Inventory{YellowSnowballs: 2, Snowballs: 4}, Aura: Aura{Good: 25}}
	from, to, err := TransferYellowSnowball(sender, recipient)
	if err != nil || from.Inventory != (Inventory{Snowballs: 3}) || to.Inventory != (Inventory{YellowSnowballs: 3, Snowballs: 4}) ||
		from.Aura != (Aura{Bad: 20}) || to.Aura != recipient.Aura {
		t.Fatalf("send yellow = %+v, %+v, %v", from, to, err)
	}
	if _, _, err := TransferYellowSnowball(from, to); !errors.Is(err, ErrInsufficientYellowSnowball) {
		t.Fatalf("empty yellow inventory: %v", err)
	}
}
