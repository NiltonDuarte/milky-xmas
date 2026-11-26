package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestLocalChatShowsAQuoteAfterEnoughMessages(t *testing.T) {
	tuning := core.DefaultTuning()
	tuning.Quotes = core.QuoteRules{EveryMessages: 2}
	// A roll of 1 keeps the random drop away.
	service := app.NewService(nil, time.Now, func(n int) int { return min(1, n-1) }).WithTuning(tuning)
	handler := newHandler(service, newGrinchFixture(time.Now, nil, tuning.Grinch))
	for i, want := range []string{"Message recorded.", "-# Charles Dickens"} {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(url.Values{"command": {"/message hi"}}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if !strings.Contains(response.Body.String(), want) {
			t.Fatalf("message %d = %q, want %q", i+1, response.Body.String(), want)
		}
	}
}
