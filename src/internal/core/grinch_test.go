package core

import (
	"strconv"
	"testing"
	"time"
)

func TestGrinchCombat(t *testing.T) {
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	for _, players := range []int{3, 6, 7} {
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
				fight.Move(GrinchMove{Phase: phase, Slot: slot, Member: strconv.Itoa(member),
					RequestID: strconv.Itoa(sequence), At: at, NextBoard: 1})
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
	deadline := start.Add(GrinchJoinDuration)
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
