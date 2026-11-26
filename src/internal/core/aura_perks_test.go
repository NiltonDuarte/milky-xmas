package core

import (
	"testing"
	"time"
)

func TestAuraListsAndPerks(t *testing.T) {
	rules := DefaultAuraRules()
	for aura, want := range map[Aura]AuraList{25: NiceList, 24: NeutralList, 0: NeutralList, -24: NeutralList, -25: NaughtyList} {
		if got := rules.List(aura); got != want {
			t.Errorf("List(%d) = %q, want %q", aura, got, want)
		}
	}
	for _, tt := range []struct {
		aura  Aura
		gifts int64
		bonus DailyBonus
	}{{25, 4, DailyBonus{Gifts: 1}}, {1, 3, DailyBonus{}}, {0, 3, DailyBonus{Coal: 1}}, {-25, 3, DailyBonus{YellowSnowballs: 1, Coal: 1}}} {
		got, bonus := rules.AddDailyBonus(GiftRecord{Aura: tt.aura, Inventory: Inventory{Rare: 3}}, RareGift)
		if bonus != tt.bonus || got.Inventory.Rare != tt.gifts || got.Inventory.YellowSnowballs != tt.bonus.YellowSnowballs || got.Inventory.Coal != tt.bonus.Coal {
			t.Errorf("daily bonus at %d = %+v, %+v", tt.aura, got.Inventory, bonus)
		}
	}
	if !rules.RainDoubles(-25, 24) || rules.RainDoubles(-25, 25) || rules.RainDoubles(0, 0) || rules.RainDoubles(50, 0) {
		t.Error("only the Naughty list doubles, one time in four")
	}
	if rules.WaterPoints(25, 2) != 3 || rules.WaterPoints(24, 2) != 2 || rules.WaterPoints(-25, 2) != 1 || rules.WaterPoints(-50, 1) != 1 {
		t.Error("the Nice list waters more and the Naughty list less, but never nothing")
	}
	if perks := rules.GrinchPerks(0, DefaultTuning().Grinch); perks != nil {
		t.Errorf("neutral perks = %+v", perks)
	}
}

func TestGrinchAppliesListPerks(t *testing.T) {
	rules, aura := classicFight(DefaultTuning().Grinch), DefaultAuraRules()
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchSnowballs, StartedAt: now, HP: GrinchBossHP,
		Joined: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}, Accepted: map[string]bool{}}
	hit := func(member string, value Aura) {
		fight.Move(rules, GrinchMove{Phase: GrinchSnowballs, At: now, Member: member, RequestID: member, Perks: aura.GrinchPerks(value, rules)})
	}
	hit("nice", 50)
	if got := GrinchBossHP - fight.HP; got != 500 {
		t.Errorf("nice snowball = %d, want 500", got)
	}
	before := fight.HP
	hit("naughty", -50)
	if got := before - fight.HP; got != rules.SnowballDamage {
		t.Errorf("naughty snowball = %d, want %d", got, rules.SnowballDamage)
	}

	fight.Phase, fight.TimedAt, fight.Board, fight.DefenseRound, fight.DefensePlayers = GrinchDefense, now, 1, 1, 2
	miss := func(member, request string) string {
		return fight.Move(rules, GrinchMove{Phase: GrinchDefense, At: now, Member: member, RequestID: request, Slot: 7, DefenseRound: fight.DefenseRound, NextBoard: 1})
	}
	if reply := miss("nice", "miss-nice"); reply != SnowflakeHit(34) || fight.PlayerDamage["nice"] != 34 {
		t.Errorf("nice miss = %q, %d", reply, fight.PlayerDamage["nice"])
	}
	if reply := miss("naughty", "miss-naughty"); reply != SnowflakeHit(10) || fight.PlayerDamage["naughty"] != 10 {
		t.Errorf("naughty miss = %q, %d", reply, fight.PlayerDamage["naughty"])
	}

	fight.Phase = GrinchEscaped
	rewards := fight.LossRewards(rules)
	if rewards["naughty"].Coal != 0 || rewards["nice"].Coal != rules.LossCoal || rewards["naughty"].YellowSnowballs != 2 {
		t.Errorf("loss rewards = %+v", rewards)
	}
}

func TestAuraMovesAtMostTheDailyLimit(t *testing.T) {
	rules := DefaultAuraRules()
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	record := GiftRecord{}
	move := func(points int, at time.Time, want Aura) {
		t.Helper()
		after := record
		after.Aura = record.Aura.Move(1, points)
		record = rules.LimitDaily(record, after, at)
		if record.Aura != want {
			t.Fatalf("move %d = aura %d, want %d", points, record.Aura, want)
		}
	}
	move(EpicGiftPoints, now, 15)
	move(5, now, 15)
	move(-10, now, 5) // Each direction has its own limit.
	move(-10, now, 0)
	move(-10, now, 0)
	tomorrow := now.Add(12 * time.Hour) // 00:00 UTC.
	move(10, tomorrow, 10)
	move(10, tomorrow, 15)
	rules.DailyLimit = 0
	move(30, tomorrow, 45)
}
