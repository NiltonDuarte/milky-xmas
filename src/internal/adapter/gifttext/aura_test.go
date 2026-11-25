package gifttext

import (
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestAuraShowsANumberOnBothSides(t *testing.T) {
	for aura, want := range map[core.Aura]string{
		-50: "```\n-50 😈 ●────┼───── 😇 0\n       👆🏻\n```",
		-15: "```\n-15 😈 ───●─┼───── 😇 0\n          👆🏻\n```",
		-1:  "```\n-1 😈 ────●┼───── 😇 0\n          👆🏻\n```",
		0:   "```\n0 😈 ─────●───── 😇 0\n          👆🏻\n```",
		20:  "```\n0 😈 ─────┼──●── 😇 +20\n             👆🏻\n```",
		50:  "```\n0 😈 ─────┼────● 😇 +50\n               👆🏻\n```",
	} {
		if got := Aura(aura); got != want {
			t.Errorf("Aura(%d) =\n%q\nwant\n%q", aura, got, want)
		}
	}
}
