package core

import (
	"testing"
	"time"
)

func TestAdventStatus(t *testing.T) {
	for _, step := range []struct {
		when string
		day  int
		want DoorStatus
	}{
		{"2026-11-30T23:59:59Z", 1, DoorLocked},
		{"2026-12-01T00:00:00Z", 1, DoorAvailable},
		{"2026-12-04T01:00:00+02:00", 4, DoorLocked},
		{"2026-12-24T23:59:59.999999999Z", 1, DoorAvailable},
		{"2026-12-25T00:00:00Z", 24, DoorExpired},
		{"2027-12-01T00:00:00Z", 25, DoorInvalid},
	} {
		now, err := time.Parse(time.RFC3339Nano, step.when)
		if err != nil {
			t.Fatal(err)
		}
		if got := AdventStatus(now, step.day, false); got != step.want {
			t.Errorf("day %d at %s = %s, want %s", step.day, step.when, got, step.want)
		}
	}
	view := AdventView(time.Date(2026, 12, 4, 0, 0, 0, 0, time.UTC), 1)
	if view.Year != 2026 || view.Total != 1 || view.Doors[0] != DoorOpened ||
		view.Doors[1] != DoorAvailable || view.Doors[3] != DoorAvailable || view.Doors[4] != DoorLocked {
		t.Errorf("calendar = %+v", view)
	}
}

func TestAdventTestWindow(t *testing.T) {
	for _, tt := range []struct {
		at     time.Time
		window bool
		status DoorStatus
		key    int
	}{
		{time.Date(2026, 11, 23, 23, 59, 0, 0, time.UTC), false, DoorLocked, 2026},
		{time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC), true, DoorAvailable, 1},
		{time.Date(2026, 11, 29, 23, 59, 0, 0, time.UTC), true, DoorAvailable, 1},
		{time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), false, DoorLocked, 2026},
		{time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), false, DoorAvailable, 2026},
	} {
		if got := AdventTestWindow(tt.at); got != tt.window {
			t.Errorf("window at %s = %v, want %v", tt.at, got, tt.window)
		}
		day := 24
		if tt.at.Month() == time.December {
			day = 1
		}
		if got := AdventStatus(tt.at, day, false); got != tt.status {
			t.Errorf("door %d at %s = %s, want %s", day, tt.at, got, tt.status)
		}
		if got := AdventYearKey(tt.at); got != tt.key {
			t.Errorf("year key at %s = %d, want %d", tt.at, got, tt.key)
		}
	}
	if got := AdventStatus(time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC), 7, true); got != DoorOpened {
		t.Errorf("opened door in the window = %s", got)
	}
}
