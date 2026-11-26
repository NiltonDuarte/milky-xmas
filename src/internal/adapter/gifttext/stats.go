package gifttext

import (
	"fmt"
	"slices"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const (
	StatsCommand     = "stats"
	StatsUnavailable = "Could not read the statistics. Try again."

	StatsServerView = "server"
	StatsBoardsView = "boards"

	statsLimit     = 10 // Players listed on each leaderboard.
	statsBarLength = 8  // Squares in the bar of the first place.
	statsBarEmpty  = "⬛"
)

// statsMedals mark the first three places.
var statsMedals = [...]string{"🥇", "🥈", "🥉"}

// StatsChoice is one entry of a statistics menu.
type StatsChoice struct {
	Value, Label, Emoji string
}

// statsGalaxy is the server's totals, as the menus name them.
var statsGalaxy = StatsChoice{StatsServerView, "Milky Way Galaxy", "🌌"}

// StatsViews are the two things /stats can show.
var StatsViews = []StatsChoice{
	{StatsBoardsView, "Leaderboards", "🏆"},
	statsGalaxy,
}

// StatsBoards are the leaderboard categories.
var StatsBoards = []StatsChoice{
	{"bad", "Bad eggs", "😈"},
	{"good", "Good eggs", "😇"},
	{"throwers", "Throwers", "❄️"},
	{"gifters", "Gifters", "🎁"},
	{"plants", "Plant moms and dads", "🪴"},
}

// StatsPanels are what the second menu can show: a leaderboard or the server's totals.
var StatsPanels = append(slices.Clone(StatsBoards), statsGalaxy)

// ServerStats is the server's totals, in parts a box can show apart. Good and bad deeds are left
// out: /spirit reveals them.
func ServerStats(stats core.EventStats) []string {
	total := stats.Total
	tree := min(10, max(0, stats.TreeLevel/10))
	return []string{
		"## 🌌 Milky Way Galaxy\n-# Everything that happened in this server",
		fmt.Sprintf("### 🎄 Christmas tree\n%s%s **%d%%**\n💧 Waterings: **%d**",
			strings.Repeat("🟩", tree), strings.Repeat(statsBarEmpty, 10-tree), stats.TreeLevel,
			total(func(a core.PlayerActivity) int64 { return a.Waterings })),
		fmt.Sprintf("### 🎁 Presents\n🎁 Gifts sent: **%d**\n🌑 Coal sent: **%d**\n📦 Drops collected: **%d**",
			total(core.PlayerActivity.GiftsSent),
			total(func(a core.PlayerActivity) int64 { return a.CoalSent }),
			total(func(a core.PlayerActivity) int64 { return a.DropsClaimed })),
		fmt.Sprintf("### ❄️ Snowball war\n❄️ Snowballs thrown: **%d**\n🟡 Yellow snowballs thrown: **%d**",
			total(func(a core.PlayerActivity) int64 { return a.SnowballsThrown }),
			total(func(a core.PlayerActivity) int64 { return a.YellowThrown })),
		fmt.Sprintf("### 👽 Grinch fights: %d\n🏆 Won: **%d**\n💀 Lost: **%d**",
			stats.Fights.Won+stats.Fights.Lost, stats.Fights.Won, stats.Fights.Lost),
	}
}

// statsPlace is a medal for the first three places and the number for the others.
func statsPlace(index int) string {
	if index < len(statsMedals) {
		return statsMedals[index]
	}
	return fmt.Sprintf("`%d`", index+1)
}

// statsBar compares a count with the first place's, rounded to whole squares and never empty.
func statsBar(square string, count, top int64) string {
	filled := max(1, int((2*statsBarLength*count+top)/(2*top)))
	return strings.Repeat(square, filled) + strings.Repeat(statsBarEmpty, statsBarLength-filled)
}

// Leaderboard is the text of one category, in parts a box can show apart. It reports false for an
// unknown category. Mention formats a player's user ID.
func Leaderboard(board string, stats core.EventStats, mention func(string) string) ([]string, bool) {
	list := func(title, square string, value func(core.PlayerActivity) int64) string {
		var text strings.Builder
		text.WriteString(title)
		entries := stats.Ranking(value)
		for i, entry := range entries[:min(statsLimit, len(entries))] {
			// The ranking is highest first, so the first entry is the top.
			fmt.Fprintf(&text, "\n%s %s — **%d**\n%s", statsPlace(i), mention(entry.Member), entry.Count, statsBar(square, entry.Count, entries[0].Count))
		}
		if len(entries) == 0 {
			text.WriteString("\nNobody yet.")
		}
		return text.String()
	}
	switch board {
	case "bad":
		return []string{list("## 😈 Bad eggs\n-# Total bad deeds", "🟥", core.PlayerActivity.BadPoints)}, true
	case "good":
		return []string{list("## 😇 Good eggs\n-# Total good deeds", "🟩", func(a core.PlayerActivity) int64 { return a.GoodPoints(stats.WaterPoints) })}, true
	case "throwers":
		return []string{
			"## ❄️ Throwers",
			list("**❄️ Normal snowballs**", "🟦", func(a core.PlayerActivity) int64 { return a.SnowballsThrown }),
			list("**🟡 Yellow snowballs**", "🟨", func(a core.PlayerActivity) int64 { return a.YellowThrown }),
		}, true
	case "gifters":
		return []string{
			"## 🎁 Gifters",
			list("**🎁 Gifts sent**", "🟪", core.PlayerActivity.GiftsSent),
			list("**🌑 Coal sent**", "🟫", func(a core.PlayerActivity) int64 { return a.CoalSent }),
		}, true
	case "plants":
		return []string{list("## 🪴 Plant moms and dads\n-# Times they watered the tree", "🟩", func(a core.PlayerActivity) int64 { return a.Waterings })}, true
	default:
		return nil, false
	}
}
