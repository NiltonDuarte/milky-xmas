package gifttext

import (
	"fmt"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func Claim(claim app.GiftClaim) string {
	if claim.Collected {
		return fmt.Sprintf("🎅🏻 Christmas won’t be Christmas without any presents. You collected %d gifts 🎁.\nGift rarity · %s.", claim.Amount, claim.Rarity)
	}
	minutes := int((claim.Remaining + time.Minute - time.Nanosecond) / time.Minute)
	var parts []string
	if hours := minutes / 60; hours > 0 {
		parts = append(parts, units(hours, "hour"))
	}
	if rest := minutes % 60; rest > 0 {
		parts = append(parts, units(rest, "minute"))
	}
	return "🎅 you already got the gift for today, the next gift 🎁 is in " + strings.Join(parts, " ") + "."
}

func units(amount int, unit string) string {
	if amount != 1 {
		unit += "s"
	}
	return fmt.Sprintf("%d %s", amount, unit)
}
