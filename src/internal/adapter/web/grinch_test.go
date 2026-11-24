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

func TestLocalGrinch(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	service := app.NewService(store, func() time.Time { return now }, func(int) int { return 0 })
	handler := newHandler(service, newGrinchFixture(func() time.Time { return now }, store, core.DefaultTuning().Grinch))
	send := func(command, want string) {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(url.Values{"command": {command}}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), want) {
			t.Fatalf("%s: %d %q, want %q", command, response.Code, response.Body, want)
		}
	}
	send("/grinch summon", "🎅 0 joined")
	send("/grinch join", "🎅 1 joined")
	send("/grinch join", "🎅 1 joined")
	now = now.Add(time.Minute)
	send("/grinch view", "100.00/100% HP\nsnowballs")
	for range 24 {
		send("/grinch hit 1", "HP\nsnowballs")
	}
	send("/grinch hit 1", "75.00/100% HP\ndefense")
	// The deterministic board is 0b01010101. Slot 2 is an unsafe defense.
	send("/grinch hit 2", "HP\ndefense")
	send("/grinch hit 2", "45.00/100% HP\nattack")
	for range 7 {
		send("/grinch hit 1", "HP\nattack")
	}
	send("/grinch hit 1", "5.00/100% HP\nfinal")
	send("/grinch hit 1", "5.00/100% HP\nfinal")
	send("/grinch view", "5 ❄️")
	send("/grinch hit 5", "5.00/100% HP\nfinal")
	send("/grinch hit 5", "5.00/100% HP\nfinal")
	send("/look in-your-stocking", "❄️ Snowballs: 2 | 🟡 Yellow Snowballs: 0\n🌑 Coal: 0")
	for range 4 {
		send("/grinch hit 2", "HP\nfinal")
	}
	send("/grinch hit 2", "Damage leaderboard\n1. fixture — 105.00 pts\n\nGift leaderboard\n1. fixture — 2 pts")
	send("/grinch view", "👽 Grinch · 0.00/100% HP")
	send("/look in-your-stocking", "❄️ Snowballs: 2 | 🟡 Yellow Snowballs: 0\n🌑 Coal: 0")
	now = now.Add(time.Second)
	send("/grinch summon", "🎅 0 joined")
	send("/grinch join", "🎅 1 joined")
	send("/grinch join", "🎅 1 joined")
	now = now.Add(time.Minute)
	send("/grinch view", "100.00/100% HP\nsnowballs")
	now = now.Add(2 * time.Minute)
	send("/grinch hit 1", "99.00/100% HP\nsnowballs")
	for range 23 {
		send("/grinch hit 1", "HP\nsnowballs")
	}
	send("/grinch hit 1", "75.00/100% HP\ndefense")
	now = now.Add(time.Minute)
	send("/grinch view", "75.00/100% HP\nescaped")
	send("/look in-your-stocking", "❄️ Snowballs: 2 | 🟡 Yellow Snowballs: 1\n🌑 Coal: 1")
}
