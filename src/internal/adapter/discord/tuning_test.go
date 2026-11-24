package discord

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGrinchMessagesQuoteTheDefaultNumbers(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name  string
		state core.Grinch
		want  string
	}{
		{"lobby", core.Grinch{ID: "f", Phase: core.GrinchLobby, StartedAt: start, HP: core.GrinchBossHP}, "**0/3 to start early**"},
		{"snowballs", core.Grinch{ID: "f", Phase: core.GrinchSnowballs, StartedAt: start, HP: 9000},
			"Throw snowballs until the Grinch reaches **75 HP**. The one-minute timer starts at 75 HP."},
		{"final", core.Grinch{ID: "f", Phase: core.GrinchFinalAttack, StartedAt: start, TimedAt: start, HP: 500},
			"⚔️ Attack for 2 damage points, pick 🎁 for 1–5 gift points, or pick ❄️ for 1 Snowball."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if content, _ := grinchMessage(rules, tt.state); !strings.Contains(content, tt.want) {
				t.Fatalf("message = %q, want it to contain %q", content, tt.want)
			}
		})
	}
}

func TestGrinchMessagesQuoteTunedNumbers(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	rules.EarlyStartPlayers = 2
	rules.RegistrationDuration = 2 * time.Minute
	rules.SnowballsBossHP = 8250
	rules.FightDuration = 90 * time.Second
	rules.FinalHitScore = 350
	rules.GiftPointsMin, rules.GiftPointsMax = 2, 4
	rules.FinalSnowballReward = 2
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)

	lobby, _ := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchLobby, StartedAt: start, HP: core.GrinchBossHP})
	if !strings.Contains(lobby, "**0/2 to start early**") || !strings.Contains(lobby, "<t:"+strconv.FormatInt(start.Add(2*time.Minute).Unix(), 10)+":R>") {
		t.Errorf("lobby = %q, want a two-player early start and a two-minute registration", lobby)
	}
	snowballs, _ := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchSnowballs, StartedAt: start, HP: 9000})
	if !strings.Contains(snowballs, "reaches **82.5 HP**. The 90-second timer starts at 82.5 HP.") {
		t.Errorf("snowballs = %q", snowballs)
	}
	final, _ := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchFinalAttack, StartedAt: start, TimedAt: start, HP: 500})
	if !strings.Contains(final, "⚔️ Attack for 3.5 damage points, pick 🎁 for 2–4 gift points, or pick ❄️ for 2 Snowballs.") {
		t.Errorf("final = %q", final)
	}
}

func TestTimerName(t *testing.T) {
	for d, want := range map[time.Duration]string{time.Minute: "one-minute", 2 * time.Minute: "2-minute", 90 * time.Second: "90-second", 30 * time.Second: "30-second"} {
		if got := timerName(d); got != want {
			t.Errorf("timerName(%s) = %q, want %q", d, got, want)
		}
	}
}

func TestDropBotUsesTunedLifetime(t *testing.T) {
	tuning := core.DefaultTuning()
	tuning.Drops.Lifetime = 3 * time.Minute
	service := app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }).WithTuning(tuning)
	if got := newDropBot(app.NewActivityDrops(service)).lifetime; got != 3*time.Minute {
		t.Fatalf("lifetime = %s, want 3m0s", got)
	}
}
