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
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
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
		{"filled", store, http.StatusOK, gifttext.Inventory(core.Inventory{Normal: 1, Rare: 2, Epic: 3, Snowballs: 4, YellowSnowballs: 6, Coal: 5}, false)},
		{"empty", &giftfixture.Store{}, http.StatusOK, gifttext.Inventory(core.Inventory{Normal: 0, Rare: 0, Epic: 0, Snowballs: 0, YellowSnowballs: 0, Coal: 0}, false)},
		{"read failure", &unavailableInventory{}, http.StatusInternalServerError, "Could not load your inventory. Try again."},
	} {
		t.Run(step.name, func(t *testing.T) {
			handler := web.NewHandler(app.NewService(step.store, time.Now, func(int) int { return 0 }), nil)
			request := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader("command=%2Flook+in-your-stocking"))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if got := strings.TrimSpace(recorder.Body.String()); recorder.Code != step.code || got != step.reply {
				t.Fatalf("/look in-your-stocking = %d %q, want %d %q", recorder.Code, got, step.code, step.reply)
			}
		})
	}
}
