package core

// GrantGift adds quantity gifts of the rarity without changing aura or cooldowns.
func GrantGift(recipient GiftRecord, rarity GiftRarity, quantity int64) (GiftRecord, error) {
	balance := recipient.Inventory.balance(rarity)
	switch {
	case balance == nil:
		return recipient, ErrGiftRarity
	case quantity <= 0:
		return recipient, ErrGiftQuantity
	case quantity > recipient.Inventory.remainingGiftCapacity():
		return recipient, ErrGiftCapacity
	}
	*balance += quantity
	return recipient, nil
}
