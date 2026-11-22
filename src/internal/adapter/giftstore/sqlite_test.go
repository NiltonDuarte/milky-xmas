package giftstore_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftstore"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func TestPersistentClaims(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gifts.sqlite")
	open := func() *giftstore.SQLite {
		t.Helper()
		store, err := giftstore.OpenSQLite(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { store.Close() })
		return store
	}

	store, other := open(), open()
	first := time.Date(2026, 12, 1, 18, 0, 0, 123, time.UTC)
	now := first
	clock := func() time.Time { return now }
	member := app.Member{ServerID: "one", UserID: "alice"}
	type result struct {
		claim app.GiftClaim
		err   error
	}
	for round := int64(1); round <= 2; round++ {
		now = first.Add(time.Duration(round-1) * 24 * time.Hour)
		start := make(chan struct{})
		results := make(chan result, 2)
		for _, db := range []*giftstore.SQLite{store, other} {
			go func() {
				<-start
				claim, err := app.NewService(db, clock).ClaimGift(t.Context(), member)
				results <- result{claim, err}
			}()
		}
		close(start)
		winners := 0
		for range 2 {
			got := <-results
			if got.err != nil || got.claim.Total != round || !got.claim.NextClaim.Equal(now.Add(24*time.Hour)) {
				t.Errorf("concurrent claim = %+v", got)
			}
			if got.claim.Collected {
				winners++
			}
		}
		if winners != 1 {
			t.Errorf("round %d awarded %d gifts", round, winners)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}

	service := app.NewService(open(), clock)
	retry, err := service.ClaimGift(t.Context(), member)
	if err != nil || retry.Collected || retry.Total != 2 || !retry.NextClaim.Equal(first.Add(48*time.Hour)) {
		t.Fatalf("claim after reopening = %+v, %v", retry, err)
	}
	for _, identity := range []app.Member{{ServerID: "two", UserID: "alice"}, {ServerID: "one", UserID: "bob"}} {
		claim, err := service.ClaimGift(t.Context(), identity)
		if err != nil || !claim.Collected || claim.Total != 1 {
			t.Fatalf("claim for %+v = %+v, %v", identity, claim, err)
		}
	}
	now = first.Add(48 * time.Hour)
	claim, err := service.ClaimGift(t.Context(), member)
	if err != nil || !claim.Collected || claim.Total != 3 {
		t.Fatalf("claim at stored boundary = %+v, %v", claim, err)
	}
}
