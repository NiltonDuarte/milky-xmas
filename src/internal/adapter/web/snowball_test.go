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

func TestThrownSnowballInChat(t *testing.T) {
	service := app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 })
	handler := web.NewHandler(service)
	for _, step := range []struct {
		command string
		code    int
		reply   string
	}{
		{"/thrown snowball fixture", http.StatusBadRequest, "Choose another player in this server."},
		{"/thrown snowball Bob", http.StatusOK, "You have thrown a ❄️ snowball at Bob. 🥶\nfixture have threw a ❄️ snowball at you, now you have 1 ❄️ snowball. 🥶"},
		{"/thrown snowball Bob", http.StatusOK, "You have thrown a ❄️ snowball at Bob. 🥶\nfixture have threw a ❄️ snowball at you, now you have 2 ❄️ snowballs. 🥶"},
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
	if err != nil || inventory.Snowballs != 2 {
		t.Fatalf("recipient inventory = %+v, %v", inventory, err)
	}
}
