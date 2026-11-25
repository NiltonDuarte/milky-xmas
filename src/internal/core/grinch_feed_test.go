package core

import (
	"slices"
	"testing"
	"time"
)

func TestGrinchFeedRecordsAttacksAndTheBossResponse(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	fight := Grinch{ID: "f", Phase: GrinchSnowballs, StartedAt: now, HP: rules.SnowballsBossHP + rules.SnowballDamage,
		Joined: map[string]bool{"alice": true, "bob": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}}

	// The last snowball moves the fight to the defense phase and the Grinch says a line.
	fight.Move(rules, GrinchMove{Phase: GrinchSnowballs, Member: "alice", RequestID: "1", At: now, NextBoard: 0b00000001, Roll: 7})
	if want := []GrinchEvent{{Kind: EventHit, Member: "alice", Damage: float64(rules.SnowballDamage) / 100}}; !slices.Equal(fight.PlayerFeed, want) {
		t.Fatalf("player feed = %+v, want %+v", fight.PlayerFeed, want)
	}
	if want := []GrinchEvent{{Kind: EventPhase, Phase: GrinchDefense, Roll: 7}}; !slices.Equal(fight.BossFeed, want) {
		t.Fatalf("boss feed = %+v, want %+v", fight.BossFeed, want)
	}

	// Tile 1 is unsafe on this board and tile 0 is safe. A miss costs the player HP and the Grinch speaks.
	fight.Move(rules, GrinchMove{Phase: GrinchDefense, Slot: 1, Member: "bob", RequestID: "2", At: now, DefenseRound: fight.DefenseRound, NextBoard: 0b00000011, Roll: 8})
	miss := GrinchEvent{Kind: EventMiss, Member: "bob", Amount: rules.DefenseMissHPLoss, Roll: 8}
	if got := fight.BossFeed[len(fight.BossFeed)-1]; got != miss || fight.PlayerDamage["bob"] != rules.DefenseMissHPLoss {
		t.Fatalf("after a miss the boss feed ends with %+v, want %+v", got, miss)
	}
	if len(fight.PlayerFeed) != 1 {
		t.Fatalf("a miss is not an attack: %+v", fight.PlayerFeed)
	}

	// A safe defense is an attack on the Grinch. It is worth the defense total split between the players.
	fight.Move(rules, GrinchMove{Phase: GrinchDefense, Slot: 0, Member: "alice", RequestID: "3", At: now, DefenseRound: fight.DefenseRound, NextBoard: 0b00000101, Roll: 9})
	want := GrinchEvent{Kind: EventDefend, Member: "alice", Roll: 9, Damage: float64(rules.DefenseTotal()) / 100 / 2}
	if got := fight.PlayerFeed[len(fight.PlayerFeed)-1]; got != want {
		t.Fatalf("defense event = %+v, want %+v", got, want)
	}
}

func TestGrinchFeedRecordsWhatPlayersTakeBack(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	fight := Grinch{ID: "f", Phase: GrinchFinalAttack, StartedAt: now, TimedAt: now, HP: 500, FinalRound: 1, FinalBoard: 0b0001,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, DamageDealt: map[string]GrinchDamage{}}
	move := GrinchMove{Phase: GrinchFinalAttack, Member: "alice", At: now, FinalRound: 1, FinalBoard: 0b0001, GiftPoints: 3, Roll: 11}

	move.RequestID, move.Slot = "gift", 0
	fight.Move(rules, move)
	move.RequestID, move.Slot, move.FinalRound = "snowball", GrinchSnowballSlot, 2
	fight.Move(rules, move)
	move.RequestID, move.Slot, move.FinalRound = "hit", 1, 3
	fight.Move(rules, move)

	wantBoss := []GrinchEvent{
		{Kind: EventGift, Member: "alice", Amount: 3, Roll: 11},
		{Kind: EventSnowball, Member: "alice", Amount: int(rules.FinalSnowballReward), Roll: 11},
	}
	if !slices.Equal(fight.BossFeed, wantBoss) {
		t.Fatalf("boss feed = %+v, want %+v", fight.BossFeed, wantBoss)
	}
	if want := []GrinchEvent{{Kind: EventHit, Member: "alice", Damage: float64(rules.FinalHitDamage) / 100}}; !slices.Equal(fight.PlayerFeed, want) {
		t.Fatalf("player feed = %+v, want %+v", fight.PlayerFeed, want)
	}
}

func TestGrinchFeedMarksTheStartAndTheDefeat(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	// A third press starts the fight at once.
	early := Grinch{ID: "f", Phase: GrinchLobby, StartedAt: now, HP: GrinchBossHP, Joined: map[string]bool{"a": true, "b": true}, Accepted: map[string]bool{}}
	early.Move(rules, GrinchMove{Phase: GrinchLobby, Member: "c", RequestID: "c", At: now, Roll: 3})
	if want := []GrinchEvent{{Kind: EventPhase, Phase: GrinchSnowballs, Roll: 3}}; !slices.Equal(early.BossFeed, want) {
		t.Fatalf("boss feed after the early start = %+v", early.BossFeed)
	}
	// The countdown ending starts it too, with a line chosen from the time because nothing random is given.
	counted := Grinch{ID: "f", Phase: GrinchLobby, StartedAt: now, CountdownAt: now, HP: GrinchBossHP, Joined: map[string]bool{"a": true}}
	deadline := now.Add(rules.RegistrationDuration)
	counted.Advance(rules, deadline)
	if want := []GrinchEvent{{Kind: EventPhase, Phase: GrinchSnowballs, Roll: int(deadline.UnixMilli())}}; !slices.Equal(counted.BossFeed, want) {
		t.Fatalf("boss feed after the countdown = %+v", counted.BossFeed)
	}
	// The last player at zero HP ends the fight and the Grinch gloats.
	defeat := Grinch{ID: "f", Phase: GrinchDefense, StartedAt: now, TimedAt: now, HP: rules.SnowballsBossHP, Board: 0b1, DefenseRound: 1, DefensePlayers: 1,
		Joined: map[string]bool{"a": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{"a": 100 - rules.DefenseMissHPLoss}, DamageDealt: map[string]GrinchDamage{}}
	defeat.Move(rules, GrinchMove{Phase: GrinchDefense, Slot: 1, Member: "a", RequestID: "x", At: now, DefenseRound: 1, NextBoard: 2, Roll: 5})
	last := defeat.BossFeed[len(defeat.BossFeed)-1]
	if defeat.Phase != GrinchDefeat || last.Kind != EventPhase || last.Phase != GrinchDefeat || last.Roll != 5 {
		t.Fatalf("defeat = %s, boss feed %+v", defeat.Phase, defeat.BossFeed)
	}
}

func TestGrinchFeedKeepsTheLatestEventsWithoutChangingEarlierCopies(t *testing.T) {
	var feed []GrinchEvent
	var copies [][]GrinchEvent
	for i := range GrinchFeedSize + 3 {
		feed = addEvent(feed, GrinchEvent{Kind: EventHit, Roll: i})
		copies = append(copies, feed)
	}
	if len(feed) != GrinchFeedSize || feed[0].Roll != 3 || feed[GrinchFeedSize-1].Roll != GrinchFeedSize+2 {
		t.Fatalf("feed = %+v, want the last %d events", feed, GrinchFeedSize)
	}
	if len(copies[GrinchFeedSize-1]) != GrinchFeedSize || copies[GrinchFeedSize-1][0].Roll != 0 {
		t.Fatalf("an earlier feed changed: %+v", copies[GrinchFeedSize-1])
	}
}
