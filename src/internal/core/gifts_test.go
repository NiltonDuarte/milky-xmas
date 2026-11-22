package core

import (
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	first := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	record, next, collected := (GiftRecord{}).Collect(first)
	if !collected || record.Total != 1 || next != first.Add(24*time.Hour) {
		t.Fatalf("first claim = %+v, %v, %v", record, next, collected)
	}
	refused, retry, collected := record.Collect(next.Add(-time.Nanosecond))
	if collected || refused != record || retry != next {
		t.Fatalf("early claim = %+v, %v, %v", refused, retry, collected)
	}
	second, retry, collected := record.Collect(next)
	if !collected || second.Total != 2 || retry != next.Add(24*time.Hour) {
		t.Fatalf("claim at boundary = %+v, %v, %v", second, retry, collected)
	}
}
