package gifttext

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

// DropCollected is the private reply after a collection.
func DropCollected(award app.GiftDropAward) string {
	if award.Coal {
		return "You collected " + CoalEmoji + " Coal."
	}
	return fmt.Sprintf("You collected a %s gift.", award.Rarity)
}

// DropRejection explains why a collection did not happen. It is empty for unexpected errors.
func DropRejection(err error) string {
	switch {
	case errors.Is(err, app.ErrDropCooldown):
		return "You cannot collect another gift yet."
	case errors.Is(err, app.ErrDropCollected):
		return "You already collected a gift from this drop."
	case errors.Is(err, app.ErrDropSlotTaken):
		return "Someone already opened that gift."
	case errors.Is(err, app.ErrDropUnavailable):
		return "This gift drop is no longer available."
	default:
		return ""
	}
}

// DropExpired is the final message of a drop whose time ran out. Mention formats a player's
// user ID.
func DropExpired(claims []app.DropClaimRecord, mention func(string) string) string {
	return dropEnd("⏰ Time's up! The Grinch stole all the presents! 😈", claims, mention)
}

// DropFinished is the final message of a drop that players used up before its time ran out.
func DropFinished(claims []app.DropClaimRecord, mention func(string) string) string {
	return dropEnd("🎉 All the gifts were collected!", claims, mention)
}

func dropEnd(title string, claims []app.DropClaimRecord, mention func(string) string) string {
	var text strings.Builder
	text.WriteString(title)
	if len(claims) == 0 {
		text.WriteString("\nNobody collected a gift.")
		return text.String()
	}
	text.WriteString("\n**Collected by**")
	for _, claim := range claims {
		prize := fmt.Sprintf("🎁 %s gift", claim.Rarity)
		if claim.Coal {
			prize = CoalEmoji + " Coal"
		}
		text.WriteString("\n" + mention(claim.UserID) + " — " + prize)
	}
	return text.String()
}
