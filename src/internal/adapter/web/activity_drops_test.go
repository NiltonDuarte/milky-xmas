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

func TestLocalActivityGiftRain(t *testing.T) {
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	store := &giftfixture.Store{}
	service := app.NewService(store, func() time.Time { return now }, func(bound int) int {
		switch bound {
		case core.GiftRainMessageRange:
			return 0
		case core.RandomActiveDropRolls:
			return 1
		default:
			return 70
		}
	})
	handler := newHandler(service, newGrinchFixture(func() time.Time { return now }, store), app.NewActivityDrops(service))
	send := func(command string) string {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(url.Values{"command": {command}}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s = %d %q", command, response.Code, response.Body.String())
		}
		return response.Body.String()
	}
	for range core.GiftRainMinimumMessages - 1 {
		if got := send("/message festive chat"); got != "Message recorded." {
			t.Fatalf("activity reply = %q", got)
		}
	}
	if got := send("/message festive chat"); !strings.Contains(got, "🎁 It's raining gifts! 🎁 Hallelujah\n🎁 🎁 🎁 🎁\n🎁 🎁 🎁 🎁\n/drop collect 1") {
		t.Fatalf("rain presentation = %q", got)
	}
	if got := send("/drop collect 1"); got != "You collected a rare gift." {
		t.Fatalf("claim = %q", got)
	}
	if got := send("/drop collect 1"); got != "You cannot collect another gift yet." || strings.Contains(got, "minute") {
		t.Fatalf("cooldown reply = %q", got)
	}
}
