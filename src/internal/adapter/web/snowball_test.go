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
	if saved, err := store.Save(t.Context(), app.GiftUpdate{Member: sender, Next: core.GiftRecord{Inventory: core.Inventory{Snowballs: 2, YellowSnowballs: 1}}}); err != nil || !saved {
		t.Fatalf("seed yellow snowball = %v, %v", saved, err)
	}
	service := app.NewService(store, time.Now, func(int) int { return 0 })
	handler := web.NewHandler(service, nil)
	for _, step := range []struct {
		command string
		code    int
		reply   string
	}{
		{"/throw snowball fixture", http.StatusBadRequest, "Choose another player in this server."},
		{"/throw snowball Bob", http.StatusOK, "You have thrown a ❄️ snowball at Bob. 🥶\nBob❕  fixture has thrown a ❄️ snowball at you. 🥶"},
		{"/throw snowball Bob", http.StatusOK, "You have thrown a ❄️ snowball at Bob. 🥶\nBob❕  fixture has thrown a ❄️ snowball at you. 🥶"},
		{"/throw snowball Bob", http.StatusBadRequest, "You don't have a Snowball to throw."},
		{"/throw snowball Bob yellow", http.StatusOK, "You have thrown a 🟡 yellow snowball at Bob. 🥶\nBob❕  fixture has thrown a 🟡 yellow snowball at you. 🥶"},
		{"/throw snowball Bob yellow", http.StatusBadRequest, "You don't have a Yellow Snowball to throw."},
		{"/aura check", http.StatusOK, "```\n😈 Bad 5%  |  0% Good 😇\n```"},
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
	if inventory, err := service.ViewInventory(t.Context(), sender); err != nil || inventory != (core.Inventory{}) {
		t.Fatalf("sender inventory = %+v, %v", inventory, err)
	}
}
