package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/web"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func TestChatPage(t *testing.T) {
	response := httptest.NewRecorder()
	web.NewHandler(app.Service{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET / returned %d", response.Code)
	}
	for _, fragment := range []string{`role="log"`, `action="/chat"`, `name="command"`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Errorf("chat page is missing %q", fragment)
		}
	}
}

func TestChatCommands(t *testing.T) {
	handler := web.NewHandler(app.Service{})
	for _, tt := range []struct {
		name, input, origin, reply string
		status                     int
	}{
		{name: "hello", input: "/hello", status: http.StatusOK, reply: "Hello, world!"},
		{name: "unknown", input: "/unknown", status: http.StatusBadRequest, reply: "Unknown command. Try /hello."},
		{name: "missing slash", input: "hello", status: http.StatusBadRequest, reply: "Unknown command. Try /hello."},
		{name: "hello after error", input: "/hello", status: http.StatusOK, reply: "Hello, world!"},
		{name: "foreign origin", input: "/hello", origin: "https://example.org", status: http.StatusForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{"command": {tt.input}}
			request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.origin != "" {
				request.Header.Set("Origin", tt.origin)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.status {
				t.Fatalf("POST /chat returned %d, want %d", response.Code, tt.status)
			}
			if tt.reply != "" {
				if got := strings.TrimSpace(response.Body.String()); got != tt.reply {
					t.Errorf("reply = %q, want %q", got, tt.reply)
				}
			}
		})
	}
}
