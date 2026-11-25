package core

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

const AdventDays = 24

type DoorStatus string

const (
	DoorAvailable DoorStatus = "available"
	DoorOpened    DoorStatus = "opened"
	DoorLocked    DoorStatus = "locked"
	DoorExpired   DoorStatus = "expired"
	DoorInvalid   DoorStatus = "invalid"
)

type AdventCalendar struct {
	Year int
	// Doors holds the status of each door: Doors[0] is door 1.
	Doors [AdventDays]DoorStatus
	// Order lists the door indexes in the order this player sees them.
	Order [AdventDays]int
}

// AdventOrder shuffles the numbers 0 to AdventDays-1. The result depends only on the seed, so
// a player keeps the same order every time without it being stored.
func AdventOrder(seed string) [AdventDays]int {
	var order [AdventDays]int
	for i := range order {
		order[i] = i
	}
	for i := AdventDays - 1; i > 0; i-- {
		sum := sha256.Sum256(fmt.Appendf(nil, "%s|%d", seed, i))
		j := binary.BigEndian.Uint64(sum[:8]) % uint64(i+1)
		order[i], order[j] = order[j], order[i]
	}
	return order
}

// TEMPORARY: a test window that opens every door from November 24 through November 29, 2026 (UTC),
// so the Advent calendar can be tried before December. Remove it after the test.
var (
	adventTestStart = time.Date(2026, time.November, 24, 0, 0, 0, 0, time.UTC)
	adventTestEnd   = time.Date(2026, time.November, 30, 0, 0, 0, 0, time.UTC)
)

// AdventTestWindow reports whether the temporary test window is open.
func AdventTestWindow(now time.Time) bool {
	now = now.UTC()
	return !now.Before(adventTestStart) && now.Before(adventTestEnd)
}

// AdventYearKey is the year under which openings are stored. The test window uses its own key,
// so doors opened while testing do not count as opened in December.
func AdventYearKey(now time.Time) int {
	if AdventTestWindow(now) {
		return 1
	}
	return now.UTC().Year()
}

func AdventStatus(now time.Time, day int, opened bool) DoorStatus {
	if day < 1 || day > AdventDays {
		return DoorInvalid
	}
	if opened {
		return DoorOpened
	}
	if AdventTestWindow(now) {
		return DoorAvailable
	}
	date := now.UTC()
	if date.Month() < time.December || day > date.Day() {
		return DoorLocked
	}
	if date.Day() > AdventDays {
		return DoorExpired
	}
	return DoorAvailable
}

func AdventView(now time.Time, opened uint32, order [AdventDays]int) AdventCalendar {
	calendar := AdventCalendar{Year: now.UTC().Year(), Order: order}
	for i := range calendar.Doors {
		calendar.Doors[i] = AdventStatus(now, i+1, opened&(1<<i) != 0)
	}
	return calendar
}
