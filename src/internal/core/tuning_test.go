package core

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDefaultTuningIsValid(t *testing.T) {
	if err := DefaultTuning().Validate(); err != nil {
		t.Fatalf("defaults are invalid: %v", err)
	}
}

func TestTuningValidate(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*Tuning)
		want   []string
	}{
		{"daily odds do not add up", func(x *Tuning) { x.Gifts.DailyRarity.Epic = 10 }, []string{"gifts.daily_rarity: odds must add up to 100, got 105"}},
		{"boosted odds negative", func(x *Tuning) { x.Gifts.BoostedRarity = RarityOdds{Normal: 110, Rare: -10, Epic: 0} },
			[]string{"gifts.good_deeds_boost.rarity.normal", "gifts.good_deeds_boost.rarity.rare"}},
		{"threshold above 100", func(x *Tuning) { x.Gifts.GoodDeedsThreshold = 101 }, []string{"gifts.good_deeds_boost.threshold"}},
		{"drop lifetime", func(x *Tuning) { x.Drops.Lifetime = 0 }, []string{"drops.lifetime"}},
		{"one in zero", func(x *Tuning) { x.Drops.RandomActiveOneIn = 0 }, []string{"drops.random_active.one_in"}},
		{"rain range reversed", func(x *Tuning) { x.Drops.GiftRain.MaxMessages = 4 }, []string{"drops.gift_rain.max_messages"}},
		{"rain claims", func(x *Tuning) { x.Drops.GiftRain.ClaimsToEnd = 0 }, []string{"drops.gift_rain.claims_to_end"}},
		{"rain cooldown may be zero", func(x *Tuning) { x.Drops.GiftRain.ClaimCooldown = 0 }, nil},
		{"rain coal", func(x *Tuning) { x.Drops.GiftRain.CoalPercent = 101 }, []string{"drops.gift_rain.coal_percent"}},
		{"spawn chance", func(x *Tuning) { x.Grinch.SpawnChancePercent = -1 }, []string{"grinch.spawn.chance_percent"}},
		{"spawn interval", func(x *Tuning) { x.Grinch.SpawnCheckInterval = 0 }, []string{"grinch.spawn.check_interval"}},
		{"early start", func(x *Tuning) { x.Grinch.EarlyStartPlayers = 0 }, []string{"grinch.registration.early_start_players"}},
		{"defense above snowballs", func(x *Tuning) { x.Grinch.DefenseBossHP = 8000 },
			[]string{"boss HP milestones must fall from 100 toward 0"}},
		{"attack milestone zero", func(x *Tuning) { x.Grinch.AttackBossHP = 0 }, []string{"boss HP milestones"}},
		{"snowball milestone at 100", func(x *Tuning) { x.Grinch.SnowballsBossHP = GrinchBossHP }, []string{"boss HP milestones"}},
		{"snowball damage", func(x *Tuning) { x.Grinch.SnowballDamage = 0 }, []string{"grinch.snowballs.damage"}},
		{"miss loss", func(x *Tuning) { x.Grinch.DefenseMissHPLoss = 0 }, []string{"grinch.defense.miss_hp_loss_percent"}},
		{"gift points", func(x *Tuning) { x.Grinch.GiftPointsMin, x.Grinch.GiftPointsMax = 3, 2 }, []string{"grinch.final_attack.gift_points.max"}},
		{"negative reward", func(x *Tuning) { x.Grinch.LossCoal = -1 }, []string{"grinch.loss_rewards.coal"}},
		{"tree start level above 100", func(x *Tuning) { x.Tree.StartLevel = 101 }, []string{"tree.start_level"}},
		{"tree start level may be zero", func(x *Tuning) { x.Tree.StartLevel = 0 }, nil},
		{"tree water points", func(x *Tuning) { x.Tree.WaterPoints = 0 }, []string{"tree.water_points"}},
		{"tree decay interval", func(x *Tuning) { x.Tree.DecayInterval = 0 }, []string{"tree.decay_interval"}},
		{"tree cooldown negative", func(x *Tuning) { x.Tree.WaterCooldown = -time.Minute }, []string{"tree.water_cooldown"}},
		{"tree cooldown may be zero", func(x *Tuning) { x.Tree.WaterCooldown = 0 }, nil},
		{"tree event end", func(x *Tuning) { x.Tree.EventEnd = time.Time{} }, []string{"tree.event_end"}},
		{"several problems", func(x *Tuning) { x.Drops.Lifetime = 0; x.Grinch.FightDuration = 0 }, []string{"drops.lifetime", "grinch.fight_duration"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tuning := DefaultTuning()
			tt.change(&tuning)
			err := tuning.Validate()
			if len(tt.want) == 0 {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() = nil")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %q", err, want)
				}
			}
		})
	}
}

func TestFormatHundredths(t *testing.T) {
	for value, want := range map[int]string{0: "0", 100: "1", 525: "5.25", 750: "7.5", 7500: "75", 10000: "100", 1: "0.01"} {
		if got := FormatHundredths(value); got != want {
			t.Errorf("FormatHundredths(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestGiftTuningRarityBands(t *testing.T) {
	tuning := DefaultTuning().Gifts
	tuning.DailyRarity = RarityOdds{Normal: 10, Rare: 10, Epic: 80}
	tuning.GoodDeedsThreshold = 20
	tuning.BoostedRarity = RarityOdds{Normal: 0, Rare: 0, Epic: 100}
	counts := make(map[GiftRarity]int)
	for roll := range GiftRarityRolls {
		counts[tuning.DailyGiftRarity(roll)]++
	}
	if counts[NormalGift] != 10 || counts[RareGift] != 10 || counts[EpicGift] != 80 {
		t.Fatalf("daily counts = %v", counts)
	}
	if got := tuning.ActiveDropRarity(0, 19); got != NormalGift {
		t.Errorf("below the threshold = %s", got)
	}
	if got := tuning.ActiveDropRarity(0, 20); got != EpicGift {
		t.Errorf("at the threshold = %s", got)
	}
}

func TestDropTuningRainRange(t *testing.T) {
	drops := DefaultTuning().Drops
	drops.GiftRain.MinMessages, drops.GiftRain.MaxMessages, drops.GiftRain.CoalPercent = 2, 4, 50
	if drops.GiftRainMessageRange() != 3 || drops.NextGiftRainThreshold(0) != 2 || drops.NextGiftRainThreshold(2) != 4 {
		t.Errorf("rain range = %d, thresholds %d..%d", drops.GiftRainMessageRange(), drops.NextGiftRainThreshold(0), drops.NextGiftRainThreshold(2))
	}
	if !drops.RainPrizeIsCoal(49) || drops.RainPrizeIsCoal(50) {
		t.Error("coal is not the first 50 of 100 rolls")
	}
}

// A fight with every number changed: the boss falls 100 → 90 → 60 → 10 → 0.
func TestGrinchFightWithChangedTuning(t *testing.T) {
	rules := DefaultTuning().Grinch
	rules.SnowballDamage, rules.SnowballsBossHP = 250, 9000
	rules.DefenseBossHP, rules.DefenseMissHPLoss = 6000, 50
	rules.AttackDamage, rules.AttackHeavyDamage, rules.AttackBossHP = 1000, 1500, 1000
	rules.FinalHitDamage, rules.FinalHitScore = 500, 300
	rules.GiftPointsMin, rules.GiftPointsMax = 2, 3
	rules.FightDuration = 30 * time.Second
	rules.EarlyStartPlayers = 2
	rules.FinalSnowballReward = 3
	rules.LossCoal, rules.LossYellowPerGroupMember = 2, 2
	if err := (Tuning{Gifts: DefaultTuning().Gifts, Drops: DefaultTuning().Drops, Grinch: rules, Tree: DefaultTreeRules()}).Validate(); err != nil {
		t.Fatal(err)
	}

	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchLobby, HP: GrinchBossHP, StartedAt: start,
		Joined: map[string]bool{}, Participants: map[string]bool{}, PlayerDamage: map[string]int{}, Accepted: map[string]bool{},
		DamageDealt: map[string]GrinchDamage{}}
	sequence := 0
	move := func(phase GrinchPhase, slot int, member string, at time.Time) string {
		sequence++
		return fight.Move(rules, GrinchMove{Phase: phase, Slot: slot, Member: member, RequestID: strconv.Itoa(sequence), At: at,
			NextBoard: 5, FinalBoard: 1, DefenseRound: fight.DefenseRound, FinalRound: fight.FinalRound, GiftPoints: 2})
	}

	move(GrinchLobby, 0, "a", start)
	if fight.Phase != GrinchLobby {
		t.Fatal("one player started combat with an early start of two")
	}
	move(GrinchLobby, 0, "b", start.Add(time.Second))
	if fight.Phase != GrinchSnowballs {
		t.Fatalf("second player did not start combat: %s", fight.Phase)
	}

	at := start.Add(2 * time.Second)
	for range 3 {
		move(GrinchSnowballs, 0, "a", at)
	}
	if reply := move(GrinchSnowballs, 0, "a", at); reply != "Snowball hit for 2.5 HP." {
		t.Errorf("snowball reply = %q", reply)
	}
	if fight.Phase != GrinchDefense || fight.Health(rules) != 90 || fight.TimedAt != at {
		t.Fatalf("defense started at %+v", fight)
	}
	if got := fight.Deadline(rules); !got.Equal(at.Add(30 * time.Second)) {
		t.Errorf("deadline = %s", got)
	}

	slot := func(on bool) int {
		for slot := range 8 {
			if (fight.Board&(1<<slot) != 0) == on {
				return slot
			}
		}
		t.Fatalf("board %08b has no slot with state %v", fight.Board, on)
		return 0
	}
	if reply := move(GrinchDefense, slot(false), "a", at); reply != "Snowflake! You lost 50% of your maximum HP." || fight.PlayerDamage["a"] != 50 {
		t.Fatalf("miss = %q, %+v", reply, fight.PlayerDamage)
	}
	move(GrinchDefense, slot(true), "a", at)
	if fight.Health(rules) != 75 {
		t.Errorf("one of two defenses left %v HP, want 75", fight.Health(rules))
	}
	move(GrinchDefense, slot(true), "b", at)
	if fight.Phase != GrinchAttack || fight.HP != 6000 {
		t.Fatalf("attack phase started at %+v", fight)
	}

	for range 3 {
		move(GrinchAttack, 7, "b", at) // slot 7 is off in board 5, so 10 HP per hit
	}
	for fight.Phase == GrinchAttack {
		move(GrinchAttack, 0, "b", at) // slot 0 is on in board 5, so 15 HP, clipped at the milestone
	}
	if fight.Phase != GrinchFinalAttack || fight.HP != 1000 {
		t.Fatalf("final attack started at %+v", fight)
	}

	// The final board starts as 🎁⚔️⚔️⚔️, so slot 0 awards gift points and slots 1-3 hit.
	if reply := move(GrinchFinalAttack, 0, "b", at); reply != "🎁 You earned 2 gift points!" || fight.Gifts["b"] != 2 {
		t.Errorf("gift reply = %q, %+v", reply, fight.Gifts)
	}
	if reply := move(GrinchFinalAttack, 1, "b", at); reply != "Final attack hit! You earned 3 damage points." {
		t.Errorf("final hit reply = %q", reply)
	}
	if reply := move(GrinchFinalAttack, GrinchSnowballSlot, "b", at); reply != "❄️ You earned 3 Snowballs!" {
		t.Errorf("snowball reply = %q", reply)
	}
	if got := fight.DamageDealt["b"].Hundredths; got < 300 {
		t.Errorf("final hit score = %d, want at least 300", got)
	}
	move(GrinchFinalAttack, 1, "b", at)
	if fight.Phase != GrinchWon || fight.HP != 0 {
		t.Fatalf("two final hits of 5 HP left %+v", fight)
	}
}

func TestGrinchChangedSpawnAndLossRewards(t *testing.T) {
	rules := DefaultTuning().Grinch
	rules.SpawnChancePercent, rules.ActivityWindow, rules.CooldownAfterFight = 50, time.Minute, time.Hour
	rules.LossCoal, rules.LossYellowPerGroupMember = 2, 2
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)

	var fresh Grinch
	if !fresh.CanSpawn(rules, now, now.Add(-time.Minute), 49, false) || fresh.CanSpawn(rules, now, now.Add(-time.Minute), 50, false) {
		t.Error("spawn chance is not the first 50 of 100 rolls")
	}
	if fresh.CanSpawn(rules, now, now.Add(-time.Minute-time.Second), 0, false) {
		t.Error("activity outside the window allowed a spawn")
	}
	ended := Grinch{ID: "old", EndedAt: now.Add(-30 * time.Minute)}
	if ended.CanSpawn(rules, now, now, 0, false) || !ended.CanSpawn(rules, now.Add(31*time.Minute), now.Add(31*time.Minute), 0, false) {
		t.Error("cooldown after a fight is not one hour")
	}

	lost := Grinch{Phase: GrinchEscaped, Joined: map[string]bool{"a": true, "b": true, "c": true}}
	for member, reward := range lost.LossRewards(rules) {
		if reward != (GrinchReward{Coal: 2, YellowSnowballs: 6}) {
			t.Errorf("%s reward = %+v", member, reward)
		}
	}
}

func TestTreeRulesChangeTheTree(t *testing.T) {
	rules := DefaultTreeRules()
	rules.StartLevel, rules.WaterPoints = 40, 7
	rules.DecayInterval, rules.WaterCooldown = 30*time.Minute, 10*time.Minute
	rules.EventEnd = treeStart.Add(10 * time.Hour)

	if got := (Tree{}).At(rules, treeStart); got.Level != 40 {
		t.Fatalf("new tree level = %d, want the configured 40", got.Level)
	}
	watered, err := Tree{}.Water(rules, treeStart, "alice")
	if err != nil || watered.Level != 47 {
		t.Fatalf("watering = %+v, %v, want 47", watered, err)
	}
	if got := watered.At(rules, treeStart.Add(90*time.Minute)).Level; got != 44 {
		t.Fatalf("level after 90 minutes = %d, want 44 (three 30 minute intervals)", got)
	}
	var cooldown TreeCooldownError
	if _, err := watered.Water(rules, treeStart.Add(4*time.Minute), "alice"); !errors.As(err, &cooldown) || cooldown.Remaining != 6*time.Minute {
		t.Fatalf("early watering = %v, want a 6 minute cooldown", err)
	}
	if _, err := watered.Water(rules, treeStart.Add(10*time.Minute), "alice"); err != nil {
		t.Fatalf("watering after the cooldown = %v", err)
	}
	if _, err := watered.Water(rules, rules.EventEnd, "bob"); !errors.Is(err, ErrTreeEventOver) {
		t.Fatalf("watering at the configured end = %v, want the event over error", err)
	}
}
