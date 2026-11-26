package core

import (
	"errors"
	"math"
)

// How many points of good or bad deeds each item sent or thrown is worth.
const (
	NormalGiftPoints     = 5
	RareGiftPoints       = 10
	EpicGiftPoints       = 30
	CoalPoints           = 5
	YellowSnowballPoints = 5
)

func TransferGift(sender, recipient GiftRecord, rarity GiftRarity, quantity int64) (GiftRecord, GiftRecord, error) {
	from := sender.Inventory.balance(rarity)
	switch {
	case from == nil:
		return sender, recipient, ErrGiftRarity
	case quantity <= 0:
		return sender, recipient, ErrGiftQuantity
	case *from < quantity:
		return sender, recipient, ErrInsufficientGift
	}

	if quantity > recipient.Inventory.remainingGiftCapacity() {
		return sender, recipient, ErrGiftCapacity
	}
	*from -= quantity
	*recipient.Inventory.balance(rarity) += quantity

	points := NormalGiftPoints
	switch rarity {
	case RareGift:
		points = RareGiftPoints
	case EpicGift:
		points = EpicGiftPoints
	}
	sender.Aura = sender.Aura.Move(quantity, points)
	return sender, recipient, nil
}

var (
	ErrInsufficientCoal = errors.New("insufficient coal")
	ErrCoalCapacity     = errors.New("coal inventory is full")
)

func TransferCoal(sender, recipient GiftRecord, quantity int64) (GiftRecord, GiftRecord, error) {
	switch {
	case quantity <= 0:
		return sender, recipient, ErrGiftQuantity
	case sender.Inventory.Coal < quantity:
		return sender, recipient, ErrInsufficientCoal
	case quantity > math.MaxInt64-recipient.Inventory.Coal:
		return sender, recipient, ErrCoalCapacity
	}
	sender.Inventory.Coal -= quantity
	recipient.Inventory.Coal += quantity
	sender.Aura = sender.Aura.Move(quantity, -CoalPoints)
	return sender, recipient, nil
}
