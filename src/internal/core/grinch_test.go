package core

import (
	"maps"
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
				fight.Move(testRules, GrinchMove{Phase: phase, Slot: slot, Member: strconv.Itoa(member),
					RequestID: strconv.Itoa(sequence), At: at, NextBoard: board, FinalRound: fight.FinalRound, DefenseRound: fight.DefenseRound})
			}
			for range 25 {
				hit(GrinchSnowballs, 0, 0, start.Add(9*time.Second))
			}
			hit(GrinchSnowballs, 0, 0, start.Add(9*time.Second))
			if fight.Phase != GrinchDefense || fight.Health(testRules) != 75 || fight.DefensePlayers != min(players, 6) {
				t.Fatalf("defense started at %+v", fight)
			}
			hit(GrinchDefense, 1, 0, start.Add(12*time.Second))
			if fight.PlayerDamage["0"] != 20 || fight.Health(testRules) != 75 {
				t.Fatalf("snowflake result: %+v", fight)
			}
			// One safe defense per player, but never more than six.
			for i := range min(players, 6) {
				hit(GrinchDefense, 0, i, start.Add(15*time.Second))
				if i == 0 && fight.Health(testRules) != 75-30/float64(min(players, 6)) {
					t.Errorf("scaled damage left %v HP", fight.Health(testRules))
				}
			}
			if fight.Phase != GrinchAttack || fight.Health(testRules) != 45 {
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
	if fight.CanSpawn(testRules, now, time.Time{}, 0, false) || fight.CanSpawn(testRules, now, now, 10, false) ||
		!fight.CanSpawn(testRules, now, now.Add(-10*time.Minute), 9, false) {
		t.Fatal("unexpected activity or roll eligibility")
	}
	fight = Grinch{ID: "fight", Phase: GrinchSnowballs, StartedAt: now, HP: 10000}
	if fight.CanSpawn(testRules, now, now, 0, true) || !fight.Deadline(testRules).Equal(now.Add(2*time.Minute)) ||
		fight.Advance(testRules, now.Add(2*time.Minute-time.Nanosecond)) || fight.Phase != GrinchSnowballs {
		t.Fatalf("the snowball phase did not last two minutes: %+v", fight)
	}
	threshold := now.Add(110 * time.Second)
	fight.HP = 7600
	fight.Joined = map[string]bool{"alice": true}
	fight.Accepted = map[string]bool{}
	fight.Move(testRules, GrinchMove{Phase: GrinchSnowballs, At: threshold, Member: "alice", RequestID: "threshold"})
	end := threshold.Add(time.Minute)
	if fight.Phase != GrinchDefense || fight.Deadline(testRules) != end || fight.Advance(testRules, end.Add(-time.Nanosecond)) {
		t.Fatalf("countdown did not begin at 75%%: %+v", fight)
	}
	if !fight.Advance(testRules, end) || fight.Phase != GrinchEscaped || fight.EndedAt != end {
		t.Fatalf("encounter deadline left %+v", fight)
	}
	if fight.CanSpawn(testRules, end, end, 0, false) || !fight.CanSpawn(testRules, end, time.Time{}, 99, true) ||
		!fight.CanSpawn(testRules, end.Add(30*time.Minute), end.Add(30*time.Minute), 0, false) {
		t.Fatal("unexpected spawn cooldown or moderator override")
	}
}

func TestGrinchRegistration(t *testing.T) {
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{
		ID: "fight", Phase: GrinchLobby, HP: 10000, StartedAt: start,
		Joined: map[string]bool{}, Accepted: map[string]bool{},
	}
	// The countdown starts at the first Start press, not at the announcement.
	pressed := start.Add(30 * time.Second)
	deadline := pressed.Add(time.Minute)
	if fight.Advance(testRules, start.Add(59*time.Second)) || !fight.CountdownAt.IsZero() {
		t.Fatal("the fight changed before anyone pressed Start")
	}
	join := GrinchMove{Phase: GrinchLobby, Member: "alice", RequestID: "join", At: pressed}
	fight.Move(testRules, join)
	if !fight.CountdownAt.Equal(pressed) || !fight.LobbyDeadline(testRules).Equal(deadline) {
		t.Fatalf("countdown = %v, deadline = %v, want it to start at the press and end at %v", fight.CountdownAt, fight.LobbyDeadline(testRules), deadline)
	}
	if fight.Advance(testRules, deadline.Add(-time.Nanosecond)) {
		t.Fatal("registration closed early")
	}
	join.RequestID, join.At = "repeat", deadline.Add(-time.Nanosecond)
	fight.Move(testRules, join)
	if len(fight.Joined) != 1 || len(fight.Accepted) != 1 || !fight.CountdownAt.Equal(pressed) {
		t.Fatal("registration was counted more than once or restarted the countdown")
	}
	join.Member, join.RequestID, join.At = "bob", "late", deadline
	fight.Move(testRules, join)
	if len(fight.Joined) != 1 || !fight.Advance(testRules, deadline) || fight.Phase != GrinchSnowballs || fight.StartedAt != deadline {
		t.Fatalf("incorrect registration cutoff: %+v", fight)
	}
	// Anyone can take part once the fight has started: a player who did not press Start attacks and is in.
	hit := GrinchMove{Phase: GrinchSnowballs, Member: "bob", RequestID: "hit", At: deadline}
	fight.Move(testRules, hit)
	if fight.HP != 9900 || !fight.Joined["bob"] {
		t.Fatalf("an attack by a player who did not press Start: HP %d, players %v", fight.HP, fight.Joined)
	}
	hit.Member, hit.RequestID = "alice", "hit2"
	fight.Move(testRules, hit)
	if fight.HP != 9800 || fight.Advance(testRules, deadline.Add(time.Minute)) || fight.Phase != GrinchSnowballs {
		t.Fatalf("unexpected combat progress: %+v", fight)
	}
}

func TestGrinchWaitsForTheFirstStartPress(t *testing.T) {
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	rules.StartWindow = 5 * time.Minute
	empty := Grinch{ID: "empty", Phase: GrinchLobby, StartedAt: start, HP: 10000}
	if !empty.LobbyDeadline(rules).Equal(start.Add(5*time.Minute)) || empty.Advance(rules, start.Add(5*time.Minute-time.Nanosecond)) {
		t.Fatal("the fight closed before the start window passed")
	}
	if !empty.Advance(rules, start.Add(5*time.Minute)) || empty.Phase != GrinchCancelled || !empty.EndedAt.Equal(start.Add(5*time.Minute)) {
		t.Fatalf("a fight nobody started was not cancelled at the end of the window: %+v", empty)
	}
	// The window does not apply once someone pressed Start.
	pressed := Grinch{ID: "p", Phase: GrinchLobby, StartedAt: start, HP: 10000, Joined: map[string]bool{}, Accepted: map[string]bool{}}
	pressed.Move(rules, GrinchMove{Phase: GrinchLobby, Member: "alice", RequestID: "1", At: start.Add(4 * time.Minute)})
	if got := pressed.LobbyDeadline(rules); !got.Equal(start.Add(5 * time.Minute)) {
		t.Fatalf("deadline after a press = %v, want a minute after it", got)
	}
}

func TestGrinchStartsAtOnceWhenEnoughPlayersPressStart(t *testing.T) {
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchLobby, HP: 10000, StartedAt: start, Joined: map[string]bool{}, Accepted: map[string]bool{}}
	for i, member := range []string{"alice", "bob", "carol"} {
		at := start.Add(time.Duration(i+1) * 5 * time.Second)
		fight.Move(testRules, GrinchMove{Phase: GrinchLobby, Member: member, RequestID: member, At: at})
		if want := i < 2; (fight.Phase == GrinchLobby) != want {
			t.Fatalf("after %d presses the phase is %s", i+1, fight.Phase)
		}
	}
	if fight.Phase != GrinchSnowballs || !fight.StartedAt.Equal(start.Add(15*time.Second)) {
		t.Fatalf("the third press did not start the fight at once: %+v", fight)
	}
}

func TestGrinchDefeat(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{
		ID: "fight", Phase: GrinchDefense, StartedAt: now, TimedAt: now,
		HP: 7500, Board: 1, Joined: map[string]bool{"alice": true, "bob": true},
		Accepted: map[string]bool{}, PlayerDamage: map[string]int{"alice": 80, "bob": 80},
	}
	for _, member := range []string{"alice", "bob"} {
		fight.Move(testRules, GrinchMove{Phase: GrinchDefense, Slot: 1, Member: member, RequestID: member, At: now.Add(time.Second), NextBoard: 5, DefenseRound: fight.DefenseRound})
		if member == "alice" && (fight.Phase != GrinchDefense || !fight.EndedAt.IsZero()) {
			t.Fatal("fight ended with another player alive")
		}
	}
	if fight.Phase != GrinchDefeat || fight.Health(testRules) != 75 || fight.EndedAt != now.Add(time.Second) {
		t.Fatalf("fight did not end immediately after the last elimination: %+v", fight)
	}
	solo := Grinch{ID: "solo", Phase: GrinchDefense, StartedAt: now, TimedAt: now, DefensePlayers: 1, HP: 7500, Board: 1,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{},
	}
	for i := range 10 {
		board := uint8(1)
		if solo.Board == board {
			board = 5
		}
		solo.Move(testRules, GrinchMove{Phase: GrinchDefense, Slot: 1, Member: "alice", RequestID: strconv.Itoa(i), At: now,
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
		fight.Move(testRules, GrinchMove{Phase: phase, Slot: slot, Member: player, RequestID: id, At: now, NextBoard: board, FinalRound: fight.FinalRound, DefenseRound: fight.DefenseRound})
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
	want := []GrinchScore{{"bob", 21}, {"alice", 12}, {"carol", 0}}
	got := fight.Leaderboard(testRules)
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
	scores := fight.Leaderboard(testRules)
	if len(scores) != 2 || scores[0].Member != "alice" || scores[0].Points != 120.0/7 || scores[0].Points+scores[1].Points != 30 {
		t.Fatalf("fractional contributions = %+v", scores)
	}
}

func TestGrinchFinalGiftRow(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{Phase: GrinchFinalAttack, StartedAt: now, TimedAt: now, HP: 500, FinalRound: 1, FinalBoard: 1,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, DamageDealt: map[string]GrinchDamage{},
	}
	move := GrinchMove{Phase: GrinchFinalAttack, Member: "alice", At: now, FinalRound: 1,
		RequestID: "gift", Slot: 0, GiftPoints: 5, FinalBoard: 2}
	fight.Move(testRules, move)
	if fight.HP != 500 || fight.Gifts["alice"] != 5 || fight.FinalBoard != 2 || fight.FinalRound != 2 {
		t.Fatalf("gift state = %+v", fight)
	}
	fight.Move(testRules, move)
	move.RequestID = "stale"
	fight.Move(testRules, move)
	if fight.HP != 500 || fight.Gifts["alice"] != 5 {
		t.Fatalf("duplicate or stale click changed score: %+v", fight)
	}
	move.RequestID, move.FinalRound = "attack", 2
	fight.Move(testRules, move)
	if fight.HP != 400 || fight.Leaderboard(testRules)[0].Points != 1 || fight.GiftLeaderboard()[0].Points != 5 {
		t.Fatalf("attack state = %+v", fight)
	}
}

func TestGrinchEarlyStartAndLateJoin(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchLobby, HP: 10000, StartedAt: now,
		Joined: map[string]bool{}, Accepted: map[string]bool{},
	}
	join := func(member, request string, at time.Time) {
		fight.Move(testRules, GrinchMove{Phase: GrinchLobby, Member: member, RequestID: request, At: at})
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
	fight.Move(testRules, GrinchMove{Phase: GrinchSnowballs, Member: "dave", RequestID: "hit", At: early.Add(2 * time.Second)})
	join("eve", "5", early.Add(3*time.Second))
	if fight.Phase != GrinchDefense || fight.DefensePlayers != 4 || len(fight.Joined) != 5 {
		t.Fatalf("late join or defense scaling failed: %+v", fight)
	}
	fight.Advance(testRules, fight.Deadline(testRules))
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
	fight.Move(testRules, move)
	if fight.Health(testRules) != 65 || fight.Board != 2 || fight.DefenseRound != 2 {
		t.Fatalf("tree did not update the board: %+v", fight)
	}
	move.Member, move.RequestID = "bob", "stale"
	fight.Move(testRules, move)
	if fight.Health(testRules) != 65 || fight.PlayerDamage["bob"] != 0 || fight.DefenseRound != 2 {
		t.Fatalf("stale tree click was applied to the new snowflake: %+v", fight)
	}
	move.RequestID, move.DefenseRound = "snowflake", 2
	fight.Move(testRules, move)
	if fight.Health(testRules) != 65 || fight.PlayerDamage["bob"] != 20 || fight.DefenseRound != 3 || fight.Board == 2 {
		t.Fatalf("snowflake did not change the board: %+v", fight)
	}
}

// testRules keeps small hits, so the mechanics tests walk through each phase step by step. The
// shipped damage is covered by TestGrinchDefaultsNeedAtMostSixHitsPerPhase.
var testRules = stepRules()

func stepRules() GrinchRules {
	rules := classicFight(DefaultTuning().Grinch)
	rules.SnowballDamage, rules.AttackDamage, rules.AttackHeavyDamage, rules.FinalHitDamage = 100, 500, 525, 100
	return rules
}

func TestGrinchCountsCollectedSnowballs(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	rules.FinalSnowballReward = 2
	fight := Grinch{Phase: GrinchFinalAttack, StartedAt: now, TimedAt: now, HP: 500, FinalRound: 0,
		Joined:   map[string]bool{"alice": true, "bob": true, "carol": true},
		Accepted: map[string]bool{}, DamageDealt: map[string]GrinchDamage{},
	}
	pick := func(member, request string) {
		fight.Move(rules, GrinchMove{Phase: GrinchFinalAttack, Member: member, At: now, FinalRound: fight.FinalRound,
			RequestID: request, Slot: GrinchSnowballSlot, FinalBoard: 1})
	}
	pick("bob", "1")
	pick("bob", "1") // The same click again is not counted.
	pick("alice", "2")
	pick("bob", "3")
	got := fight.SnowballLeaderboard()
	if len(got) != 3 || got[0] != (GrinchScore{"bob", 4}) || got[1] != (GrinchScore{"alice", 2}) || got[2] != (GrinchScore{"carol", 0}) {
		t.Fatalf("snowball leaderboard = %+v, want bob 4, alice 2, carol 0", got)
	}
	if (Grinch{Joined: map[string]bool{"a": true}}).SnowballLeaderboard() != nil {
		t.Error("an old fight without scores has a leaderboard")
	}
}

func TestWithPointsKeepsOnlyScoringPlayersInOrder(t *testing.T) {
	got := WithPoints([]GrinchScore{{"a", 3}, {"b", 0}, {"c", 0.5}, {"d", 0}})
	if len(got) != 2 || got[0].Member != "a" || got[1].Member != "c" {
		t.Fatalf("WithPoints = %+v, want a then c", got)
	}
	if got := WithPoints(nil); got == nil || len(got) != 0 {
		t.Fatalf("WithPoints(nil) = %#v, want an empty list", got)
	}
}

func TestGrinchSnowballTimerEndsTheFightWhenTimeRunsOut(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	fight := Grinch{ID: "fight", Phase: GrinchSnowballs, StartedAt: now, HP: testRules.SnowballsBossHP + 100,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}}
	late := now.Add(testRules.SnowballsDuration)
	if reply := fight.Move(testRules, GrinchMove{Phase: GrinchSnowballs, Member: "alice", RequestID: "late", At: late}); reply != "This part of the fight has ended." || fight.HP != testRules.SnowballsBossHP+100 {
		t.Fatalf("a click after the timer = %q, HP %d, want it refused", reply, fight.HP)
	}
	if !fight.Advance(testRules, late) || fight.Phase != GrinchEscaped || !fight.EndedAt.Equal(late) {
		t.Fatalf("the Grinch did not win when the timer ran out: %+v", fight)
	}
	// Reaching the milestone in time moves on, and the next timer starts from that click.
	ontime := Grinch{ID: "fight", Phase: GrinchSnowballs, StartedAt: now, HP: testRules.SnowballsBossHP + 100,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}}
	click := now.Add(testRules.SnowballsDuration - time.Second)
	ontime.Move(testRules, GrinchMove{Phase: GrinchSnowballs, Member: "alice", RequestID: "ontime", At: click})
	if ontime.Phase != GrinchDefense || !ontime.Deadline(testRules).Equal(click.Add(testRules.FightDuration)) {
		t.Fatalf("the milestone did not start the next timer: %+v", ontime)
	}
	// A longer timer can be tuned.
	longer := testRules
	longer.SnowballsDuration = 5 * time.Minute
	slow := Grinch{ID: "fight", Phase: GrinchSnowballs, StartedAt: now, HP: 10000}
	if slow.Advance(longer, now.Add(4*time.Minute)) || !slow.Advance(longer, now.Add(5*time.Minute)) || slow.Phase != GrinchEscaped {
		t.Fatalf("tuned timer = %+v", slow)
	}
}

func TestGrinchAnyoneCanTakePartAfterTheFightStarted(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	fight := Grinch{ID: "f", Phase: GrinchSnowballs, StartedAt: now, HP: rules.SnowballsBossHP + 2*rules.SnowballDamage,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}}
	hit := func(member, request string, phase GrinchPhase, at time.Time) string {
		return fight.Move(rules, GrinchMove{Phase: phase, Member: member, RequestID: request, At: at, NextBoard: 0b0101, DefenseRound: fight.DefenseRound})
	}

	// A player who never pressed Start attacks and is in the fight.
	hit("bob", "1", GrinchSnowballs, now)
	if !fight.Joined["bob"] || fight.HP != rules.SnowballsBossHP+rules.SnowballDamage {
		t.Fatalf("the attack of a new player: HP %d, players %v", fight.HP, fight.Joined)
	}
	// They count when the defense phase is sized, and their actions are scored.
	hit("carol", "2", GrinchSnowballs, now)
	if fight.Phase != GrinchDefense || fight.DefensePlayers != 3 || fight.DamageDealt["carol"].Hundredths != rules.SnowballDamage {
		t.Fatalf("defense = %s with %d players, carol dealt %+v", fight.Phase, fight.DefensePlayers, fight.DamageDealt["carol"])
	}
	// A player who arrives during the defense phase takes part too and can be hit.
	hit("dave", "3", GrinchDefense, now) // Slot 0 of the board is safe.
	if !fight.Joined["dave"] || fight.Defenses != 1 {
		t.Fatalf("a player who arrived in the defense phase: players %v, defenses %d", fight.Joined, fight.Defenses)
	}
	// An action that is refused does not put anyone in the fight.
	fight.Move(rules, GrinchMove{Phase: GrinchDefense, Member: "erin", RequestID: "4", At: now, DefenseRound: fight.DefenseRound + 5})
	fight.Move(rules, GrinchMove{Phase: GrinchSnowballs, Member: "erin", RequestID: "5", At: now})
	if fight.Joined["erin"] || fight.Accepted["4"] || fight.Accepted["5"] {
		t.Fatalf("a refused action registered the player: %v", fight.Joined)
	}
	// Everyone who took part gets the loss reward.
	fight.Phase, fight.EndedAt = GrinchEscaped, now
	rewards := fight.LossRewards(rules)
	if len(rewards) != 4 || rewards["dave"].YellowSnowballs != 4*rules.LossYellowPerGroupMember {
		t.Fatalf("loss rewards = %+v, want the four players who took part", rewards)
	}
}

func TestGrinchDamageLeaderboardShowsDamageAndOnlyTheFinalAttackAwardsPoints(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := testRules
	fight := Grinch{ID: "f", Phase: GrinchSnowballs, StartedAt: now, TimedAt: now, HP: rules.SnowballsBossHP + rules.SnowballDamage,
		Joined: map[string]bool{"alice": true}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]GrinchDamage{}}
	play := func(phase GrinchPhase, request string, slot int) {
		fight.Move(rules, GrinchMove{Phase: phase, Member: "alice", RequestID: request, At: now, Slot: slot, NextBoard: 0b01, FinalBoard: 0b10,
			DefenseRound: fight.DefenseRound, FinalRound: fight.FinalRound})
	}

	// The snowball, defense, and attack phases award no points.
	play(GrinchSnowballs, "snowball", 0)
	play(GrinchDefense, "defense", 0)
	play(GrinchAttack, "attack", 0)
	base := fight.DamageDealt["alice"]
	if base.Points != 0 || base.Hundredths != rules.SnowballDamage+rules.AttackHeavyDamage || base.Defenses != 1 {
		t.Fatalf("after the first phases: %+v, want damage from the snowball, the defense, and the attack, and no points", base)
	}
	// Each final hit removes FinalHitDamage and awards FinalHitScore. The leaderboard counts the damage.
	fight.Phase, fight.HP, fight.FinalRound, fight.FinalBoard = GrinchFinalAttack, 500, 1, 0b10
	for i := range 3 {
		play(GrinchFinalAttack, "final"+string(rune('a'+i)), 0)
	}
	got := fight.DamageDealt["alice"]
	if got.Hundredths != base.Hundredths+3*rules.FinalHitDamage || got.Points != 3*rules.FinalHitScore || rules.FinalHitScore == rules.FinalHitDamage {
		t.Fatalf("after three final hits: %+v, want damage %d and points %d", got, base.Hundredths+3*rules.FinalHitDamage, 3*rules.FinalHitScore)
	}
	wantDamage := float64(got.Hundredths*fight.DefensePlayers+got.Defenses*rules.DefenseTotal()) / (100 * float64(fight.DefensePlayers))
	if board := fight.Leaderboard(rules); len(board) != 1 || board[0].Points != wantDamage {
		t.Fatalf("damage leaderboard = %+v, want only the damage, %v", board, wantDamage)
	}
	// The last hit removes only the HP the Grinch has left, but still awards the points.
	fight.HP = 30
	play(GrinchFinalAttack, "last", 0)
	if fight.HP != 0 || fight.Phase != GrinchWon || fight.DamageDealt["alice"].Hundredths != base.Hundredths+3*rules.FinalHitDamage+30 ||
		fight.DamageDealt["alice"].Points != 4*rules.FinalHitScore {
		t.Fatalf("the killing hit: HP %d, phase %s, %+v", fight.HP, fight.Phase, fight.DamageDealt["alice"])
	}
}

// With the shipped numbers each attack phase takes eight hits per player, and never under fifteen.
func TestGrinchDefaultHitsGrowWithThePlayers(t *testing.T) {
	rules := DefaultTuning().Grinch
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	for _, players := range []int{1, 3, 5} {
		fight := Grinch{ID: "f", Phase: GrinchSnowballs, HP: GrinchBossHP, StartedAt: now,
			Joined: map[string]bool{}, PlayerDamage: map[string]int{}, Accepted: map[string]bool{},
		}
		for player := range players {
			fight.Joined[strconv.Itoa(player)] = true
		}
		hits := map[GrinchPhase]int{}
		for sequence := 0; fight.Phase != GrinchWon && sequence < 400; sequence++ {
			phase := fight.Phase
			move := GrinchMove{Phase: phase, Member: strconv.Itoa(sequence % players), RequestID: strconv.Itoa(sequence), At: now.Add(time.Duration(sequence) * time.Second),
				NextBoard: 5, FinalBoard: 2, FinalRound: fight.FinalRound, DefenseRound: fight.DefenseRound}
			if phase == GrinchDefense {
				for fight.Board&(1<<move.Slot) == 0 {
					move.Slot++
				}
			}
			if phase == GrinchAttack {
				for fight.Board&(1<<move.Slot) != 0 { // The weaker tile.
					move.Slot++
				}
			}
			fight.Move(rules, move)
			if !fight.Accepted[move.RequestID] {
				t.Fatalf("%d players: refused %+v in %+v", players, move, fight)
			}
			hits[phase]++
		}
		needed := map[int]int{1: 15, 3: 24, 5: 40}[players]
		// Rounding the damage up lets five players finish the last 5% one hit early.
		final := map[int]int{1: 15, 3: 24, 5: 39}[players]
		want := map[GrinchPhase]int{GrinchSnowballs: needed, GrinchDefense: max(8, 4*players), GrinchAttack: needed, GrinchFinalAttack: final}
		if fight.Phase != GrinchWon || !maps.Equal(hits, want) {
			t.Errorf("%d players needed %v hits, want %v (phase %s, HP %d)", players, hits, want, fight.Phase, fight.HP)
		}
	}
}

// classicFight keeps the numbers these tests were written for: fixed damage, no wait between
// clicks, no counterattack, no last chance, and a one-minute fight clock.
func classicFight(rules GrinchRules) GrinchRules {
	rules.HitsPerPlayer, rules.MinHits, rules.ClickCooldown, rules.CounterattackPercent = 0, 0, 0, 0
	rules.LastChance = 0
	rules.FightDuration = time.Minute
	rules.DefensePerPlayer, rules.DefenseMinimum, rules.DefenseMaxDefenses = 0, 0, 6
	return rules
}
