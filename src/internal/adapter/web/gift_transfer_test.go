package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/web"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGiftTransferAndAuraInChat(t *testing.T) {
	store := &giftfixture.Store{}
	now := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	seed := app.GiftUpdate{
		Member: app.Member{ServerID: "local", UserID: "fixture"},
		Next:   core.GiftRecord{Inventory: core.Inventory{Coal: 2}, Aura: 20},
	}
	if saved, err := store.Save(t.Context(), seed); err != nil || !saved {
		t.Fatalf("seed aura = %v, %v", saved, err)
	}
	// The daily limit has its own tests.
	tuning := core.DefaultTuning()
	tuning.Aura.DailyLimit = 0
	handler := web.NewHandler(app.NewService(store, func() time.Time { return now }, func(n int) int {
		if n == core.GiftRarityRolls {
			return 95
		}
		return 1
	}).WithTuning(tuning), nil)
	for _, step := range []struct {
		command string
		status  int
		reply   string
	}{
		{"/aura check", http.StatusOK, gifttext.Aura(20)},
		{"/daily claim", http.StatusOK, "## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected 2 gifts 🎁.\nGift rarity · epic ·\n```"},
		{"/send gift friend normal 1", http.StatusBadRequest, "You don't have enough gifts of that rarity."},
		{"/send gift fixture epic 1", http.StatusBadRequest, "Choose another player in this server."},
		{"/send gift friend epic -1", http.StatusBadRequest, "Enter a positive whole number of items."},
		{"/send gift friend epic 1.5", http.StatusBadRequest, "Enter a positive whole number of items."},
		{"/send gift friend mystery 1", http.StatusBadRequest, "Choose normal, rare, epic, or coal."},
		{"/send gift friend epic 2", http.StatusOK, "Sent 2 🎁 epic gifts to friend. Your 😇 good deeds are going up."},
		{"/send gift friend epic 1", http.StatusBadRequest, "You don't have enough gifts of that rarity."},
		{"/aura check", http.StatusOK, gifttext.Aura(50)},
		{"/daily claim", http.StatusOK, "🎅 You already got the gift for today, the next gift 🎁 is in 24 hours."},
		{"/send gift friend coal 3", http.StatusBadRequest, "You don't have enough Coal."},
		{"/aura check", http.StatusOK, gifttext.Aura(50)},
		{"/send gift friend coal 2", http.StatusOK, "Sent 2 🌑 Coal to friend. Your 😈 bad deeds are increasing. Beware!"},
		{"/aura check", http.StatusOK, gifttext.Aura(40)},
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
	if err != nil || friend.Inventory.Epic != 2 || friend.Inventory.Coal != 2 || friend.Aura != 0 || !friend.LastClaim.IsZero() {
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
		{"/give gift friend 3 rare", http.StatusOK, "Gave 3 🎁 rare gifts to friend."},
		{"/give gift friend 0 rare", http.StatusBadRequest, "Enter a positive whole number of items."},
		{"/give gift friend 1 mythic", http.StatusBadRequest, "Choose normal, rare, epic, or coal."},
		{"/give gift fixture 1 rare", http.StatusBadRequest, "Choose another player in this server."},
		{"/give gift friend", http.StatusBadRequest, "Use /give gift <recipient> <amount> <rarity>."},
		{"/look in-your-stocking", http.StatusOK, gifttext.Inventory(core.Inventory{Normal: 0, Rare: 0, Epic: 0, Snowballs: 0, YellowSnowballs: 0, Coal: 0}, false)},
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
		{"/help", http.StatusOK, "/throw snowball <player> <type>"},
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
		!strings.Contains(body, "🟦⬜⬜⬜⬜⬜⬜⬜⬜⬜  **12%** 💧") || strings.Contains(body, "Top waterers") {
		t.Fatalf("first /water = %d %q", code, body)
	}
	if code, body := post("/water"); code != http.StatusOK || !strings.Contains(body, "You already watered the tree") {
		t.Fatalf("second /water = %d %q", code, body)
	}
	if code, body := post("/water now"); code != http.StatusBadRequest || !strings.Contains(body, "Use /water.") {
		t.Fatalf("/water now = %d %q", code, body)
	}
}
