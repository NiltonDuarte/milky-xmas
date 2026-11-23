package gifttext

import (
	"errors"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func SnowballThrown(recipient string) string {
	return fmt.Sprintf("You have thrown a ❄️ snowball at %s. 🥶", recipient)
}

func SnowballReceived(sender string, total int64) string {
	item := "snowballs"
	if total == 1 {
		item = "snowball"
	}
	return fmt.Sprintf("%s have threw a ❄️ snowball at you, now you have %d ❄️ %s. 🥶", sender, total, item)
}

func SnowballRejection(err error) string {
	switch {
	case errors.Is(err, app.ErrInvalidRecipient):
		return "Choose another player in this server."
	case errors.Is(err, core.ErrSnowballCapacity):
		return "The recipient's snowball inventory is full."
	default:
		return ""
	}
}
