package giftstore_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestAdventPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	now := time.Date(2026, 12, 24, 12, 0, 0, 0, time.UTC)
	open := func() (*giftstore.SQLite, app.Service) {
		t.Helper()
		store, err := giftstore.OpenSQLite(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { store.Close() })
		return store, app.NewService(store, func() time.Time { return now })
	}
	db, service := open()
	secondDB, second := open()
	member := app.Member{ServerID: "server", UserID: "alice"}
	daily, err := service.ClaimGift(t.Context(), member)
	if err != nil || !daily.Collected {
		t.Fatalf("daily claim = %+v, %v", daily, err)
	}

	type result struct {
		opening app.AdventOpening
		err     error
	}
	results := make(chan result, 2)
	start := make(chan struct{})
	for _, svc := range []app.Service{service, second} {
		go func() {
			<-start
			got, err := svc.OpenAdvent(t.Context(), member, 1)
			results <- result{got, err}
		}()
	}
	close(start)
	awards := 0
	for range 2 {
		got := <-results
		if got.err != nil || got.opening.Status != core.DoorOpened || got.opening.Total != 1 {
			t.Errorf("opening = %+v", got)
		}
		if got.opening.Awarded {
			awards++
		}
	}
	if awards != 1 {
		t.Errorf("awards = %d, want 1", awards)
	}
	retry, err := service.ClaimGift(t.Context(), member)
	if err != nil || retry.Collected || retry.Total != daily.Total || retry.NextClaim != daily.NextClaim {
		t.Fatalf("daily claim after Advent = %+v, %v", retry, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := secondDB.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, service := open()
	view, err := service.ViewAdvent(t.Context(), member)
	if err != nil || view.Total != 1 || view.Doors[0] != core.DoorOpened {
		t.Fatalf("calendar after restart = %+v, %v", view, err)
	}
	bob := app.Member{ServerID: "server", UserID: "bob"}
	other, err := service.ViewAdvent(t.Context(), bob)
	if err != nil || other.Total != 0 || other.Doors[0] != core.DoorAvailable {
		t.Fatalf("other member calendar = %+v, %v", other, err)
	}
	now = now.AddDate(1, 0, 0)
	award, err := service.OpenAdvent(t.Context(), member, 1)
	if err != nil || !award.Awarded || award.Total != 1 {
		t.Fatalf("next year opening = %+v, %v", award, err)
	}
	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
	if award, err := service.OpenAdvent(t.Context(), member, 2); err == nil || award.Awarded {
		t.Fatalf("opening with failed storage = %+v, %v", award, err)
	}
}
