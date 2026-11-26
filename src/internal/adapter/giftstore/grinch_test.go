package giftstore_test

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type grinchDisplayFixture struct {
	created     int
	updated     int
	now         time.Time
	sendError   error
	updateError error
}

func (d *grinchDisplayFixture) Announce(_ context.Context, _ core.Grinch) (string, time.Time, error) {
	d.created++
	return "message", d.now, d.sendError
}

func (d *grinchDisplayFixture) Update(_ context.Context, _ core.Grinch) error {
	d.updated++
	return d.updateError
}

func TestGrinchSQLite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "grinch.sqlite")
	open := func() *giftstore.SQLite {
		t.Helper()
		store, err := giftstore.OpenSQLite(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { store.Close() })
		return store
	}
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	newEncounter := func(store *giftstore.SQLite) *app.GrinchEncounter {
		return app.NewGrinchEncounter(store, "guild", "guild:channel", classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 })
	}
	store, other := open(), open()
	first, second := newEncounter(store), newEncounter(other)
	display := &grinchDisplayFixture{now: now}
	// The test display is shared, but only the transaction winner can call Announce.
	var group sync.WaitGroup
	for i, encounter := range []*app.GrinchEncounter{first, second} {
		group.Go(func() {
			if _, err := encounter.Spawn(t.Context(), now, i == 0, display); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	if display.created != 1 {
		t.Fatalf("announcements = %d", display.created)
	}
	fight, err := first.Sync(t.Context(), display)
	if err != nil || fight.Phase != core.GrinchLobby {
		t.Fatalf("spawned state = %+v, %v", fight, err)
	}
	for i := range 3 {
		_, _, err := first.Play(t.Context(), fight.ID, fight.MessageID, core.GrinchMove{
			Phase: core.GrinchLobby, At: now, Member: strconv.Itoa(i), RequestID: "join" + strconv.Itoa(i),
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	now = now.Add(core.GrinchJoinDuration)
	if _, err := first.Sync(t.Context(), display); err != nil {
		t.Fatal(err)
	}
	for i := range 30 {
		group.Go(func() {
			encounter := first
			if i%2 == 0 {
				encounter = second
			}
			_, _, err := encounter.Play(t.Context(), fight.ID, fight.MessageID, core.GrinchMove{
				Phase: core.GrinchSnowballs, At: now.Add(time.Second), Member: strconv.Itoa(i % 3), RequestID: strconv.Itoa(i),
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	display.updateError = errors.New("edit failed")
	state, err := first.Sync(t.Context(), display)
	if !errors.Is(err, display.updateError) || state.Phase != core.GrinchDefense || state.Health(core.DefaultTuning().Grinch) != 75 || len(state.Accepted) != 9 { // Three joins and the six snowballs that end the phase.
		t.Fatalf("state after concurrent attacks = %+v, %v", state, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	display.updateError = nil
	restarted := newEncounter(open())
	recovered, err := restarted.Sync(t.Context(), display)
	if err != nil || recovered.ID != fight.ID || recovered.MessageID != fight.MessageID || recovered.Health(core.DefaultTuning().Grinch) != 75 || display.updated != 4 {
		t.Fatalf("recovered state = %+v, %v, %d edits", recovered, err, display.updated)
	}
	var total float64
	for _, score := range recovered.Leaderboard(core.DefaultTuning().Grinch) {
		total += score.Points
	}
	if len(recovered.Leaderboard(core.DefaultTuning().Grinch)) != 3 || total != 25 {
		t.Fatalf("leaderboard after reopening SQLite = %+v", recovered.Leaderboard(core.DefaultTuning().Grinch))
	}
	for i := range 5 {
		group.Go(func() {
			_, _, err := restarted.Play(t.Context(), fight.ID, fight.MessageID, core.GrinchMove{
				Phase: core.GrinchDefense, Slot: 0, DefenseRound: recovered.DefenseRound, At: now.Add(2 * time.Second), Member: "1", RequestID: "defense" + strconv.Itoa(i),
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	state, err = restarted.Sync(t.Context(), display)
	if err != nil || state.Phase != core.GrinchDefense || state.Health(core.DefaultTuning().Grinch) != 65 || state.Defenses != 1 {
		t.Fatalf("state after concurrent defenses = %+v, %v", state, err)
	}
}

func TestGrinchUncertainAnnouncement(t *testing.T) {
	store, err := giftstore.OpenSQLite(t.Context(), filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	encounter := app.NewGrinchEncounter(store, "guild", "guild:channel", classicFight(core.DefaultTuning().Grinch), func() time.Time { return now }, func(int) int { return 0 })
	display := &grinchDisplayFixture{now: now, sendError: errors.New("send uncertain")}
	if ok, err := encounter.Spawn(t.Context(), now, false, display); ok || !errors.Is(err, display.sendError) {
		t.Fatalf("first spawn = %v, %v", ok, err)
	}
	if ok, err := encounter.Spawn(t.Context(), now, true, display); ok || err != nil || display.created != 1 {
		t.Fatalf("second spawn = %v, %v; %d announcements", ok, err, display.created)
	}
	now = now.Add(time.Minute)
	if _, err := encounter.Sync(t.Context(), display); err != nil {
		t.Fatal(err)
	}
	display.sendError, display.now = nil, now
	if ok, err := encounter.Spawn(t.Context(), now, true, display); !ok || err != nil || display.created != 2 {
		t.Fatalf("later spawn = %v, %v; %d announcements", ok, err, display.created)
	}
}

// classicFight keeps the numbers these tests were written for: fixed damage, no wait between
// clicks, no counterattack, no last chance, and a one-minute fight clock.
func classicFight(rules core.GrinchRules) core.GrinchRules {
	rules.HitsPerPlayer, rules.MinHits, rules.ClickCooldown, rules.CounterattackPercent = 0, 0, 0, 0
	rules.LastChance = 0
	rules.FightDuration = time.Minute
	return rules
}
