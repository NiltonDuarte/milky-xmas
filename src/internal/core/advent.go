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

func AdventStatus(now time.Time, day int, opened bool) DoorStatus {
	if day < 1 || day > AdventDays {
		return DoorInvalid
	}
	if opened {
		return DoorOpened
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
