package gifttext

import (
	"fmt"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func Claim(claim app.GiftClaim) string {
	if claim.Collected {
		text := fmt.Sprintf("## 🎅🏻 Christmas won’t be Christmas without any presents.\n```\nYou collected %d gifts 🎁.\nGift rarity · %s ·\n```", claim.Amount, claim.Rarity)
		if claim.Bonus.Gifts > 0 {
			text += fmt.Sprintf("\n😇 Nice list: %d of them extra.", claim.Bonus.Gifts)
		}
		if claim.Bonus.YellowSnowballs > 0 {
			text += fmt.Sprintf("\n😈 Naughty list: +%d 🟡 yellow snowball%s.", claim.Bonus.YellowSnowballs, plural(claim.Bonus.YellowSnowballs))
		}
		if claim.Bonus.Coal > 0 {
			text += fmt.Sprintf("\n%s +%d Coal.", CoalEmoji, claim.Bonus.Coal)
		}
		return text
	}
	minutes := int((claim.Remaining + time.Minute - time.Nanosecond) / time.Minute)
	var parts []string
	if hours := minutes / 60; hours > 0 {
		parts = append(parts, units(hours, "hour"))
	}
	if rest := minutes % 60; rest > 0 {
		parts = append(parts, units(rest, "minute"))
	}
	return "🎅 You already got the gift for today, the next gift 🎁 is in " + strings.Join(parts, " ") + "."
}

func plural(amount int64) string {
	if amount == 1 {
		return ""
	}
	return "s"
}

func units(amount int, unit string) string {
	if amount != 1 {
		unit += "s"
	}
	return fmt.Sprintf("%d %s", amount, unit)
}
