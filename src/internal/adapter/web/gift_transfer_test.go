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
	handler := web.NewHandler(app.NewService(store, func() time.Time { return now }, func(n int) int {
		if n == core.GiftRarityRolls {
			return 95
		}
		return 1
	}))
	for _, step := range []struct {
		command string
		status  int
		reply   string
	}{
		{"/aura check", http.StatusOK, "Bad deeds 0%  |  Good deeds 0%"},
		{"/daily claim", http.StatusOK, "🎅🏻 Christmas won’t be Christmas without any presents. You collected 2 gifts 🎁.\nGift rarity · epic."},
		{"/send gift friend normal 1", http.StatusBadRequest, "You don't have enough gifts of that rarity."},
		{"/send gift fixture epic 1", http.StatusBadRequest, "Choose another player in this server."},
		{"/send gift friend epic -1", http.StatusBadRequest, "Enter a positive whole number of gifts."},
		{"/send gift friend epic 1.5", http.StatusBadRequest, "Enter a positive whole number of gifts."},
		{"/send gift friend mystery 1", http.StatusBadRequest, "Choose a gift rarity: normal, rare, or epic."},
		{"/send gift friend epic 2", http.StatusOK, "Sent 2 epic gifts. Good deeds 60%."},
		{"/send gift friend epic 1", http.StatusBadRequest, "You don't have enough gifts of that rarity."},
		{"/aura check", http.StatusOK, "Bad deeds 0%  |  Good deeds 60%"},
		{"/daily claim", http.StatusOK, "🎅 you already got the gift for today, the next gift 🎁 is in 24 hours."},
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
	if err != nil || friend.Inventory.Epic != 2 || friend.Aura != (core.Aura{}) || !friend.LastClaim.IsZero() {
		t.Fatalf("recipient record = %+v, %v", friend, err)
	}
}
