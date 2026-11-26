package core

import "math"

type GrinchReward struct {
	Snowballs       int64
	Coal            int64
	YellowSnowballs int64
	// SpentGift, when set, takes one gift of the rarity, which got the player back up.
	SpentGift GiftRarity
}

// Gifts is how many gifts of the rarity the inventory holds.
func (inventory Inventory) Gifts(rarity GiftRarity) int64 {
	if balance := inventory.balance(rarity); balance != nil {
		return *balance
	}
	return 0
}

func (inventory Inventory) AddGrinchReward(reward GrinchReward) Inventory {
	if balance := inventory.balance(reward.SpentGift); balance != nil && *balance > 0 {
		*balance--
	}
	inventory.Snowballs += min(reward.Snowballs, math.MaxInt64-inventory.Snowballs)
	inventory.Coal += min(reward.Coal, math.MaxInt64-inventory.Coal)
	inventory.YellowSnowballs += min(reward.YellowSnowballs, math.MaxInt64-inventory.YellowSnowballs)
	return inventory
}

// LossRewards returns the grant of each registered participant on a loss. A player's list can
// change the Coal.
func (g Grinch) LossRewards(rules GrinchRules) map[string]GrinchReward {
	if g.Phase != GrinchDefeat && g.Phase != GrinchEscaped {
		return nil
	}
	rewards := make(map[string]GrinchReward, len(g.Joined))
	for member := range g.Joined {
		rewards[member] = GrinchReward{Coal: g.perks(rules, member).LossCoal, YellowSnowballs: int64(len(g.Joined)) * rules.LossYellowPerGroupMember}
	}
	return rewards
}
