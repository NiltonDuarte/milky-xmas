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

func TestAdventInChat(t *testing.T) {
	now := time.Date(2026, 12, 4, 0, 0, 0, 0, time.UTC)
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, func() time.Time { return now }))
	for _, step := range []struct{ command, reply string }{
		{"/advent view", "04 · available\n05 · locked"},
		{"/advent open 5", "Door 5 unlocks on December 5"},
		{"/advent open 1", "You opened door 1! Advent gifts: 1."},
		{"/advent open 1", "Door 1 is already open."},
		{"/advent view", "01 · opened\n02 · available"},
		{"/advent open 25", "Choose an Advent door from 1 to 24."},
		{"/daily claim", "You collected one gift! Total gifts: 1."},
	} {
		form := url.Values{"command": {step.command}}
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), step.reply) {
			t.Errorf("%s = %d %q, expected %q", step.command, response.Code, response.Body, step.reply)
		}
	}
}
