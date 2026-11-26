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

func TestStatsInChat(t *testing.T) {
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }), nil)
	post := func(command string) (int, string) {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command="+url.QueryEscape(command)))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder.Code, recorder.Body.String()
	}
	post("/water")
	for command, want := range map[string]string{
		"/stats":          "Waterings: **1**",
		"/stats plants":   "1. fixture — 1",
		"/stats good":     "1. fixture — 1",
		"/stats bad":      "Nobody yet.",
		"/stats gifters":  "Coal sent",
		"/stats throwers": "Yellow snowballs",
	} {
		if code, body := post(command); code != http.StatusOK || !strings.Contains(body, want) {
			t.Errorf("%s = %d %q, want %q", command, code, body, want)
		}
	}
	if code, body := post("/stats boss"); code != http.StatusBadRequest {
		t.Errorf("/stats boss = %d %q", code, body)
	}
}
