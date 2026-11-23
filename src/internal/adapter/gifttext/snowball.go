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

func SnowballReceived(sender string, total int64, kind core.SnowballType) string {
	item := snowballName(kind)
	plural := item
	if total != 1 {
		plural += "s"
	}
	emoji := SnowballEmoji(kind)
	return fmt.Sprintf("%s have threw a %s %s at you, now you have %d %s %s. 🥶", sender, emoji, item, total, emoji, plural)
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
	case errors.Is(err, core.ErrInsufficientYellowSnowball):
		return "You don't have a Yellow Snowball to throw."
	case errors.Is(err, core.ErrSnowballCapacity):
		return "The recipient's snowball inventory is full."
	default:
		return ""
	}
}
