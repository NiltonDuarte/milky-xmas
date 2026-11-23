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
	var total int64
	for round, amount := range []int64{3, 10} {
		now = first.Add(time.Duration(round) * 24 * time.Hour)
		total += amount
		start := make(chan struct{})
		results := make(chan result, 2)
		for _, db := range []*giftstore.SQLite{store, other} {
			go func() {
				<-start
				service := app.NewService(db, clock, func(int) int { return int(amount - 1) })
				claim, err := service.ClaimGift(t.Context(), member)
				results <- result{claim, err}
			}()
		}
		close(start)
		winners := 0
		for range 2 {
			got := <-results
			if got.err != nil || got.claim.Total != total || !got.claim.NextClaim.Equal(now.Add(24*time.Hour)) {
				t.Errorf("concurrent claim = %+v", got)
			}
			if got.claim.Collected {
				winners++
				if got.claim.Amount != amount {
					t.Errorf("winning claim awarded %d gifts, want %d", got.claim.Amount, amount)
				}
			} else if got.claim.Amount != 0 {
				t.Errorf("cooldown response awarded %d gifts", got.claim.Amount)
			}
		}
		if winners != 1 {
			t.Errorf("round %d had %d successful claims, want 1", round+1, winners)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}

	service := app.NewService(open(), clock, func(int) int { return 6 })
	retry, err := service.ClaimGift(t.Context(), member)
	if err != nil || retry.Collected || retry.Total != total || !retry.NextClaim.Equal(first.Add(48*time.Hour)) {
		t.Fatalf("claim after reopening = %+v, %v", retry, err)
	}
	for _, identity := range []app.Member{{ServerID: "two", UserID: "alice"}, {ServerID: "one", UserID: "bob"}} {
		claim, err := service.ClaimGift(t.Context(), identity)
		if err != nil || !claim.Collected || claim.Amount != 7 || claim.Total != 7 {
			t.Fatalf("claim for %+v = %+v, %v", identity, claim, err)
		}
	}
	now = first.Add(48 * time.Hour)
	claim, err := service.ClaimGift(t.Context(), member)
	if err != nil || !claim.Collected || claim.Amount != 7 || claim.Total != total+7 {
		t.Fatalf("claim at stored boundary = %+v, %v", claim, err)
	}
}
