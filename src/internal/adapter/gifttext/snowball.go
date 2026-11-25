package gifttext

import (
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func SnowballThrown(recipient string, kind core.SnowballType) string {
	return fmt.Sprintf("You have thrown a %s %s at %s. 🥶", SnowballEmoji(kind), snowballName(kind), recipient)
}

// SnowballReceived takes the recipient's mention or name, which the adapter formats.
func SnowballReceived(recipient, sender string, kind core.SnowballType) string {
	return fmt.Sprintf("%s❕  %s has thrown a %s %s at you. 🥶", recipient, sender, SnowballEmoji(kind), snowballName(kind))
}

func SnowballEmoji(kind core.SnowballType) string {
	if kind == core.YellowSnowball {
		return "🟡"
	}
	return "❄️"
}

func snowballName(kind core.SnowballType) string {
	if kind == core.YellowSnowball {
		return "yellow snowball"
	}
	return "snowball"
}

func SnowballRejection(err error) string {
	switch {
	case errors.Is(err, app.ErrInvalidRecipient):
		return "Choose another player in this server."
	case errors.Is(err, core.ErrSnowballType):
		return "Choose a normal or yellow snowball."
	case errors.Is(err, core.ErrInsufficientSnowball):
		return "You don't have a Snowball to throw."
	case errors.Is(err, core.ErrInsufficientYellowSnowball):
		return "You don't have a Yellow Snowball to throw."
	case errors.Is(err, core.ErrSnowballCapacity):
		return "The recipient's snowball inventory is full."
	default:
		return ""
	}
}
