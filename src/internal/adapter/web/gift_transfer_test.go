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
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGiftTransferAndAuraInChat(t *testing.T) {
	store := &giftfixture.Store{}
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	seed := app.GiftUpdate{
		Member: app.Member{ServerID: "local", UserID: "fixture"},
		Next:   core.GiftRecord{Inventory: core.Inventory{Coal: 2}, Aura: core.Aura{Bad: 95, Good: 90}},
	}
	if saved, err := store.Save(t.Context(), seed); err != nil || !saved {
		t.Fatalf("seed aura = %v, %v", saved, err)
	}
	handler := web.NewHandler(app.NewService(store, func() time.Time { return now }, func(n int) int {
		if n == core.GiftRarityRolls {
			return 95
		}
		return 1
	}), nil)
	for _, step := range []struct {
		command string
		status  int
		reply   string
	}{
		{"/aura check", http.StatusOK, "```\n😈 Bad 95%  |  90% Good 😇\n```"},
		{"/daily claim", http.StatusOK, "## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected 2 gifts 🎁.\nGift rarity · epic ·\n```"},
		{"/send gift friend normal 1", http.StatusBadRequest, "You don't have enough gifts of that rarity."},
		{"/send gift fixture epic 1", http.StatusBadRequest, "Choose another player in this server."},
		{"/send gift friend epic -1", http.StatusBadRequest, "Enter a positive whole number of items."},
		{"/send gift friend epic 1.5", http.StatusBadRequest, "Enter a positive whole number of items."},
		{"/send gift friend mystery 1", http.StatusBadRequest, "Choose normal, rare, epic, or coal."},
		{"/send gift friend epic 2", http.StatusOK, "Sent 2 🎁epic gifts to friend. Your 😇 good deeds is 10% more high."},
		{"/send gift friend epic 1", http.StatusBadRequest, "You don't have enough gifts of that rarity."},
		{"/aura check", http.StatusOK, "```\n😈 Bad 95%  |  100% Good 😇\n```"},
		{"/daily claim", http.StatusOK, "🎅 you already got the gift for today, the next gift 🎁 is in 24 hours."},
		{"/send gift friend coal 3", http.StatusBadRequest, "You don't have enough Coal."},
		{"/aura check", http.StatusOK, "```\n😈 Bad 95%  |  100% Good 😇\n```"},
		{"/send gift friend coal 2", http.StatusOK, "Sent 2 🌑 Coal to friend. Your 😈 bad deeds increased by 5%."},
		{"/aura check", http.StatusOK, "```\n😈 Bad 100%  |  100% Good 😇\n```"},
	} {
		form := url.Values{"command": {step.command}}
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != step.status || strings.TrimSpace(response.Body.String()) != step.reply {
			t.Errorf("%s returned %d %q, want %d %q", step.command, response.Code, response.Body.String(), step.status, step.reply)
		}
	}
	friend, err := store.Load(t.Context(), app.Member{ServerID: "local", UserID: "friend"})
	if err != nil || friend.Inventory.Epic != 2 || friend.Inventory.Coal != 2 || friend.Aura != (core.Aura{}) || !friend.LastClaim.IsZero() {
		t.Fatalf("recipient record = %+v, %v", friend, err)
	}
}

func TestGiveGiftInChat(t *testing.T) {
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }), nil)
	post := func(command string) (int, string) {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command="+url.QueryEscape(command)))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code, strings.TrimSpace(recorder.Body.String())
	}
	for _, step := range []struct {
		command string
		code    int
		reply   string
	}{
		{"/give gift friend 3 rare", http.StatusOK, "Gave 3 🎁rare gifts to friend."},
		{"/give gift friend 0 rare", http.StatusBadRequest, "Enter a positive whole number of items."},
		{"/give gift friend 1 mythic", http.StatusBadRequest, "Choose normal, rare, epic, or coal."},
		{"/give gift fixture 1 rare", http.StatusBadRequest, "Choose another player in this server."},
		{"/give gift friend", http.StatusBadRequest, "Use /give gift <recipient> <amount> <rarity>."},
		{"/look in-your-stocking", http.StatusOK, "## 🧦 Your inventory\n```\n🎁 Gifts: 0 - normal | 0 rare | 0 epic\n❄️ Snowballs: 0 | 🟡 Yellow Snowballs: 0\n🌑 Coal: 0\n```"},
	} {
		if code, reply := post(step.command); code != step.code || reply != step.reply {
			t.Errorf("%s = %d %q, want %d %q", step.command, code, reply, step.code, step.reply)
		}
	}
}

func TestHelpInChat(t *testing.T) {
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }), nil)
	for _, step := range []struct {
		command string
		code    int
		contain string
	}{
		{"/helpmemilky", http.StatusOK, "/look in-your-stocking"},
		{"/helpmemilky", http.StatusOK, "/give gift"},
		{"/helpmemilky now", http.StatusBadRequest, "Use /helpmemilky."},
	} {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command="+url.QueryEscape(step.command)))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != step.code || !strings.Contains(recorder.Body.String(), step.contain) {
			t.Errorf("%s = %d %q", step.command, recorder.Code, recorder.Body.String())
		}
	}
}

func TestWaterInChat(t *testing.T) {
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }), nil)
	post := func(command string) (int, string) {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command="+url.QueryEscape(command)))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code, recorder.Body.String()
	}
	if code, body := post("/water"); code != http.StatusOK || !strings.Contains(body, "fixture watered the tree! +2%") ||
		!strings.Contains(body, "💧 Water level: 12%") || !strings.Contains(body, "1. fixture — 1") {
		t.Fatalf("first /water = %d %q", code, body)
	}
	if code, body := post("/water"); code != http.StatusOK || !strings.Contains(body, "You watered the tree a moment ago") {
		t.Fatalf("second /water = %d %q", code, body)
	}
	if code, body := post("/water now"); code != http.StatusBadRequest || !strings.Contains(body, "Use /water.") {
		t.Fatalf("/water now = %d %q", code, body)
	}
}
