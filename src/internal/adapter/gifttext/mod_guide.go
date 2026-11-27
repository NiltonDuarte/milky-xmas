package gifttext

import (
	"fmt"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const (
	ModGuideCommand = "mod-guide"
	ModGuideDenied  = "You need Manage Messages or Manage Server permission to read the moderator guide."
)

// ModGuideTopics are the choices of the moderator guide's drop-down, in order.
var ModGuideTopics = []StatsChoice{
	{Value: GuideStart, Label: "Overview and commands", Emoji: "🛠️"},
	{Value: "setup", Label: "Setup and channels", Emoji: "⚙️"},
	{Value: "tools", Label: "Moderator commands", Emoji: "🧰"},
	{Value: "gifts", Label: "Daily, open and send", Emoji: "🎁"},
	{Value: "drops", Label: "Gift drops", Emoji: "📦"},
	{Value: "aura", Label: "Aura", Emoji: "😇"},
	{Value: "grinch", Label: "Grinch: how a fight runs", Emoji: "👽"},
	{Value: "fight", Label: "Grinch: numbers", Emoji: "⚔️"},
	{Value: "world", Label: "Tree, Advent, snowballs, quotes", Emoji: "🌲"},
	{Value: "records", Label: "Stats, data and changing numbers", Emoji: "📊"},
}

// ModGuide explains one topic of the bot to a moderator, with the numbers in use. It reports false
// for an unknown topic.
func ModGuide(topic string, tuning core.Tuning) ([]string, bool) {
	switch topic {
	case GuideStart:
		return modOverview(), true
	case "setup":
		return modSetup(), true
	case "tools":
		return modTools(), true
	case "gifts":
		return modGifts(tuning), true
	case "drops":
		return modDrops(tuning), true
	case "aura":
		return modAura(tuning), true
	case "grinch":
		return modGrinch(tuning), true
	case "fight":
		return modFight(tuning), true
	case "world":
		return modWorld(tuning), true
	case "records":
		return modRecords(), true
	}
	return nil, false
}

func modOverview() []string {
	return []string{
		"## 🛠️ Moderator guide\nHow the whole bot works and the numbers it is using.\n-# Pick a topic in the menu. Only you see this.",
		"### 👥 Everyone\n" +
			"`/daily claim` · `/look in-your-stocking` · `/send gift` · `/open gift` · `/aura check` · `/throw snowball` · " +
			"`/advent view` · `/water` · `/stats` · `/guide` · `/help me-milky`",
		"### 🧰 Manage Messages\n`/give gift` · `/grinch summon` · `/spirit` · `/mod-guide`\n" +
			"### ⚙️ Manage Server\n`/setup view` · `/setup channel`\n" +
			"-# Administrators can use all of them.",
		"### 🤖 On its own\nGift drops, Grinch visits, Christmas quotes, the tree losing water and the Advent doors unlocking.",
	}
}

func modSetup() []string {
	return []string{
		"## ⚙️ Setup and channels\n**`/setup view`** lists the chosen channels.\n**`/setup channel`** sets one text channel:",
		"• **commands**: commands work there, and gift drops can appear\n" +
			"• **boss**: the Grinch can appear there\n" +
			"• **quotes** (optional): Christmas quotes are posted there; leaving it out keeps the choice\n" +
			"-# Only `/setup` works in a channel with commands off. To use `/grinch summon`, the channel needs commands **and** boss.",
		"### 🔑 What the bot needs\n" +
			"• View Channel and Send Messages\n" +
			"• Attach Files: pictures and fight standings\n" +
			"• Manage Roles, with its role above `elfs`: elf army button",
		"-# The Christmas introduction and its elf army button are turned off for now. Turning off boss stops the Grinch in that channel.",
	}
}

func modTools() []string {
	return []string{
		"## 🧰 Moderator commands\n**`/give gift <player> <amount> <rarity>`**\n" +
			"Adds Normal, Rare or Epic gifts. The reply is private. It changes no aura and no daily wait. " +
			"You cannot give to yourself or to a bot.",
		"**`/grinch summon`**\n" +
			"Calls the Grinch to the boss channel you are in, right away: no chat activity, chance or rest time needed. " +
			"Refused while a fight is already running there.",
		"**`/spirit`**\n" +
			"Posts the server's Christmas spirit **for everyone to see**: the good auras added up against the bad ones, " +
			"as they are at that moment. Meant for the end of the event.",
	}
}

func modGifts(tuning core.Tuning) []string {
	odds, aura := tuning.Gifts.DailyRarity, tuning.Aura
	return []string{
		fmt.Sprintf("## 🎁 Daily, open and send\n**`/daily claim`**: once every %s, 1 to %d gifts (each amount equally likely), all of one rarity: "+
			"Normal %d%% · Rare %d%% · Epic %d%%.\n"+
			"Nice list: +%d gift. Naughty list: +%d 🟡. Aura 0 or below: +%d %s Coal.",
			interval(core.GiftCooldown), core.MaxDailyGifts, odds.Normal, odds.Rare, odds.Epic,
			aura.Nice.DailyExtraGifts, aura.Naughty.DailyYellowSnowballs, aura.DailyCoal, CoalEmoji),
		fmt.Sprintf("### 🎀 `/open gift <rarity>`\n"+
			"Takes 1 gift. Each surprise of the rarity is equally likely, and the aura does not change.\n"+
			"**Normal**: 1 Snowball, 1 Coal or a quote\n"+
			"**Rare**: %d Snowballs or an extra watering\n"+
			"**Epic**: a shield or %d normal gifts\n"+
			"-# Only one shield and one extra watering are kept: with one already, the other surprise comes.",
			core.SurpriseSnowballCount, core.SurpriseGiftCount),
		fmt.Sprintf("### 📬 `/send gift <player> <item> <n>`\n"+
			"Normal, Rare, Epic or Coal from your own stocking, to another player (not a bot).\n"+
			"Aura for each one: Normal +%d · Rare +%d · Epic +%d · Coal -%d\n"+
			"-# The aura limits are in the Aura topic. The items move even when the aura does not.",
			core.NormalGiftPoints, core.RareGiftPoints, core.EpicGiftPoints, core.CoalPoints),
		"**`/look in-your-stocking`**: gifts, Snowballs, Coal, shield and extra watering.",
	}
}

func modDrops(tuning core.Tuning) []string {
	drops, rain, aura := tuning.Drops, tuning.Drops.GiftRain, tuning.Aura
	daily, boosted := tuning.Gifts.DailyRarity, tuning.Gifts.BoostedRarity
	return []string{
		fmt.Sprintf("## 📦 Gift drops\nThey come from players' messages in channels with commands on. A drop stays open for %s.", span(drops.Lifetime)),
		fmt.Sprintf("### 🎁 Random Active Drop\n"+
			"1 in %d chance per message. %d boxes, one per player, each box opens once.\n"+
			"Odds: Normal %d%% · Rare %d%% · Epic %d%%\n"+
			"Nice list: Normal %d%% · Rare %d%% · Epic %d%%",
			drops.RandomActiveOneIn, core.GiftDropSlots, daily.Normal, daily.Rare, daily.Epic, boosted.Normal, boosted.Rare, boosted.Epic),
		fmt.Sprintf("### 🌧️ Gift Rain\n"+
			"Comes after %d to %d messages in a channel, drawn again each time. Ends after %d collections.\n"+
			"Each collection: %d%% Coal, otherwise a gift with the odds Normal %d%% · Rare %d%% · Epic %d%%.\n"+
			"Wait between collections: %s. Naughty list: %s, and %d%% chance of a double gift.",
			rain.MinMessages, rain.MaxMessages, rain.ClaimsToEnd, rain.CoalPercent, daily.Normal, daily.Rare, daily.Epic,
			span(rain.ClaimCooldown), span(min(rain.ClaimCooldown, aura.Naughty.RainClaimCooldown)), aura.Naughty.RainDoublePercent),
	}
}

func modAura(tuning core.Tuning) []string {
	aura, tree, grinch := tuning.Aura, tuning.Tree, tuning.Grinch
	limits := fmt.Sprintf("## 😇 Aura\nGoes from %d to +%d. `/aura check` shows only the player's own.\n"+
		"**Up:** gifts sent (+%d/+%d/+%d), watering (+%d, only from an aura of 0 up)\n"+
		"**Down:** each Coal sent (-%d), each 🟡 thrown (-%d)",
		core.AuraMin, core.AuraMax, core.NormalGiftPoints, core.RareGiftPoints, core.EpicGiftPoints, tree.WaterAura,
		core.CoalPoints, core.YellowSnowballPoints)
	if aura.SendCooldown > 0 {
		limits += fmt.Sprintf("\n• Sends to the same player count once every %s in each direction", interval(aura.SendCooldown))
	}
	if aura.DailyLimit > 0 {
		limits += fmt.Sprintf("\n• At most %d points up and %d down per UTC day", aura.DailyLimit, aura.DailyLimit)
	}
	return []string{
		limits,
		fmt.Sprintf("### 😇 Nice list (+%d or more)\n"+
			"+%d daily gift · watering +%d%% · better Random Active Drops · +%d%% Grinch damage · a hit costs %d%% HP",
			aura.NiceAt, aura.Nice.DailyExtraGifts, aura.Nice.WaterExtraPoints, aura.Nice.BossDamageBonusPercent, aura.Nice.MissHPLossPercent),
		fmt.Sprintf("### 😈 Naughty list (%d or less)\n"+
			"+%d daily 🟡 · %d%% double Gift Rain gift · %s between rain gifts · a hit costs %d%% HP · watering -%d%% (at least 1%%) · %s when the Grinch wins",
			aura.NaughtyAt, aura.Naughty.DailyYellowSnowballs, aura.Naughty.RainDoublePercent,
			span(min(tuning.Drops.GiftRain.ClaimCooldown, aura.Naughty.RainClaimCooldown)), aura.Naughty.MissHPLossPercent,
			aura.Naughty.WaterLessPoints, coalReward(aura.Naughty.LossCoal)),
		fmt.Sprintf("### 😐 In between\nA hit costs %d%% HP, normal damage, %s on a loss.\n"+
			"-# Aura 0 or below: +%d Coal with the daily claim. A fight uses the aura a player had on entering it.",
			grinch.DefenseMissHPLoss, coalReward(grinch.LossCoal), aura.DailyCoal),
	}
}

func modGrinch(tuning core.Tuning) []string {
	grinch := tuning.Grinch
	return []string{
		fmt.Sprintf("## 👽 How he appears\nEvery %s, each boss channel has a %d%% chance, if a player spoke there in the last %s, "+
			"no fight is running, and %s passed since the last one ended. `/grinch summon` skips all but the running fight.",
			interval(grinch.SpawnCheckInterval), grinch.SpawnChancePercent, span(grinch.ActivityWindow), span(grinch.CooldownAfterFight)),
		fmt.Sprintf("### 🎅 Joining\n"+
			"The announcement waits %s for a first **Start**; with none the fight is cancelled. "+
			"The countdown is then %s, or the fight begins at once with %d players. "+
			"Anyone can still join later just by acting.",
			span(grinch.StartWindow), span(grinch.RegistrationDuration), grinch.EarlyStartPlayers),
		fmt.Sprintf("### ⏳ Phases and clocks\n"+
			"1. Snowballs, ☃️: %s to bring him from 100 to %s HP\n"+
			"2. Defense, 🎄 is safe and ❄️ hits: down to %s HP\n"+
			"3. Attack, ☃️ or 🧊: down to %s HP\n"+
			"4. Final, ⚔️ 🎁 ❄️: down to 0\n"+
			"Phases 2 to 4 share one clock of %s. If a clock runs out, he escapes.",
			span(grinch.SnowballsDuration), core.FormatHundredths(grinch.SnowballsBossHP), core.FormatHundredths(grinch.DefenseBossHP),
			core.FormatHundredths(grinch.AttackBossHP), span(grinch.FightDuration)),
		fmt.Sprintf("### 🏁 Endings\n"+
			"Won, escaped, or everyone down. On the last two, each player in the fight gets %d %s Coal and %d 🟡 per player in it "+
			"(Naughty list: %s). A cancelled fight gives nothing.",
			grinch.LossCoal, CoalEmoji, grinch.LossYellowPerGroupMember, coalReward(tuning.Aura.Naughty.LossCoal)),
	}
}

func modFight(tuning core.Tuning) []string {
	grinch := tuning.Grinch
	var sizes []string
	for _, players := range []int{1, 3, 5, 10} {
		hits := "fixed damage"
		if needed := max(grinch.MinHits, grinch.HitsPerPlayer*players); needed > 0 {
			hits = fmt.Sprintf("%d hits", needed)
		}
		sizes = append(sizes, fmt.Sprintf("%d: %s, %d defenses", players, hits, grinch.DefensesNeeded(players)))
	}
	need := "## ⚔️ Grinch numbers\n"
	if grinch.HitsPerPlayer > 0 || grinch.MinHits > 0 {
		need += fmt.Sprintf("Each attack phase needs %d hits per player, at least %d. ", grinch.HitsPerPlayer, grinch.MinHits)
	}
	need += fmt.Sprintf("Defense needs %d safe picks per player, from %d to %d.\n**Players → per phase**\n%s",
		grinch.DefensePerPlayer, grinch.DefenseMinimum, grinch.DefenseMaxDefenses, strings.Join(sizes, "\n"))

	buttons := fmt.Sprintf("### 🎮 Buttons\n"+
		"🧊 hits about %d%% harder than ☃️.\n"+
		"Final row: ⚔️ attacks, 🎁 gives %d to %d gift points, ❄️ gives %d Snowball. ❄️ moves after every pick.",
		(grinch.AttackHeavyDamage*100+grinch.AttackDamage/2)/max(1, grinch.AttackDamage)-100,
		grinch.GiftPointsMin, grinch.GiftPointsMax, grinch.FinalSnowballReward)
	if grinch.ClickCooldown > 0 {
		buttons += fmt.Sprintf("\nWait between a player's moves: %s.", span(grinch.ClickCooldown))
	}

	health := fmt.Sprintf("### ❤️ HP\nEveryone has 100%%. A ❄️ in the defense costs %d%% (Nice %d%%, Naughty %d%%). "+
		"Any hit on the Grinch has a %d%% chance of a counterattack for the same cost. A shield blocks one of either.",
		grinch.DefenseMissHPLoss, tuning.Aura.Nice.MissHPLossPercent, tuning.Aura.Naughty.MissHPLossPercent, grinch.CounterattackPercent)
	revival := "### 💀 Coming back\nNobody can come back."
	if grinch.CanRevive() {
		revival = fmt.Sprintf("### 💀 Coming back\nA gift gives back: Normal %d%% · Rare %d%% · Epic %d%% HP (0%% = not usable).\n"+
			"A fallen player spends their own. A standing player's gift raises whoever fell first. "+
			"With everyone down there are %s left.",
			grinch.ReviveNormalHP, grinch.ReviveRareHP, grinch.ReviveEpicHP, span(grinch.LastChance))
	}
	return []string{need, buttons, health, revival, "-# The end shows damage, gift and Snowball leaderboards, with the full standings attached."}
}

func modWorld(tuning core.Tuning) []string {
	tree, quotes := tuning.Tree, tuning.Quotes
	return []string{
		fmt.Sprintf("## 🌲 Christmas tree\nOne per server. Starts at %d%%, loses 1%% every %s. `/water` adds %d%% once every %s per player, "+
			"and is refused at 100%%. It stops on %s UTC.",
			tree.StartLevel, interval(tree.DecayInterval), tree.WaterPoints, interval(tree.WaterCooldown),
			tree.EventEnd.UTC().Format("January 2, 2006 at 15:04")),
		fmt.Sprintf("### 📅 Advent calendar\n%d doors, shuffled for each player. One unlocks each day from December 1 at 00:00 UTC; "+
			"missed ones stay open until the end of December 24 UTC.", core.AdventDays),
		"### ❄️ Snowballs\n`/throw snowball` moves one to the other player and posts in the channel. " +
			"A normal one changes no aura; a 🟡 is a bad deed.",
		fmt.Sprintf("### 💬 Quotes\nIn channels with quotes on: one after %d player messages, at most one every %s per channel. "+
			"None repeats until all were shown.", quotes.EveryMessages, interval(quotes.MinInterval)),
	}
}

func modRecords() []string {
	return []string{
		"## 📊 Stats and leaderboards\n`/stats` is private and open to everyone: Milky Way Galaxy totals, and Bad eggs, Good eggs, " +
			"Throwers, Gifters and Plant moms and dads.\n-# Bad and Good eggs add up deeds without cancelling each other and without the aura's limits.",
		"### 💾 What is saved\nInventories, aura, waits, tree, opened doors, channel choices, fights and the history of player actions.",
		"### 🔁 If the bot restarts\nOpen gift drops stop working, and the message counts for rains and quotes start over.",
		"### 🔧 Changing numbers\nNot possible from Discord. Whoever runs the bot edits `tuning.yaml` and restarts it. " +
			"This guide and `/guide` then show the new numbers.",
	}
}
