package adventtext

import (
	"fmt"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// Calendar lists the doors in the player's own order, for the local chat.
func Calendar(calendar core.AdventCalendar) string {
	var text strings.Builder
	text.WriteString(Header(calendar))
	for _, door := range calendar.Order {
		fmt.Fprintf(&text, "\n%02d · %s", door+1, calendar.Doors[door])
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
		return fmt.Sprintf("**🎁 Door %d**\n\n%s", day, Content(opening.Item))
	default:
		return "Could not open the Advent door. Try again."
	}
}

// Header is the line above the Advent buttons.
func Header(calendar core.AdventCalendar) string {
	return fmt.Sprintf("🎄 Advent %d · December 1–24 UTC", calendar.Year)
}
