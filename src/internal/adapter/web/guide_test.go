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

func TestGuideInChat(t *testing.T) {
	handler := web.NewHandler(app.NewService(&giftfixture.Store{}, time.Now, func(int) int { return 0 }), nil)
	for command, want := range map[string]struct {
		status int
		text   string
	}{
		"/guide":          {http.StatusOK, "## 🎄 Milky Xmas guide"},
		"/guide tree":     {http.StatusOK, "## 🌲 Christmas tree"},
		"/guide nothing":  {http.StatusBadRequest, "Use /guide start, gifts"},
		"/mod-guide":      {http.StatusOK, "## 🛠️ Moderator guide"},
		"/mod-guide aura": {http.StatusOK, "## 😇 Aura"},
		"/mod-guide x":    {http.StatusBadRequest, "Use /mod-guide start, setup"},
	} {
		request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(url.Values{"command": {command}}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want.status || !strings.Contains(response.Body.String(), want.text) {
			t.Errorf("%s = %d %q", command, response.Code, response.Body.String())
		}
	}
}
