package app_test

import (
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// The shell supplies the random number the core uses to pick the Grinch's lines.
func TestPlayGivesTheFightARandomRollForTheGrinchsLines(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	rules := core.DefaultTuning().Grinch
	store := &giftfixture.Store{}
	fight := core.Grinch{ID: "f", MessageID: "m", Phase: core.GrinchSnowballs, StartedAt: now, HP: rules.SnowballsBossHP + rules.SnowballDamage,
		Joined: map[string]bool{"alice": true}, PlayerDamage: map[string]int{},
		DamageDealt: map[string]core.GrinchDamage{}, Accepted: map[string]bool{}}
	if saved, err := store.SaveGrinch(t.Context(), "scope", 0, fight, "server", nil); !saved || err != nil {
		t.Fatalf("seed: %v, %v", saved, err)
	}
	var bounds []int
	encounter := app.NewGrinchEncounter(store, "server", "scope", rules, func() time.Time { return now }, func(bound int) int {
		bounds = append(bounds, bound)
		return bound - 1
	})
	state, _, err := encounter.Play(t.Context(), "f", "m", core.GrinchMove{Phase: core.GrinchSnowballs, Member: "alice", RequestID: "1", At: now})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.BossFeed) != 1 || state.BossFeed[0].Kind != core.EventPhase || state.BossFeed[0].Roll != 999_999 {
		t.Fatalf("boss feed = %+v, want one line for the defense phase with the shell's roll", state.BossFeed)
	}
	if len(state.PlayerFeed) != 1 || state.PlayerFeed[0].Member != "alice" {
		t.Fatalf("player feed = %+v", state.PlayerFeed)
	}
	// The feeds are stored with the fight.
	stored, _, err := store.LoadGrinch(t.Context(), "scope")
	if err != nil || len(stored.BossFeed) != 1 || len(stored.PlayerFeed) != 1 {
		t.Fatalf("stored fight = %+v, %v", stored, err)
	}
}
