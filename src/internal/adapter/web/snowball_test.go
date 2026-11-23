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

func TestThrownSnowballInChat(t *testing.T) {
	store := &giftfixture.Store{}
	sender := app.Member{ServerID: "local", UserID: "fixture"}
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: sender, Next: core.GiftRecord{Inventory: core.Inventory{YellowSnowballs: 1}}}); err != nil || !saved {
		t.Fatalf("seed yellow snowball = %v, %v", saved, err)
	}
	service := app.NewService(store, time.Now, func(int) int { return 0 })
	handler := web.NewHandler(service, nil)
	for _, step := range []struct {
		command string
		code    int
		reply   string
	}{
		{"/thrown snowball fixture", http.StatusBadRequest, "Choose another player in this server."},
		{"/thrown snowball Bob", http.StatusOK, "You have thrown a ❄️ snowball at Bob. 🥶\nfixture have threw a ❄️ snowball at you, now you have 1 ❄️ snowball. 🥶"},
		{"/thrown snowball Bob", http.StatusOK, "You have thrown a ❄️ snowball at Bob. 🥶\nfixture have threw a ❄️ snowball at you, now you have 2 ❄️ snowballs. 🥶"},
		{"/thrown snowball Bob yellow", http.StatusOK, "You have thrown a 🟡 yellow snowball at Bob. 🥶\nfixture have threw a 🟡 yellow snowball at you, now you have 1 🟡 yellow snowball. 🥶"},
		{"/thrown snowball Bob yellow", http.StatusBadRequest, "You don't have a Yellow Snowball to throw."},
		{"/aura check", http.StatusOK, "😈 Bad 5%  |  😇 Good 0%"},
	} {
		form := url.Values{"command": {step.command}}
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != step.code || strings.TrimSpace(response.Body.String()) != step.reply {
			t.Errorf("%s = %d %q", step.command, response.Code, response.Body.String())
		}
	}
	inventory, err := service.ViewInventory(t.Context(), app.Member{ServerID: "local", UserID: "Bob"})
	if err != nil || inventory.Snowballs != 2 || inventory.YellowSnowballs != 1 {
		t.Fatalf("recipient inventory = %+v, %v", inventory, err)
	}
	if inventory, err := service.ViewInventory(t.Context(), sender); err != nil || inventory.YellowSnowballs != 0 {
		t.Fatalf("sender inventory = %+v, %v", inventory, err)
	}
}
