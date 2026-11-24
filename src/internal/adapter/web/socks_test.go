package web_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/web"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type unavailableInventory struct{ giftfixture.Store }

func (*unavailableInventory) Load(context.Context, app.Member) (core.GiftRecord, error) {
	return core.GiftRecord{}, errors.New("test inventory read failure")
}

func TestSocksInChat(t *testing.T) {
	store := &giftfixture.Store{}
	update := app.GiftUpdate{
		Member: app.Member{ServerID: "local", UserID: "fixture"},
		Next:   core.GiftRecord{Inventory: core.Inventory{Normal: 1, Rare: 2, Epic: 3, Snowballs: 4, YellowSnowballs: 6, Coal: 5}},
	}
	if saved, err := store.Save(t.Context(), update); err != nil || !saved {
		t.Fatalf("seed inventory = %v, %v", saved, err)
	}
	for _, step := range []struct {
		name  string
		store app.Store
		code  int
		reply string
	}{
		{"filled", store, http.StatusOK, "## 🧦 Your inventory\n```\n🎁 Gifts: 1 - normal | 2 rare | 3 epic\n❄️ Snowballs: 4 | 🟡 Yellow Snowballs: 6\n🌑 Coal: 5\n```"},
		{"empty", &giftfixture.Store{}, http.StatusOK, "## 🧦 Your inventory\n```\n🎁 Gifts: 0 - normal | 0 rare | 0 epic\n❄️ Snowballs: 0 | 🟡 Yellow Snowballs: 0\n🌑 Coal: 0\n```"},
		{"read failure", &unavailableInventory{}, http.StatusInternalServerError, "Could not load your inventory. Try again."},
	} {
		t.Run(step.name, func(t *testing.T) {
			handler := web.NewHandler(app.NewService(step.store, time.Now, func(int) int { return 0 }), nil)
			request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command=%2Fsocks"))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if got := strings.TrimSpace(recorder.Body.String()); recorder.Code != step.code || got != step.reply {
				t.Fatalf("/socks = %d %q, want %d %q", recorder.Code, got, step.code, step.reply)
			}
		})
	}
}
