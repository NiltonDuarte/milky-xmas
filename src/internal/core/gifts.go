package core

import "time"

const GiftCooldown = 24 * time.Hour

type GiftRecord struct {
	Total     int64
	LastClaim time.Time
}

func (record GiftRecord) Collect(now time.Time) (GiftRecord, time.Time, bool) {
	if record.Total > 0 {
		next := record.LastClaim.Add(GiftCooldown)
		if now.Before(next) {
			return record, next, false
		}
	}
	return GiftRecord{Total: record.Total + 1, LastClaim: now}, now.Add(GiftCooldown), true
}
