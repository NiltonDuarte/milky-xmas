package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func TestLocalGrinch(t *testing.T) {
	now := time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC)
	handler := newHandler(app.Service{}, newGrinchFixture(func() time.Time { return now }))
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
	send("/grinch join", "You already joined")
	now = now.Add(time.Minute)
	send("/grinch view", "100.00/100 HP · snowballs")
	for range 24 {
		send("/grinch hit 1", "Snowball hit for 1 HP.")
	}
	send("/grinch hit 1", "75.00/100 HP · defense")
	// The deterministic board is 0b01010101. Slot 2 is an unsafe defense.
	send("/grinch hit 2", "You lost 10%")
	send("/grinch hit 2", "45.00/100 HP · attack")
	for range 7 {
		send("/grinch hit 1", "Attack hit!")
	}
	send("/grinch hit 1", "5.00/100 HP · final")
	send("/grinch hit 1", "You earned 2 gift points!")
	for range 4 {
		send("/grinch hit 2", "Final attack hit!")
	}
	send("/grinch hit 2", "Damage leaderboard\n1. fixture — 105.00 pts\n\nGift leaderboard\n1. fixture — 2 pts")
	now = now.Add(time.Second)
	send("/grinch summon", "🎅 0 joined")
	send("/grinch join", "🎅 1 joined")
	send("/grinch join", "You already joined")
	now = now.Add(time.Minute)
	send("/grinch view", "100.00/100 HP · snowballs")
	now = now.Add(2 * time.Minute)
	send("/grinch hit 1", "99.00/100 HP · snowballs")
	for range 24 {
		send("/grinch hit 1", "Snowball hit for 1 HP.")
	}
	send("/grinch view", "75.00/100 HP · defense")
	now = now.Add(time.Minute)
	send("/grinch view", "75.00/100 HP · escaped")
}
