package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/web"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func TestChatPage(t *testing.T) {
	response := httptest.NewRecorder()
	web.NewHandler(app.Service{}, nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET / returned %d", response.Code)
	}
	for _, fragment := range []string{`role="log"`, `action="/chat"`, `name="command"`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Errorf("chat page is missing %q", fragment)
		}
	}
}

func TestChatCommands(t *testing.T) {
	handler := web.NewHandler(app.Service{}, nil)
	for _, tt := range []struct {
		name, input, origin, reply string
		status                     int
	}{
		{name: "hello", input: "/hello", status: http.StatusOK, reply: "Hello, world!"},
		{name: "unknown", input: "/unknown", status: http.StatusBadRequest, reply: "Unknown command. Try /hello, /daily claim, /advent view, /grinch view, /send gift, /aura check, /look in-your-stocking, or /helpmemilky."},
		{name: "missing slash", input: "hello", status: http.StatusBadRequest, reply: "Unknown command. Try /hello, /daily claim, /advent view, /grinch view, /send gift, /aura check, /look in-your-stocking, or /helpmemilky."},
		{name: "hello after error", input: "/hello", status: http.StatusOK, reply: "Hello, world!"},
		{name: "foreign origin", input: "/hello", origin: "https://example.org", status: http.StatusForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{"command": {tt.input}}
			request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.origin != "" {
				request.Header.Set("Origin", tt.origin)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.status {
				t.Fatalf("POST /chat returned %d, want %d", response.Code, tt.status)
			}
			if tt.reply != "" {
				if got := strings.TrimSpace(response.Body.String()); got != tt.reply {
					t.Errorf("reply = %q, want %q", got, tt.reply)
				}
			}
		})
	}
}

func TestDailyClaimInChat(t *testing.T) {
	first := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	now := first
	roll := 0
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, func() time.Time { return now }, func(int) int { return roll }), nil)
	for _, step := range []struct {
		elapsed time.Duration
		roll    int
		reply   string
	}{
		{0, 0, "## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected 1 gifts 🎁.\nGift rarity · normal ·\n```"},
		{30 * time.Minute, 3, "🎅 you already got the gift for today, the next gift 🎁 is in 23 hours 30 minutes."},
		{24*time.Hour - time.Second, 6, "🎅 you already got the gift for today, the next gift 🎁 is in 1 minute."},
		{24 * time.Hour, 9, "## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected 10 gifts 🎁.\nGift rarity · normal ·\n```"},
	} {
		now = first.Add(step.elapsed)
		roll = step.roll
		form := url.Values{"command": {"/daily claim"}}
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.String() != step.reply {
			t.Errorf("claim at %v got %d %q, want %q", now, response.Code, response.Body, step.reply)
		}
	}
}
