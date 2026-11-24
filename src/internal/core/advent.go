package core

import (
	"math/bits"
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
	Year  int
	Total int
	Doors [AdventDays]DoorStatus
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

func AdventView(now time.Time, opened uint32) AdventCalendar {
	calendar := AdventCalendar{Year: now.UTC().Year(), Total: bits.OnesCount32(opened)}
	for i := range calendar.Doors {
		calendar.Doors[i] = AdventStatus(now, i+1, opened&(1<<i) != 0)
	}
	return calendar
}
