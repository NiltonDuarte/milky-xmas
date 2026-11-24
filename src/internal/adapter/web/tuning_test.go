package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func tunedChat(t *testing.T, tuning core.Tuning) func(string) string {
	t.Helper()
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	service := app.NewService(store, func() time.Time { return now }, func(bound int) int {
		if bound == tuning.Drops.RandomActiveOneIn {
			return 1 // never a Random Active Drop
		}
		return 0
	}).WithTuning(tuning)
	handler := NewHandler(service, store)
	return func(command string) string {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(url.Values{"command": {command}}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s = %d %q", command, response.Code, response.Body.String())
		}
		return response.Body.String()
	}
}

func TestLocalChatAppliesTunedGiftRain(t *testing.T) {
	tuning := core.DefaultTuning()
	tuning.Drops.GiftRain.MinMessages, tuning.Drops.GiftRain.MaxMessages = 2, 2
	tuning.Drops.GiftRain.ClaimsToEnd, tuning.Drops.GiftRain.ClaimCooldown = 2, 0
	tuning.Drops.GiftRain.CoalPercent = 0
	send := tunedChat(t, tuning)

	if got := send("/message festive chat"); got != "Message recorded." {
		t.Fatalf("first message = %q", got)
	}
	if got := send("/message festive chat"); !strings.Contains(got, "It's raining gifts") {
		t.Fatalf("second message did not start the rain: %q", got)
	}
	if got := send("/drop collect 1"); strings.Contains(got, "expired") {
		t.Fatalf("first collection ended the rain: %q", got)
	}
	if got := send("/drop collect 1"); !strings.Contains(got, "⌛ This drop has expired.") {
		t.Fatalf("second collection did not end the rain: %q", got)
	}
	if got := send("/drop collect 1"); got != "This gift drop is no longer available." {
		t.Fatalf("collection after the end = %q", got)
	}
}

func TestLocalChatAppliesTunedGrinchFight(t *testing.T) {
	tuning := core.DefaultTuning()
	tuning.Grinch.EarlyStartPlayers = 1
	tuning.Grinch.SnowballsBossHP = 9000
	send := tunedChat(t, tuning)

	send("/grinch summon")
	if got := send("/grinch join"); !strings.Contains(got, "snowballs") {
		t.Fatalf("one player did not start the fight with an early start of one: %q", got)
	}
	for range 9 {
		send("/grinch hit 1")
	}
	if got := send("/grinch view"); !strings.Contains(got, "91.00/100% HP") || !strings.Contains(got, "snowballs") {
		t.Fatalf("view after nine snowballs = %q", got)
	}
	send("/grinch hit 1")
	if got := send("/grinch view"); !strings.Contains(got, "90.00/100% HP") || !strings.Contains(got, "defense") {
		t.Fatalf("view after ten snowballs = %q, want the defense phase at 90 HP", got)
	}
}
