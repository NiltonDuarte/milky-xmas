package gifttext

import (
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestAuraDrawsTheScaleWithTheNumberUnderTheMarker(t *testing.T) {
	for aura, want := range map[core.Aura]string{
		-50: "```\n😈 -50 ●▱▱▱▱|▱▱▱▱▱ +50 😇\n      -50\n```",
		-15: "```\n😈 -50 ▱▱▱●▱|▱▱▱▱▱ +50 😇\n         -15\n```",
		-1:  "```\n😈 -50 ▱▱▱▱●|▱▱▱▱▱ +50 😇\n          -1\n```",
		0:   "```\n😈 -50 ▱▱▱▱▱●▱▱▱▱▱ +50 😇\n            0\n```",
		20:  "```\n😈 -50 ▱▱▱▱▱|▱▱●▱▱ +50 😇\n              +20\n```",
		50:  "```\n😈 -50 ▱▱▱▱▱|▱▱▱▱● +50 😇\n                +50\n```",
	} {
		if got := Aura(aura); got != want {
			t.Errorf("Aura(%d) =\n%s\nwant\n%s", aura, got, want)
		}
	}
}
