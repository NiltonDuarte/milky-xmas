package core

import (
	"testing"
	"time"
)

func TestGrinchFinalSnowball(t *testing.T) {
	now := time.Now()
	fight := Grinch{ID: "fight", Phase: GrinchFinalAttack, HP: 500, StartedAt: now, TimedAt: now, FinalRound: 1,
		Joined: map[string]bool{"alice": true}, Accepted: make(map[string]bool)}
	action := GrinchMove{Phase: GrinchFinalAttack, Slot: GrinchSnowballSlot, Member: "alice", At: now, FinalRound: 1, RequestID: "one"}
	if reply := fight.Move(testRules, action); reply != testRules.SnowballEarned() || !fight.Accepted[action.RequestID] || fight.HP != 500 {
		t.Fatalf("initial snowball = %q, %+v", reply, fight)
	}
	if reply := fight.Move(testRules, action); reply == testRules.SnowballEarned() {
		t.Fatal("duplicate action awarded a snowball")
	}
	action.RequestID = "two"
	if reply := fight.Move(testRules, action); reply == testRules.SnowballEarned() || fight.Accepted[action.RequestID] {
		t.Fatal("stale row awarded a snowball")
	}
	action.FinalRound = fight.FinalRound
	if reply := fight.Move(testRules, action); reply != testRules.SnowballEarned() || !fight.Accepted[action.RequestID] || fight.HP != 500 {
		t.Fatalf("second snowball = %q, %+v", reply, fight)
	}
}

func TestGrinchLossRewards(t *testing.T) {
	fight := Grinch{Joined: map[string]bool{"a": true, "b": true, "c": true, "d": true},
		PlayerDamage: map[string]int{"d": 100}}
	for _, phase := range []GrinchPhase{GrinchDefeat, GrinchEscaped, GrinchWon, GrinchCancelled} {
		fight.Phase = phase
		rewards := fight.LossRewards(testRules)
		if phase == GrinchWon || phase == GrinchCancelled {
			if len(rewards) != 0 {
				t.Fatalf("%s incorrectly awarded %+v", phase, rewards)
			}
			continue
		}
		if len(rewards) != 4 {
			t.Fatalf("%s roster rewards = %+v", phase, rewards)
		}
		for member := range fight.Joined {
			if rewards[member] != (GrinchReward{Coal: 1, YellowSnowballs: 4}) {
				t.Errorf("%s reward for %s: %+v", phase, member, rewards[member])
			}
		}
	}
}
