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

func TestSpiritInChat(t *testing.T) {
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }), nil)
	post := func(command string) (int, string) {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command="+url.QueryEscape(command)))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code, recorder.Body.String()
	}
	post("/water") // One good deed.
	if code, body := post("/spirit"); code != http.StatusOK || !strings.Contains(body, "Good deeds: **1**") ||
		!strings.Contains(body, "The good deeds win!") {
		t.Fatalf("/spirit = %d %q", code, body)
	}
	if code, body := post("/spirit now"); code != http.StatusBadRequest || !strings.Contains(body, "Use /spirit.") {
		t.Fatalf("/spirit now = %d %q", code, body)
	}
}
