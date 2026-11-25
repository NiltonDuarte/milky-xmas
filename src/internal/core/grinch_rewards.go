package core

import "math"

type GrinchReward struct {
	Snowballs       int64
	Coal            int64
	YellowSnowballs int64
}

func (inventory Inventory) AddGrinchReward(reward GrinchReward) Inventory {
	inventory.Snowballs += min(reward.Snowballs, math.MaxInt64-inventory.Snowballs)
	inventory.Coal += min(reward.Coal, math.MaxInt64-inventory.Coal)
	inventory.YellowSnowballs += min(reward.YellowSnowballs, math.MaxInt64-inventory.YellowSnowballs)
	return inventory
}

// LossRewards returns the same grant for each registered participant on a loss.
func (g Grinch) LossRewards(rules GrinchRules) map[string]GrinchReward {
	if g.Phase != GrinchDefeat && g.Phase != GrinchEscaped {
		return nil
	}
	rewards := make(map[string]GrinchReward, len(g.Joined))
	for member := range g.Joined {
		rewards[member] = GrinchReward{Coal: rules.LossCoal, YellowSnowballs: int64(len(g.Joined)) * rules.LossYellowPerGroupMember}
	}
	return rewards
}
