package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/grinchtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type grinchFixture struct {
	encounter *app.GrinchEncounter
	now       func() time.Time
	sequence  *atomic.Uint64
}

func newGrinchFixture(now func() time.Time, store app.GrinchStore) grinchFixture {
	return grinchFixture{
		encounter: app.NewGrinchEncounter(store, fixtureServerID, fixtureServerID, now, func(n int) int { return n / 3 }),
		now:       now,
		sequence:  new(atomic.Uint64),
	}
}

func (fixture grinchFixture) Announce(_ context.Context, fight core.Grinch) (string, time.Time, error) {
	return fight.ID, fixture.now(), nil
}

func (fixture grinchFixture) Update(context.Context, core.Grinch) error {
	return nil
}

func (fixture grinchFixture) reply(w http.ResponseWriter, r *http.Request, args []string) {
	ctx := r.Context()
	now := fixture.now()
	state, err := fixture.encounter.Sync(ctx, fixture)
	if err != nil {
		http.Error(w, "Could not load the fight.", http.StatusInternalServerError)
		return
	}
	var result string
	switch {
	case len(args) == 1 && args[0] == "summon":
		var spawned bool
		spawned, err = fixture.encounter.Spawn(ctx, time.Time{}, true, fixture)
		if spawned {
			state, err = fixture.encounter.Sync(ctx, fixture)
		} else if err == nil {
			result = "A Grinch encounter is already in progress."
		}
	case len(args) == 1 && args[0] == "view":
		// The state above already includes deadline transitions.
	case len(args) == 1 && args[0] == "join":
		state, _, err = fixture.encounter.Play(ctx, state.ID, state.MessageID, core.GrinchMove{
			Phase: core.GrinchLobby, At: now, Member: fixtureUserID, RequestID: strconv.FormatUint(fixture.sequence.Add(1), 10),
		})
	case len(args) == 2 && args[0] == "hit":
		slot, parseErr := strconv.Atoi(args[1])
		if parseErr != nil || slot < 1 || slot > 8 {
			http.Error(w, "Choose a tile from 1 to 8.", http.StatusBadRequest)
			return
		}
		state, _, err = fixture.encounter.Play(ctx, state.ID, state.MessageID, core.GrinchMove{
			Phase: state.Phase, Slot: slot - 1, DefenseRound: state.DefenseRound, FinalRound: state.FinalRound, At: now, Member: fixtureUserID,
			RequestID: strconv.FormatUint(fixture.sequence.Add(1), 10),
		})
	default:
		http.Error(w, "Use /grinch summon, /grinch join, /grinch view, or /grinch hit <tile>.", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Could not record the fight action.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if result != "" {
		fmt.Fprintln(w, result)
	}
	fmt.Fprint(w, fixtureSummary(state))
}

func fixtureSummary(state core.Grinch) string {
	if state.ID == "" {
		return "No Grinch encounter. Use /grinch summon."
	}
	var message strings.Builder
	fmt.Fprintf(&message, "%s\n%s", grinchtext.Health(state.Health()), state.Phase)
	switch state.Phase {
	case core.GrinchLobby:
		fmt.Fprintf(&message, "\n🎅 %d joined", len(state.Joined))
		if !state.Joined[fixtureUserID] {
			fmt.Fprint(&message, " · /grinch join")
		}
	case core.GrinchSnowballs:
		fmt.Fprint(&message, "\n/grinch hit 1")
	case core.GrinchFinalAttack:
		fmt.Fprintln(&message)
		for slot := range 4 {
			emoji := "⚔️"
			if state.FinalBoard&(1<<slot) != 0 {
				emoji = "🎁"
			}
			fmt.Fprintf(&message, "%d %s  ", slot+1, emoji)
		}
		fmt.Fprintf(&message, "%d ❄️\n/grinch hit <tile>", core.GrinchSnowballSlot+1)
	case core.GrinchDefense, core.GrinchAttack:
		on, off := "🎄", "❄️"
		if state.Phase == core.GrinchAttack {
			on, off = "🧊", "☃️"
		}
		for slot := range 8 {
			if slot%4 == 0 {
				fmt.Fprintln(&message)
			}
			emoji := off
			if state.Board&(1<<slot) != 0 {
				emoji = on
			}
			fmt.Fprintf(&message, "%d %s  ", slot+1, emoji)
		}
		fmt.Fprint(&message, "\n/grinch hit <tile>")
	}
	if state.EndedAt.IsZero() && state.Phase != core.GrinchLobby && state.Joined != nil {
		fmt.Fprintf(&message, "\n🎅 %d joined", len(state.Joined))
		if !state.Joined[fixtureUserID] {
			fmt.Fprint(&message, " · /grinch join")
		}
	}
	if !state.EndedAt.IsZero() && state.Phase != core.GrinchCancelled {
		if scores := state.Leaderboard(); len(scores) > 0 {
			fmt.Fprint(&message, "\n\nDamage leaderboard")
			for index, score := range scores {
				fmt.Fprintf(&message, "\n%d. %s — %.2f pts", index+1, score.Member, score.Points)
			}
		}
		if scores := state.GiftLeaderboard(); len(scores) > 0 {
			fmt.Fprint(&message, "\n\nGift leaderboard")
			for index, score := range scores {
				fmt.Fprintf(&message, "\n%d. %s — %.0f pts", index+1, score.Member, score.Points)
			}
		}
	}
	return message.String()
}
