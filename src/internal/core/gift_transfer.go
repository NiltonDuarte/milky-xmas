package core

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

	if quantity > recipient.Inventory.remainingCapacity() {
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
	// Beyond 20 gifts the award already reaches the cap at any rarity.
	sender.Aura.Good = min(100, sender.Aura.Good+int(min(quantity, 20))*points)
	return sender, recipient, nil
}
