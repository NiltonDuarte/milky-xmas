package gifttext

import (
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const (
	SpiritCommand        = "spirit"
	SpiritNeedsModerator = "Only moderators can reveal the Christmas spirit."
	SpiritUnavailable    = "Could not read the Christmas spirit. Try again."
)

// SpiritBlocks is the server's Christmas spirit in two parts: the totals and the verdict.
func SpiritBlocks(spirit core.Spirit) []string {
	verdict := "⚖️ Good and bad deeds are in perfect balance."
	switch {
	case spirit.Good > spirit.Bad:
		verdict = "✨ The good deeds win! The Christmas spirit shines bright."
	case spirit.Bad > spirit.Good:
		verdict = "🌑 The bad deeds win... The Grinch is very pleased."
	}
	return []string{
		fmt.Sprintf("## 🎄 The Christmas spirit of this server\n😇 Good deeds: **%d**\n😈 Bad deeds: **%d**", spirit.Good, spirit.Bad),
		verdict,
	}
}
