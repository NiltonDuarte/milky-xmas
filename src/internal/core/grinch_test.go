package core

import (
	"strconv"
	"testing"
	"time"
)

func TestGrinchCombat(t *testing.T) {
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	for _, players := range []int{1, 3, 6, 7} {
		t.Run(strconv.Itoa(players), func(t *testing.T) {
			fight := Grinch{ID: "fight", Phase: GrinchSnowballs, HP: 10000, StartedAt: start,
				Joined: map[string]bool{}, PlayerDamage: map[string]int{}, Accepted: map[string]bool{},
			}
			for player := range players {
				fight.Joined[strconv.Itoa(player)] = true
			}
			sequence := 0
			hit := func(phase GrinchPhase, slot, member int, at time.Time) {
				sequence++
				board := uint8(1)
				if fight.Board == board {
					board = 5
				}
				fight.Move(GrinchMove{Phase: phase, Slot: slot, Member: strconv.Itoa(member),
					RequestID: strconv.Itoa(sequence), At: at, NextBoard: board, FinalRound: fight.FinalRound, DefenseRound: fight.DefenseRound})
			}
			for range 25 {
				hit(GrinchSnowballs, 0, 0, start.Add(9*time.Second))
			}
			hit(GrinchSnowballs, 0, 0, start.Add(9*time.Second))
			if fight.Phase != GrinchDefense || fight.Health() != 75 || fight.DefensePlayers != players {
				t.Fatalf("defense started at %+v", fight)
			}
			hit(GrinchDefense, 1, 0, start.Add(12*time.Second))
			if fight.PlayerDamage["0"] != 10 || fight.Health() != 75 {
				t.Fatalf("snowflake result: %+v", fight)
			}
			for i := range players {
				hit(GrinchDefense, 0, i, start.Add(15*time.Second))
				if i == 0 && fight.Health() != 75-30/float64(players) {
					t.Errorf("scaled damage left %v HP", fight.Health())
				}
			}
			if fight.Phase != GrinchAttack || fight.Health() != 45 {
				t.Fatalf("attack phase started at %+v", fight)
			}
			hit(GrinchAttack, 1, 0, start.Add(20*time.Second))
			hit(GrinchAttack, 0, 0, start.Add(20*time.Second))
			if fight.HP != 3475 {
				t.Fatalf("phase-three damage left %d hundredths HP", fight.HP)
			}
			for range 8 {
				hit(GrinchAttack, 0, 0, start.Add(20*time.Second))
			}
			if fight.Phase != GrinchFinalAttack || fight.HP != 500 {
				t.Fatalf("final stage started at %+v", fight)
			}
			for range 5 {
				hit(GrinchFinalAttack, 0, 0, start.Add(59*time.Second))
			}
			if fight.Phase != GrinchWon || fight.HP != 0 {
				t.Fatalf("fight outcome = %+v", fight)
			}
		})
	}
}

func TestGrinchDeadlines(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	var fight Grinch
	if fight.CanSpawn(now, time.Time{}, 0, false) || fight.CanSpawn(now, now, 10, false) ||
		!fight.CanSpawn(now, now.Add(-10*time.Minute), 9, false) {
		t.Fatal("unexpected activity or roll eligibility")
	}
	fight = Grinch{ID: "fight", Phase: GrinchSnowballs, StartedAt: now, HP: 10000}
	if fight.CanSpawn(now, now, 0, true) || fight.Advance(now.Add(2*time.Minute)) || fight.Phase != GrinchSnowballs {
		t.Fatalf("untimed snowball phase left %+v", fight)
	}
	threshold := now.Add(3 * time.Minute)
	fight.HP = 7600
	fight.Joined = map[string]bool{"alice": true}
	fight.Accepted = map[string]bool{}
	fight.Move(GrinchMove{Phase: GrinchSnowballs, At: threshold, Member: "alice", RequestID: "threshold"})
	end := threshold.Add(time.Minute)
	if fight.Phase != GrinchDefense || fight.Deadline() != end || fight.Advance(end.Add(-time.Nanosecond)) {
		t.Fatalf("countdown did not begin at 75%%: %+v", fight)
	}
	if !fight.Advance(end) || fight.Phase != GrinchEscaped || fight.EndedAt != end {
		t.Fatalf("encounter deadline left %+v", fight)
	}
	if fight.CanSpawn(end, end, 0, false) || !fight.CanSpawn(end, time.Time{}, 99, true) ||
		!fight.CanSpawn(end.Add(30*time.Minute), end.Add(30*time.Minute), 0, false) {
		t.Fatal("unexpected spawn cooldown or moderator override")
	}
}

func TestGrinchRegistration(t *testing.T) {
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{
		ID: "fight", Phase: GrinchLobby, HP: 10000, StartedAt: start,
		Joined: map[string]bool{}, Participants: map[string]bool{}, Accepted: map[string]bool{},
	}
	deadline := start.Add(time.Minute)
	if fight.Advance(deadline.Add(-time.Nanosecond)) {
		t.Fatal("registration closed early")
	}
	join := GrinchMove{Phase: GrinchLobby, Member: "alice", RequestID: "join", At: deadline.Add(-time.Nanosecond)}
	fight.Move(join)
	join.RequestID = "repeat"
	fight.Move(join)
	if len(fight.Joined) != 1 || len(fight.Accepted) != 1 {
		t.Fatal("registration was counted more than once")
	}
	join.Member, join.RequestID, join.At = "bob", "late", deadline
	fight.Move(join)
	if len(fight.Joined) != 1 || !fight.Advance(deadline) || fight.Phase != GrinchSnowballs || fight.StartedAt != deadline {
		t.Fatalf("incorrect registration cutoff: %+v", fight)
	}
	hit := GrinchMove{Phase: GrinchSnowballs, Member: "bob", RequestID: "hit", At: deadline}
	fight.Move(hit)
	if fight.HP != 10000 {
		t.Fatal("unregistered attack was counted")
	}
	hit.Member = "alice"
	fight.Move(hit)
	if fight.HP != 9900 || fight.Advance(deadline.Add(time.Minute)) || fight.Phase != GrinchSnowballs {
		t.Fatalf("unexpected combat progress: %+v", fight)
	}
	empty := Grinch{ID: "empty", Phase: GrinchLobby, StartedAt: start, HP: 10000}
	if !empty.Advance(deadline) || empty.Phase != GrinchCancelled {
		t.Fatalf("empty registration was not cancelled: %+v", empty)
	}
}

func TestGrinchDefeat(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{
		ID: "fight", Phase: GrinchDefense, StartedAt: now, TimedAt: now,
		HP: 7500, Board: 1, Joined: map[string]bool{"alice": true, "bob": true},
		Accepted: map[string]bool{}, PlayerDamage: map[string]int{"alice": 90, "bob": 90},
	}
	for _, member := range []string{"alice", "bob"} {
		fight.Move(GrinchMove{Phase: GrinchDefense, Slot: 1, Member: member, RequestID: member, At: now.Add(time.Second), NextBoard: 5, DefenseRound: fight.DefenseRound})
		if member == "alice" && (fight.Phase != GrinchDefense || !fight.EndedAt.IsZero()) {
			t.Fatal("fight ended with another player alive")
		}
	}
	if fight.Phase != GrinchDefeat || fight.Health() != 75 || fight.EndedAt != now.Add(time.Second) {
		t.Fatalf("fight did not end immediately after the last elimination: %+v", fight)
	}
	solo := Grinch{ID: "solo", Phase: GrinchDefense, StartedAt: now, HP: 7500, Board: 1,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{},
	}
	for i := range 10 {
		board := uint8(1)
		if solo.Board == board {
			board = 5
		}
		solo.Move(GrinchMove{Phase: GrinchDefense, Slot: 1, Member: "alice", RequestID: strconv.Itoa(i), At: now,
			NextBoard: board, DefenseRound: solo.DefenseRound})
	}
	if solo.Phase != GrinchDefeat || solo.PlayerDamage["alice"] != 100 {
		t.Fatalf("solo player was not defeated: %+v", solo)
	}
}

func TestGrinchLeaderboardCountsAppliedDamage(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{Phase: GrinchSnowballs, StartedAt: now, HP: 7600,
		Joined:   map[string]bool{"alice": true, "bob": true, "carol": true},
		Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{},
	}
	hit := func(phase GrinchPhase, slot int, player, id string) {
		board := uint8(1)
		if fight.Board == board {
			board = 5
		}
		fight.Move(GrinchMove{Phase: phase, Slot: slot, Member: player, RequestID: id, At: now, NextBoard: board, FinalRound: fight.FinalRound, DefenseRound: fight.DefenseRound})
	}
	hit(GrinchSnowballs, 0, "alice", "snowball")
	hit(GrinchSnowballs, 0, "alice", "snowball")
	hit(GrinchDefense, 1, "carol", "failed defense")
	hit(GrinchDefense, 0, "bob", "defense 1")
	hit(GrinchDefense, 0, "bob", "defense 2")
	hit(GrinchDefense, 0, "alice", "defense 3")
	// Simulate the last grid attack when only 1 HP remains above the floor.
	fight.HP = 600
	hit(GrinchAttack, 0, "alice", "ice")
	hit(GrinchFinalAttack, 0, "bob", "final")
	want := []GrinchScore{{"bob", 22}, {"alice", 12}, {"carol", 0}}
	got := fight.Leaderboard()
	if len(got) != len(want) {
		t.Fatalf("leaderboard = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("rank %d = %+v, want %+v", i+1, got[i], want[i])
		}
	}
}

func TestGrinchLeaderboardPreservesFractionalDefenseDamage(t *testing.T) {
	fight := Grinch{
		Joined:         map[string]bool{"alice": true, "bob": true},
		DefensePlayers: 7,
		DamageDealt:    map[string]GrinchDamage{"alice": {Defenses: 4}, "bob": {Defenses: 3}},
	}
	scores := fight.Leaderboard()
	if len(scores) != 2 || scores[0].Member != "alice" || scores[0].Points != 120.0/7 || scores[0].Points+scores[1].Points != 30 {
		t.Fatalf("fractional contributions = %+v", scores)
	}
}

func TestGrinchFinalGiftRow(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{Phase: GrinchFinalAttack, StartedAt: now, HP: 500, FinalRound: 1, FinalBoard: 1,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, DamageDealt: map[string]GrinchDamage{},
	}
	move := GrinchMove{Phase: GrinchFinalAttack, Member: "alice", At: now, FinalRound: 1,
		RequestID: "gift", Slot: 0, GiftPoints: 5, FinalBoard: 2}
	fight.Move(move)
	if fight.HP != 500 || fight.Gifts["alice"] != 5 || fight.FinalBoard != 2 || fight.FinalRound != 2 {
		t.Fatalf("gift state = %+v", fight)
	}
	fight.Move(move)
	move.RequestID = "stale"
	fight.Move(move)
	if fight.HP != 500 || fight.Gifts["alice"] != 5 {
		t.Fatalf("duplicate or stale click changed score: %+v", fight)
	}
	move.RequestID, move.FinalRound = "attack", 2
	fight.Move(move)
	if fight.HP != 400 || fight.Leaderboard()[0].Points != 2 || fight.GiftLeaderboard()[0].Points != 5 {
		t.Fatalf("attack state = %+v", fight)
	}
}

func TestGrinchEarlyStartAndLateJoin(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchLobby, HP: 10000, StartedAt: now,
		Joined: map[string]bool{}, Accepted: map[string]bool{},
	}
	join := func(member, request string, at time.Time) {
		fight.Move(GrinchMove{Phase: GrinchLobby, Member: member, RequestID: request, At: at})
	}
	join("alice", "1", now)
	join("bob", "2", now)
	join("alice", "duplicate", now)
	if fight.Phase != GrinchLobby || len(fight.Joined) != 2 {
		t.Fatalf("started before three distinct players: %+v", fight)
	}
	early := now.Add(15 * time.Second)
	join("carol", "3", early)
	if fight.Phase != GrinchSnowballs || fight.StartedAt != early {
		t.Fatalf("third player did not start combat: %+v", fight)
	}
	join("dave", "4", early.Add(time.Second))
	fight.HP = 7600
	fight.Move(GrinchMove{Phase: GrinchSnowballs, Member: "dave", RequestID: "hit", At: early.Add(2 * time.Second)})
	join("eve", "5", early.Add(3*time.Second))
	if fight.Phase != GrinchDefense || fight.DefensePlayers != 4 || len(fight.Joined) != 5 {
		t.Fatalf("late join or defense scaling failed: %+v", fight)
	}
	fight.Advance(fight.Deadline())
	join("frank", "6", early.Add(2*time.Minute))
	if fight.Joined["frank"] {
		t.Fatal("player joined an ended fight")
	}
}

func TestGrinchDefenseBoardRerollsAndRejectsStaleClicks(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchDefense, StartedAt: now, TimedAt: now,
		HP: 7500, Board: 1, DefenseRound: 1, DefensePlayers: 3,
		Joined:   map[string]bool{"alice": true, "bob": true, "carol": true},
		Accepted: map[string]bool{}, PlayerDamage: map[string]int{},
	}
	move := GrinchMove{Phase: GrinchDefense, Member: "alice", Slot: 0, RequestID: "tree",
		At: now, DefenseRound: 1, NextBoard: 2}
	fight.Move(move)
	if fight.Health() != 65 || fight.Board != 2 || fight.DefenseRound != 2 {
		t.Fatalf("tree did not update the board: %+v", fight)
	}
	move.Member, move.RequestID = "bob", "stale"
	fight.Move(move)
	if fight.Health() != 65 || fight.PlayerDamage["bob"] != 0 || fight.DefenseRound != 2 {
		t.Fatalf("stale tree click was applied to the new snowflake: %+v", fight)
	}
	move.RequestID, move.DefenseRound = "snowflake", 2
	fight.Move(move)
	if fight.Health() != 65 || fight.PlayerDamage["bob"] != 10 || fight.DefenseRound != 3 || fight.Board == 2 {
		t.Fatalf("snowflake did not change the board: %+v", fight)
	}
}
