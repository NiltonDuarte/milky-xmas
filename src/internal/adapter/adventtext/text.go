package adventtext

import (
	"fmt"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func Calendar(calendar core.AdventCalendar) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Advent %d · December 1–24 UTC · Advent gifts: %d", calendar.Year, calendar.Total)
	for i, status := range calendar.Doors {
		fmt.Fprintf(&text, "\n%02d · %s", i+1, status)
	}
	return text.String()
}

func Opening(day int, opening app.AdventOpening) string {
	switch opening.Status {
	case core.DoorInvalid:
		return "Choose an Advent door from 1 to 24."
	case core.DoorLocked:
		return fmt.Sprintf("Door %d unlocks on December %d at 00:00 UTC.", day, day)
	case core.DoorExpired:
		return "This Advent calendar closed at the end of December 24 UTC."
	case core.DoorOpened:
		if opening.Awarded {
			return fmt.Sprintf("🎁 You opened door %d! Advent gifts: %d.", day, opening.Total)
		}
		return fmt.Sprintf("Door %d is already open. Advent gifts: %d.", day, opening.Total)
	default:
		return "Could not open the Advent door. Try again."
	}
}
