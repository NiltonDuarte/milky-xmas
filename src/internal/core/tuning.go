package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GrinchBossHP is the Grinch's full health in hundredths of HP (100%). It is
// structural and not tunable: every other HP number is a share of it.
const GrinchBossHP = 10000

// RarityOdds are the chances, in percent, of each gift rarity. They add up to 100.
type RarityOdds struct {
	Normal int
	Rare   int
	Epic   int
}

// Roll picks a rarity from a uniform roll in [0, GiftRarityRolls).
func (odds RarityOdds) Roll(roll int) GiftRarity {
	switch {
	case roll < odds.Normal:
		return NormalGift
	case roll < odds.Normal+odds.Rare:
		return RareGift
	default:
		return EpicGift
	}
}

// GiftTuning holds the rarity odds shared by the daily claim and the drops.
type GiftTuning struct {
	DailyRarity RarityOdds
	// GoodDeedsThreshold is the aura, on the good side of the scale, from which BoostedRarity
	// applies to a Random Active Drop.
	GoodDeedsThreshold int
	BoostedRarity      RarityOdds
}

// DailyGiftRarity takes a uniform roll in [0, GiftRarityRolls).
func (tuning GiftTuning) DailyGiftRarity(roll int) GiftRarity {
	return tuning.DailyRarity.Roll(roll)
}

// ActiveDropRarity takes a uniform roll in [0, GiftRarityRolls). Players whose aura is at or
// above the threshold have a better chance of rare and epic gifts.
func (tuning GiftTuning) ActiveDropRarity(roll int, aura Aura) GiftRarity {
	if int(aura) < tuning.GoodDeedsThreshold {
		return tuning.DailyRarity.Roll(roll)
	}
	return tuning.BoostedRarity.Roll(roll)
}

// GiftRainTuning holds the numbers of a Gift Rain.
type GiftRainTuning struct {
	// MinMessages and MaxMessages bound, inclusively, the channel messages before a rain.
	MinMessages int
	MaxMessages int
	// ClaimsToEnd is how many collections end a rain.
	ClaimsToEnd int
	// ClaimCooldown is how long a player waits between rain collections.
	ClaimCooldown time.Duration
	// CoalPercent is the chance that a collection gives Coal.
	CoalPercent int
}

// DropTuning holds the numbers shared by both activity drops.
type DropTuning struct {
	Lifetime time.Duration
	// RandomActiveOneIn makes a Random Active Drop appear on one message in this many.
	RandomActiveOneIn int
	GiftRain          GiftRainTuning
}

// RandomActiveDropEligible reports whether a qualifying activity message starts a random
// drop. The roll is uniform in [0, RandomActiveOneIn).
func (tuning DropTuning) RandomActiveDropEligible(roll int) bool {
	return roll == 0
}

// GiftRainMessageRange is the number of distinct thresholds a rain can draw.
func (tuning DropTuning) GiftRainMessageRange() int {
	return tuning.GiftRain.MaxMessages - tuning.GiftRain.MinMessages + 1
}

// NextGiftRainThreshold converts a uniform roll in [0, GiftRainMessageRange) to the
// inclusive activity range of a Gift Rain.
func (tuning DropTuning) NextGiftRainThreshold(roll int) int {
	return tuning.GiftRain.MinMessages + roll
}

// RainPrizeIsCoal takes a uniform roll in [0, GiftRarityRolls).
func (tuning DropTuning) RainPrizeIsCoal(roll int) bool {
	return roll < tuning.GiftRain.CoalPercent
}

// GrinchRules holds the numbers of the Grinch fight. HP and damage are in hundredths of
// HP, so 7500 is 75% of the boss.
type GrinchRules struct {
	// SpawnCheckInterval is how often the bot rolls for an automatic appearance.
	SpawnCheckInterval time.Duration
	SpawnChancePercent int
	// ActivityWindow is how recently someone must have spoken in the channel.
	ActivityWindow     time.Duration
	CooldownAfterFight time.Duration

	// StartWindow is how long the announcement waits for the first Start press before the fight is
	// cancelled. RegistrationDuration is the countdown that begins at that press.
	StartWindow          time.Duration
	RegistrationDuration time.Duration
	EarlyStartPlayers    int
	// FightDuration is the combat clock that starts when the boss reaches SnowballsBossHP.
	FightDuration time.Duration

	// SnowballsBossHP, DefenseBossHP, and AttackBossHP are the boss HP at the end of each
	// phase. The final attack ends at zero.
	// SnowballsDuration is how long the group has to reach SnowballsBossHP. If it runs out the Grinch wins.
	SnowballsDuration time.Duration
	SnowballDamage    int
	SnowballsBossHP   int
	DefenseBossHP     int
	// DefenseMissHPLoss is the share, in percent, of a player's maximum HP that a miss costs.
	DefenseMissHPLoss int
	// DefenseMaxDefenses caps the safe defenses the group needs: one per player up to this.
	DefenseMaxDefenses int
	AttackDamage       int
	AttackHeavyDamage  int
	AttackBossHP       int

	FinalHitDamage int
	// FinalHitScore is the score points a final hit awards, regardless of the HP it removes. The
	// points are kept for each player and no leaderboard shows them yet.
	FinalHitScore       int
	GiftPointsMin       int
	GiftPointsMax       int
	FinalSnowballReward int64

	LossCoal int64
	// LossYellowPerGroupMember multiplies the number of registered players.
	LossYellowPerGroupMember int64
}

// DefenseTotal is the HP the defense phase removes in total, split across the players.
func (rules GrinchRules) DefenseTotal() int {
	return rules.SnowballsBossHP - rules.DefenseBossHP
}

// Tuning is every number that can be adjusted without changing the code.
type Tuning struct {
	Gifts  GiftTuning
	Drops  DropTuning
	Grinch GrinchRules
	Tree   TreeRules
	Quotes QuoteRules
}

// The default numbers, documented in docs/activity-drops-numbers.md and
// docs/grinch-fight-numbers.md.
const (
	GiftDropSlots           = 8
	RandomActiveDropRolls   = 20
	GiftRainMinimumMessages = 5
	GiftRainMessageRange    = 96
	GiftRainClaimLimit      = 10
	RainCoalPercent         = 25
	GoodDeedsBoostThreshold = 25
	GiftDropLifetime        = 10 * time.Minute
	GiftRainClaimCooldown   = time.Minute

	GrinchJoinDuration = time.Minute
	GrinchDuration     = time.Minute
)

// DefaultTuning returns the numbers the bot uses without a tuning file.
func DefaultTuning() Tuning {
	return Tuning{
		Gifts: GiftTuning{
			DailyRarity:        RarityOdds{Normal: 70, Rare: 25, Epic: 5},
			GoodDeedsThreshold: GoodDeedsBoostThreshold,
			BoostedRarity:      RarityOdds{Normal: 50, Rare: 35, Epic: 15},
		},
		Drops: DropTuning{
			Lifetime:          GiftDropLifetime,
			RandomActiveOneIn: RandomActiveDropRolls,
			GiftRain: GiftRainTuning{
				MinMessages:   GiftRainMinimumMessages,
				MaxMessages:   GiftRainMinimumMessages + GiftRainMessageRange - 1,
				ClaimsToEnd:   GiftRainClaimLimit,
				ClaimCooldown: GiftRainClaimCooldown,
				CoalPercent:   RainCoalPercent,
			},
		},
		Grinch: GrinchRules{
			SpawnCheckInterval:       time.Minute,
			SpawnChancePercent:       10,
			ActivityWindow:           10 * time.Minute,
			CooldownAfterFight:       30 * time.Minute,
			StartWindow:              GrinchJoinDuration,
			RegistrationDuration:     GrinchJoinDuration,
			EarlyStartPlayers:        3,
			FightDuration:            GrinchDuration,
			SnowballsDuration:        2 * time.Minute,
			SnowballDamage:           417,
			SnowballsBossHP:          7500,
			DefenseBossHP:            4500,
			DefenseMissHPLoss:        10,
			DefenseMaxDefenses:       6,
			AttackDamage:             667,
			AttackHeavyDamage:        700,
			AttackBossHP:             500,
			FinalHitDamage:           84,
			FinalHitScore:            200,
			GiftPointsMin:            1,
			GiftPointsMax:            5,
			FinalSnowballReward:      1,
			LossCoal:                 1,
			LossYellowPerGroupMember: 1,
		},
		Tree:   DefaultTreeRules(),
		Quotes: DefaultQuoteRules(),
	}
}

// Validate reports every number outside its allowed range. Each message starts with the
// key path of the tuning file so the error can be fixed without reading the code.
func (tuning Tuning) Validate() error {
	var problems []error
	bad := func(path, format string, args ...any) {
		problems = append(problems, fmt.Errorf("%s: %s", path, fmt.Sprintf(format, args...)))
	}
	odds := func(path string, o RarityOdds) {
		for _, chance := range []struct {
			name  string
			value int
		}{{"normal", o.Normal}, {"rare", o.Rare}, {"epic", o.Epic}} {
			if chance.value < 0 || chance.value > 100 {
				bad(path+"."+chance.name, "must be between 0 and 100, got %d", chance.value)
			}
		}
		if sum := o.Normal + o.Rare + o.Epic; sum != 100 {
			bad(path, "odds must add up to 100, got %d", sum)
		}
	}
	percent := func(path string, value int) {
		if value < 0 || value > 100 {
			bad(path, "must be between 0 and 100, got %d", value)
		}
	}
	positive := func(path string, d time.Duration) {
		if d <= 0 {
			bad(path, "must be longer than zero, got %s", d)
		}
	}
	atLeastZero := func(path string, d time.Duration) {
		if d < 0 {
			bad(path, "must not be negative, got %s", d)
		}
	}
	atLeast := func(path string, value, minimum int) {
		if value < minimum {
			bad(path, "must be at least %d, got %d", minimum, value)
		}
	}

	gifts := tuning.Gifts
	odds("gifts.daily_rarity", gifts.DailyRarity)
	if gifts.GoodDeedsThreshold < 0 || gifts.GoodDeedsThreshold > int(AuraMax) {
		bad("gifts.good_deeds_boost.threshold", "must be between 0 and %d, got %d", AuraMax, gifts.GoodDeedsThreshold)
	}
	odds("gifts.good_deeds_boost.rarity", gifts.BoostedRarity)

	drops := tuning.Drops
	positive("drops.lifetime", drops.Lifetime)
	atLeast("drops.random_active.one_in", drops.RandomActiveOneIn, 1)
	rain := drops.GiftRain
	atLeast("drops.gift_rain.min_messages", rain.MinMessages, 1)
	if rain.MaxMessages < rain.MinMessages {
		bad("drops.gift_rain.max_messages", "must be at least min_messages (%d), got %d", rain.MinMessages, rain.MaxMessages)
	}
	atLeast("drops.gift_rain.claims_to_end", rain.ClaimsToEnd, 1)
	atLeastZero("drops.gift_rain.claim_cooldown", rain.ClaimCooldown)
	percent("drops.gift_rain.coal_percent", rain.CoalPercent)

	g := tuning.Grinch
	positive("grinch.spawn.check_interval", g.SpawnCheckInterval)
	percent("grinch.spawn.chance_percent", g.SpawnChancePercent)
	atLeastZero("grinch.spawn.activity_window", g.ActivityWindow)
	atLeastZero("grinch.spawn.cooldown_after_fight", g.CooldownAfterFight)
	positive("grinch.registration.start_window", g.StartWindow)
	positive("grinch.registration.duration", g.RegistrationDuration)
	atLeast("grinch.registration.early_start_players", g.EarlyStartPlayers, 1)
	positive("grinch.fight_duration", g.FightDuration)
	positive("grinch.snowballs.duration", g.SnowballsDuration)

	hundredths := func(path string, value int) {
		if value <= 0 {
			bad(path, "must be above zero, got %s", FormatHundredths(value))
		}
	}
	hundredths("grinch.snowballs.damage", g.SnowballDamage)
	hundredths("grinch.attack.damage", g.AttackDamage)
	hundredths("grinch.attack.heavy_damage", g.AttackHeavyDamage)
	hundredths("grinch.final_attack.hit_damage", g.FinalHitDamage)
	hundredths("grinch.final_attack.hit_score", g.FinalHitScore)
	if g.AttackBossHP <= 0 || g.AttackBossHP >= g.DefenseBossHP || g.DefenseBossHP >= g.SnowballsBossHP || g.SnowballsBossHP >= GrinchBossHP {
		bad("grinch", "boss HP milestones must fall from 100 toward 0 (0 < attack.boss_hp_after < defense.boss_hp_after < snowballs.boss_hp_after < 100), got attack %s, defense %s, snowballs %s",
			FormatHundredths(g.AttackBossHP), FormatHundredths(g.DefenseBossHP), FormatHundredths(g.SnowballsBossHP))
	}
	if g.DefenseMissHPLoss < 1 || g.DefenseMissHPLoss > 100 {
		bad("grinch.defense.miss_hp_loss_percent", "must be between 1 and 100, got %d", g.DefenseMissHPLoss)
	}
	atLeast("grinch.defense.max_defenses", g.DefenseMaxDefenses, 1)
	atLeast("grinch.final_attack.gift_points.min", g.GiftPointsMin, 1)
	if g.GiftPointsMax < g.GiftPointsMin {
		bad("grinch.final_attack.gift_points.max", "must be at least min (%d), got %d", g.GiftPointsMin, g.GiftPointsMax)
	}
	for _, reward := range []struct {
		path  string
		value int64
	}{
		{"grinch.final_attack.snowballs_reward", g.FinalSnowballReward},
		{"grinch.loss_rewards.coal", g.LossCoal},
		{"grinch.loss_rewards.yellow_snowballs_per_group_member", g.LossYellowPerGroupMember},
	} {
		if reward.value < 0 {
			bad(reward.path, "must not be negative, got %d", reward.value)
		}
	}

	tree := tuning.Tree
	if tree.StartLevel < 0 || tree.StartLevel > TreeMaxLevel {
		bad("tree.start_level", "must be between 0 and %d, got %d", TreeMaxLevel, tree.StartLevel)
	}
	if tree.WaterPoints < 1 || tree.WaterPoints > TreeMaxLevel {
		bad("tree.water_points", "must be between 1 and %d, got %d", TreeMaxLevel, tree.WaterPoints)
	}
	if tree.WaterAura < 0 || tree.WaterAura > int(AuraMax) {
		bad("tree.water_aura", "must be between 0 and %d, got %d", AuraMax, tree.WaterAura)
	}
	positive("tree.decay_interval", tree.DecayInterval)
	atLeastZero("tree.water_cooldown", tree.WaterCooldown)
	if tree.EventEnd.IsZero() {
		bad("tree.event_end", "must be a date and time")
	}
	atLeast("quotes.every_messages", tuning.Quotes.EveryMessages, 1)
	atLeastZero("quotes.min_interval", tuning.Quotes.MinInterval)
	return errors.Join(problems...)
}

// FormatHundredths writes hundredths of a unit without trailing zeros: 7500 is "75",
// 525 is "5.25", and 750 is "7.5".
func FormatHundredths(value int) string {
	text := strconv.FormatFloat(float64(value)/100, 'f', 2, 64)
	text = strings.TrimRight(text, "0")
	return strings.TrimSuffix(text, ".")
}
