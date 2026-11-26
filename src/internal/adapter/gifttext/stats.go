package gifttext

import (
	"fmt"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const (
	StatsCommand     = "stats"
	StatsUnavailable = "Could not read the statistics. Try again."

	StatsServerView = "server"
	StatsBoardsView = "boards"

	statsLimit = 10 // Players listed on each leaderboard.
)

// StatsChoice is one entry of a statistics menu.
type StatsChoice struct {
	Value, Label, Emoji string
}

// StatsViews are the two things /stats can show.
var StatsViews = []StatsChoice{
	{StatsServerView, "Server stats", "📊"},
	{StatsBoardsView, "Leaderboards", "🏆"},
}

// StatsBoards are the leaderboard categories.
var StatsBoards = []StatsChoice{
	{"bad", "Bad eggs", "😈"},
	{"good", "Good eggs", "😇"},
	{"throwers", "Throwers", "❄️"},
	{"gifters", "Gifters", "🎁"},
	{"plants", "Plant moms and dads", "🪴"},
}

// ServerStats is the server's totals. Good and bad deeds are left out: /spirit reveals them.
func ServerStats(stats core.EventStats) string {
	total := stats.Total
	return fmt.Sprintf("## 📊 Server stats\n🎄 Christmas tree: **%d%%**\n💧 Waterings: **%d**\n🎁 Gifts sent: **%d**\n🌑 Coal sent: **%d**\n"+
		"❄️ Snowballs thrown: **%d**\n🟡 Yellow snowballs thrown: **%d**\n📦 Drops collected: **%d**\n👽 Grinch fights: **%d** (won %d · lost %d)",
		stats.TreeLevel,
		total(func(a core.PlayerActivity) int64 { return a.Waterings }),
		total(core.PlayerActivity.GiftsSent),
		total(func(a core.PlayerActivity) int64 { return a.CoalSent }),
		total(func(a core.PlayerActivity) int64 { return a.SnowballsThrown }),
		total(func(a core.PlayerActivity) int64 { return a.YellowThrown }),
		total(func(a core.PlayerActivity) int64 { return a.DropsClaimed }),
		stats.Fights.Won+stats.Fights.Lost, stats.Fights.Won, stats.Fights.Lost)
}

// Leaderboard is the text of one category, in parts a box can show apart. It reports false for an
// unknown category. Mention formats a player's user ID.
func Leaderboard(board string, stats core.EventStats, mention func(string) string) ([]string, bool) {
	list := func(title string, value func(core.PlayerActivity) int64) string {
		var text strings.Builder
		text.WriteString(title)
		entries := stats.Ranking(value)
		for i, entry := range entries[:min(statsLimit, len(entries))] {
			fmt.Fprintf(&text, "\n%d. %s — %d", i+1, mention(entry.Member), entry.Count)
		}
		if len(entries) == 0 {
			text.WriteString("\nNobody yet.")
		}
		return text.String()
	}
	switch board {
	case "bad":
		return []string{list("## 😈 Bad eggs\nTotal bad deeds", core.PlayerActivity.BadPoints)}, true
	case "good":
		return []string{list("## 😇 Good eggs\nTotal good deeds", func(a core.PlayerActivity) int64 { return a.GoodPoints(stats.WaterPoints) })}, true
	case "throwers":
		return []string{
			"## ❄️ Throwers",
			list("**❄️ Normal snowballs**", func(a core.PlayerActivity) int64 { return a.SnowballsThrown }),
			list("**🟡 Yellow snowballs**", func(a core.PlayerActivity) int64 { return a.YellowThrown }),
		}, true
	case "gifters":
		return []string{
			"## 🎁 Gifters",
			list("**🎁 Gifts sent**", core.PlayerActivity.GiftsSent),
			list("**🌑 Coal sent**", func(a core.PlayerActivity) int64 { return a.CoalSent }),
		}, true
	case "plants":
		return []string{list("## 🪴 Plant moms and dads\nTimes they watered the tree", func(a core.PlayerActivity) int64 { return a.Waterings })}, true
	default:
		return nil, false
	}
}
