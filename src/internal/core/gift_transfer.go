package core

import (
	"errors"
	"math"
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

	points := 5
	switch rarity {
	case RareGift:
		points = 10
	case EpicGift:
		points = 30
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
	sender.Aura = sender.Aura.Move(quantity, -5)
	return sender, recipient, nil
}
