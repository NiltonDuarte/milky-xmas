package core

import (
	"math"
	"time"
)

// AuraList is the group a player's aura puts them in.
type AuraList string

const (
	NeutralList AuraList = ""
	NiceList    AuraList = "nice"
	NaughtyList AuraList = "naughty"
)

// AuraRules holds the adjustable numbers of the Nice and Naughty lists.
type AuraRules struct {
	// NiceAt is the aura from which a player is on the Nice list.
	NiceAt int
	// NaughtyAt is the aura, below zero, from which a player is on the Naughty list.
	NaughtyAt int
	// SendCooldown is how long a player's aura ignores further sends to the same player, in the same
	// direction, after one counted. Zero counts every send.
	SendCooldown time.Duration
	// DailyLimit is how far the aura can move in each direction per UTC day. Zero has no limit.
	DailyLimit int
	// DailyCoal is the Coal the daily claim adds for a player whose aura is zero or below.
	DailyCoal int
	Nice      NicePerks
	Naughty   NaughtyPerks
}

// NicePerks is what the Nice list gets, besides the boosted Random Active Drop.
type NicePerks struct {
	DailyExtraGifts        int
	WaterExtraPoints       int
	MissHPLossPercent      int // Replaces the HP a wrong defense costs.
	BossDamageBonusPercent int // Added to the damage of every attack on the Grinch.
}

// NaughtyPerks is what the Naughty list gets.
type NaughtyPerks struct {
	WaterLessPoints      int // Taken off the water a watering adds, which stays at least 1%.
	MissHPLossPercent    int // Replaces the HP a wrong defense costs.
	DailyYellowSnowballs int
	RainDoublePercent    int           // Chance that a Gift Rain gift comes twice.
	RainClaimCooldown    time.Duration // Replaces the wait between Gift Rain collections when shorter.
	LossCoal             int64
}

// DefaultAuraRules returns the numbers used without a tuning file.
func DefaultAuraRules() AuraRules {
	return AuraRules{
		NiceAt: 25, NaughtyAt: -25, SendCooldown: 24 * time.Hour, DailyLimit: 15, DailyCoal: 1,
		Nice:    NicePerks{DailyExtraGifts: 1, WaterExtraPoints: 1, MissHPLossPercent: 34, BossDamageBonusPercent: 20},
		Naughty: NaughtyPerks{WaterLessPoints: 1, MissHPLossPercent: 10, DailyYellowSnowballs: 1, RainDoublePercent: 25, RainClaimCooldown: 30 * time.Second},
	}
}

// List says which list the aura puts a player on.
func (rules AuraRules) List(aura Aura) AuraList {
	switch {
	case int(aura) >= rules.NiceAt:
		return NiceList
	case int(aura) <= rules.NaughtyAt:
		return NaughtyList
	default:
		return NeutralList
	}
}

// LimitDaily keeps the aura's move from before to after within the day's limit and counts it.
func (rules AuraRules) LimitDaily(before, after GiftRecord, now time.Time) GiftRecord {
	if rules.DailyLimit <= 0 || after.Aura == before.Aura {
		return after
	}
	if day := now.UTC().Format(time.DateOnly); after.AuraDay != day {
		after.AuraDay, after.AuraRaised, after.AuraLowered = day, 0, 0
	}
	move := int(after.Aura - before.Aura)
	if move > 0 {
		move = min(move, max(0, rules.DailyLimit-after.AuraRaised))
		after.AuraRaised += move
	} else {
		move = max(move, -max(0, rules.DailyLimit-after.AuraLowered))
		after.AuraLowered -= move
	}
	after.Aura = before.Aura + Aura(move)
	return after
}

// WaterCounts reports whether watering is a good deed for the player's aura. Below zero it is not,
// so that watering does not pull anyone out of the Naughty list.
func (rules AuraRules) WaterCounts(aura Aura) bool { return aura >= 0 }

// DailyBonus is what the aura added to a daily claim.
type DailyBonus struct {
	Gifts           int64
	YellowSnowballs int64
	Coal            int64
}

// AddDailyBonus adds the aura's bonus to a record that just claimed gifts of the rarity. What
// does not fit in the inventory is left out.
func (rules AuraRules) AddDailyBonus(claimed GiftRecord, rarity GiftRarity) (GiftRecord, DailyBonus) {
	var bonus DailyBonus
	switch rules.List(claimed.Aura) {
	case NiceList:
		if balance := claimed.Inventory.balance(rarity); balance != nil {
			bonus.Gifts = min(int64(rules.Nice.DailyExtraGifts), claimed.Inventory.remainingGiftCapacity())
			*balance += bonus.Gifts
		}
	case NaughtyList:
		bonus.YellowSnowballs = min(int64(rules.Naughty.DailyYellowSnowballs), math.MaxInt64-claimed.Inventory.YellowSnowballs)
		claimed.Inventory.YellowSnowballs += bonus.YellowSnowballs
	}
	if claimed.Aura <= 0 {
		bonus.Coal = min(int64(rules.DailyCoal), math.MaxInt64-claimed.Inventory.Coal)
		claimed.Inventory.Coal += bonus.Coal
	}
	return claimed, bonus
}

// RainCooldown is how long the player waits between Gift Rain collections when others wait usual.
func (rules AuraRules) RainCooldown(aura Aura, usual time.Duration) time.Duration {
	if rules.List(aura) == NaughtyList {
		return min(usual, rules.Naughty.RainClaimCooldown)
	}
	return usual
}

// RainDoubles takes a uniform roll in [0, 100) and reports whether the player's Gift Rain gift
// comes twice.
func (rules AuraRules) RainDoubles(aura Aura, roll int) bool {
	return rules.List(aura) == NaughtyList && roll < rules.Naughty.RainDoublePercent
}

// WaterPoints is the water, in percent, the player's watering adds when others add points.
func (rules AuraRules) WaterPoints(aura Aura, points int) int {
	switch rules.List(aura) {
	case NiceList:
		return points + rules.Nice.WaterExtraPoints
	case NaughtyList:
		return max(1, points-rules.Naughty.WaterLessPoints)
	default:
		return points
	}
}

// GrinchPerks is how a fight treats one player on a list.
type GrinchPerks struct {
	MissHPLoss  int
	DamageBonus int // Percent added to attack damage.
	LossCoal    int64
}

// GrinchPerks returns the player's perks in a fight, or nil for a player on neither list.
func (rules AuraRules) GrinchPerks(aura Aura, fight GrinchRules) *GrinchPerks {
	perks := GrinchPerks{MissHPLoss: fight.DefenseMissHPLoss, LossCoal: fight.LossCoal}
	switch rules.List(aura) {
	case NiceList:
		perks.MissHPLoss, perks.DamageBonus = rules.Nice.MissHPLossPercent, rules.Nice.BossDamageBonusPercent
	case NaughtyList:
		perks.MissHPLoss, perks.LossCoal = rules.Naughty.MissHPLossPercent, rules.Naughty.LossCoal
	default:
		return nil
	}
	return &perks
}
