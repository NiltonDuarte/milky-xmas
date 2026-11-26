package gifttext

import (
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestLeaderboardShowsMedalsAndBars(t *testing.T) {
	stats := core.EventStats{Players: map[string]core.PlayerActivity{
		"ana": {Waterings: 105}, "bia": {Waterings: 40}, "caio": {Waterings: 22}, "duda": {Waterings: 1},
	}}
	blocks, ok := Leaderboard("plants", stats, func(userID string) string { return userID })
	want := strings.Join([]string{
		"## 🪴 Plant moms and dads", "-# Times they watered the tree",
		"🥇 ana — **105**", "🟩🟩🟩🟩🟩🟩🟩🟩",
		"🥈 bia — **40**", "🟩🟩🟩⬛⬛⬛⬛⬛",
		"🥉 caio — **22**", "🟩🟩⬛⬛⬛⬛⬛⬛",
		"`4` duda — **1**", "🟩⬛⬛⬛⬛⬛⬛⬛",
	}, "\n")
	if !ok || len(blocks) != 1 || blocks[0] != want {
		t.Errorf("leaderboard = %q, %v", blocks, ok)
	}
}
