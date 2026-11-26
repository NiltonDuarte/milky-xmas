package discord

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
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
			"Escapes <t:" + strconv.FormatInt(start.Add(2*time.Minute).Unix(), 10) + ":R>\nThrow snowballs to bring the Grinch down to **75 HP** before time runs out, or he wins. Then the 3-minute timer starts."},
		{"final", core.Grinch{ID: "f", Phase: core.GrinchFinalAttack, StartedAt: start, TimedAt: start, HP: 500},
			"⚔️ Attack for 0.34 damage, pick 🎁 for 1–5 gift points, or pick ❄️ for 1 Snowball."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if content, _ := grinchMessage(rules, tt.state); !strings.Contains(content, tt.want) {
				t.Fatalf("message = %q, want it to contain %q", content, tt.want)
			}
		})
	}
}

func TestGrinchMessagesQuoteTunedNumbers(t *testing.T) {
	rules := classicFight(core.DefaultTuning().Grinch)
	rules.EarlyStartPlayers = 2
	rules.RegistrationDuration = 2 * time.Minute
	rules.SnowballsBossHP = 8250
	rules.FightDuration = 90 * time.Second
	rules.FinalHitDamage = 150
	rules.GiftPointsMin, rules.GiftPointsMax = 2, 4
	rules.FinalSnowballReward = 2
	start := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)

	label := func(rows []discordgo.MessageComponent) string {
		return rows[len(rows)-1].(discordgo.ActionsRow).Components[0].(discordgo.Button).Label
	}
	lobby, rows := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchLobby, StartedAt: start, HP: core.GrinchBossHP})
	if !strings.Contains(lobby, "Press **Start**") || !strings.Contains(lobby, "starts 2 minutes after the first press, or right away when 2 players press it") ||
		!strings.Contains(lobby, "**0/2 to start early**") || strings.Contains(lobby, "<t:") || label(rows) != "Start" {
		t.Errorf("lobby = %q, label %q, want Start, a two-minute countdown, and a two-player early start", lobby, label(rows))
	}
	counting, _ := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchLobby, StartedAt: start, CountdownAt: start.Add(30 * time.Second),
		Joined: map[string]bool{"alice": true}, HP: core.GrinchBossHP})
	if !strings.Contains(counting, "**1/2 to start early**") || !strings.Contains(counting, "The fight starts <t:"+strconv.FormatInt(start.Add(150*time.Second).Unix(), 10)+":R>") {
		t.Errorf("lobby after the first press = %q, want the countdown to end 2 minutes after it", counting)
	}
	// The combat screen has the phase's buttons and no Join.
	for _, phase := range []core.GrinchPhase{core.GrinchSnowballs, core.GrinchDefense, core.GrinchAttack, core.GrinchFinalAttack} {
		_, rows := grinchMessage(rules, core.Grinch{ID: "f", Phase: phase, StartedAt: start, TimedAt: start, HP: 9000, Joined: map[string]bool{"a": true}, FinalRound: 1})
		data, _ := json.Marshal(rows)
		if len(rows) == 0 || strings.Contains(string(data), "Join") || strings.Contains(string(data), ":registration:") {
			t.Errorf("%s: combat buttons = %s, want the phase's buttons and no Join", phase, data)
		}
	}
	snowballs, _ := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchSnowballs, StartedAt: start, HP: 9000})
	if !strings.Contains(snowballs, "down to **82.5 HP** before time runs out, or he wins. Then the 90-second timer starts.") {
		t.Errorf("snowballs = %q", snowballs)
	}
	final, _ := grinchMessage(rules, core.Grinch{ID: "f", Phase: core.GrinchFinalAttack, StartedAt: start, TimedAt: start, HP: 500})
	if !strings.Contains(final, "⚔️ Attack for 1.5 damage, pick 🎁 for 2–4 gift points, or pick ❄️ for 2 Snowballs.") {
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
