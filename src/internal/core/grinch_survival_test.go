package core

import (
	"testing"
	"time"
)

func TestGrinchCooldownCounterattackAndRevival(t *testing.T) {
	rules := DefaultTuning().Grinch
	rules.ClickCooldown = 2 * time.Second
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "f", Phase: GrinchSnowballs, StartedAt: now, HP: GrinchBossHP,
		Joined: map[string]bool{"alice": true, "bob": true}, PlayerDamage: map[string]int{"bob": 80},
		DamageDealt: map[string]GrinchDamage{}, Accepted: map[string]bool{}}
	move := func(member, request string, after time.Duration, counter int, gift GiftRarity) string {
		return fight.Move(rules, GrinchMove{Phase: GrinchSnowballs, Member: member, RequestID: request,
			At: now.Add(after), CounterRoll: counter, Revive: gift})
	}

	// Eight hits for each of two players: 2500 / 16, rounded up.
	if move("alice", "1", 0, 89, ""); GrinchBossHP-fight.HP != 157 || fight.PlayerDamage["alice"] != 0 {
		t.Fatalf("first hit = %d damage, alice %d", GrinchBossHP-fight.HP, fight.PlayerDamage["alice"])
	}
	if reply := move("alice", "2", time.Second, 0, ""); fight.Accepted["2"] || reply != "Slow down! Wait a moment before your next move." {
		t.Fatalf("second click after 1s = %q", reply)
	}
	// A roll in the top tenth brings the Grinch's answer, which knocks bob out.
	if reply := move("bob", "3", 2*time.Second, 90, ""); reply != "Snowball hit for 1.57 HP. The Grinch hit you back: -20% HP." {
		t.Fatalf("counterattack = %q", reply)
	}
	if fight.PlayerDamage["bob"] != 100 || !fight.SomeoneDown() || !fight.DownAt.IsZero() || !fight.EndedAt.IsZero() {
		t.Fatalf("after the counterattack: bob %d, down at %v", fight.PlayerDamage["bob"], fight.DownAt)
	}
	// A standing player's gift brings back the one who fell.
	if reply := move("alice", "4", 4*time.Second, 0, EpicGift); reply != "You brought them back with 100% HP!" || fight.PlayerDamage["bob"] != 0 || fight.NextFallen() != "" {
		t.Fatalf("helping bob = %q, bob %d, next %q", reply, fight.PlayerDamage["bob"], fight.NextFallen())
	}
	if reply := move("alice", "4b", 4*time.Second, 0, EpicGift); fight.Accepted["4b"] || reply != "Nobody needs a gift right now." {
		t.Fatalf("with nobody down = %q", reply)
	}
	// Getting back up does not wait for the click cooldown.
	fight.PlayerDamage["bob"], fight.Fallen = 100, []string{"bob"}
	if reply := move("bob", "5", 2*time.Second, 0, RareGift); reply != "You are back with 50% HP!" || fight.PlayerDamage["bob"] != 50 {
		t.Fatalf("revival = %q, bob %d", reply, fight.PlayerDamage["bob"])
	}

	// With everybody down the Grinch waits for the last chance, and then wins.
	fight.PlayerDamage["alice"], fight.PlayerDamage["bob"] = 80, 100
	move("alice", "6", 12*time.Second, 99, "")
	if !fight.EndedAt.IsZero() || fight.LastChanceEnds(rules) != now.Add(32*time.Second) {
		t.Fatalf("everybody down: phase %s, last chance %v", fight.Phase, fight.LastChanceEnds(rules))
	}
	saved := fight.PlayerDamage["alice"]
	if move("alice", "7", 20*time.Second, 0, NormalGift); !fight.DownAt.IsZero() || fight.PlayerDamage["alice"] != 75 || saved != 100 {
		t.Fatalf("revival in the last chance: alice %d, down at %v", fight.PlayerDamage["alice"], fight.DownAt)
	}
	fight.PlayerDamage["alice"] = 80
	move("alice", "8", 22*time.Second, 99, "")
	if fight.Advance(rules, now.Add(41*time.Second)) || !fight.Advance(rules, now.Add(42*time.Second)) || fight.Phase != GrinchDefeat {
		t.Fatalf("after the last chance: phase %s", fight.Phase)
	}
}

func TestDefenseTakesSeveralHitsAndClicksDoNotWait(t *testing.T) {
	rules := DefaultTuning().Grinch
	if rules.ClickCooldown != 0 || rules.DefensesNeeded(1) != 8 || rules.DefensesNeeded(3) != 12 || rules.DefensesNeeded(20) != 30 {
		t.Fatalf("defaults: cooldown %s, defenses %d/%d/%d", rules.ClickCooldown, rules.DefensesNeeded(1), rules.DefensesNeeded(3), rules.DefensesNeeded(20))
	}
}

func TestFinalRowSnowballMoves(t *testing.T) {
	rules := DefaultTuning().Grinch
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{Phase: GrinchFinalAttack, StartedAt: now, TimedAt: now, HP: 500, FinalRound: 1, FinalBoard: 1, SnowballLeft: 3,
		Joined: map[string]bool{"alice": true}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}, Accepted: map[string]bool{}}
	// The Snowball is the second button, so the first is tile 0, a gift, and the third is tile 1, an attack.
	if fight.SnowballSlot() != 1 || !fight.FinalGift(0) || fight.FinalGift(2) {
		t.Fatalf("row: snowball at %d, gift %v %v", fight.SnowballSlot(), fight.FinalGift(0), fight.FinalGift(2))
	}
	reply := fight.Move(rules, GrinchMove{Phase: GrinchFinalAttack, Member: "alice", RequestID: "1", At: now, FinalRound: 1, Slot: 1, FinalBoard: 2, SnowballLeft: 4})
	if reply != rules.SnowballEarned() || fight.SnowballSlot() != 0 {
		t.Fatalf("pick = %q, snowball now at %d", reply, fight.SnowballSlot())
	}
}
