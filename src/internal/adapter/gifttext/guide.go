package gifttext

import (
	"fmt"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const GuideCommand = "guide"

// GuideStart is the topic a guide opens on.
const GuideStart = "start"

// GuideTopics are the choices of the guide's drop-down, in order.
var GuideTopics = []StatsChoice{
	{Value: GuideStart, Label: "Start here", Emoji: "🎄"},
	{Value: "gifts", Label: "Gifts and stocking", Emoji: "🎁"},
	{Value: "drops", Label: "Gift drops", Emoji: "📦"},
	{Value: "aura", Label: "Aura: Nice and Naughty", Emoji: "😇"},
	{Value: "snowballs", Label: "Snowball war", Emoji: "❄️"},
	{Value: "grinch", Label: "Grinch fight", Emoji: "👽"},
	{Value: "tree", Label: "Christmas tree", Emoji: "🌲"},
	{Value: "advent", Label: "Advent calendar", Emoji: "📅"},
	{Value: "more", Label: "Stats and more", Emoji: "📊"},
}

// Guide explains one topic of the event with the numbers in use. It reports false for an unknown
// topic.
func Guide(topic string, tuning core.Tuning) ([]string, bool) {
	switch topic {
	case GuideStart:
		return guideStart(), true
	case "gifts":
		return guideGifts(tuning), true
	case "drops":
		return guideDrops(tuning), true
	case "aura":
		return guideAura(tuning), true
	case "snowballs":
		return guideSnowballs(tuning), true
	case "grinch":
		return guideGrinch(tuning), true
	case "tree":
		return guideTree(tuning), true
	case "advent":
		return guideAdvent(), true
	case "more":
		return guideMore(), true
	}
	return nil, false
}

func guideStart() []string {
	return []string{
		"## 🎄 Milky Xmas guide\nCollect gifts, help the tree, choose between good and bad deeds, and stop the Grinch.\n-# Pick a topic in the menu to read more.",
		"### ☀️ Every day\n" +
			"🎁 `/daily claim` your presents\n" +
			"💧 `/water` the Christmas tree\n" +
			"📅 `/advent view` and open the day's door\n" +
			"💬 Chat, and watch for falling gifts and the Grinch",
	}
}

func guideGifts(tuning core.Tuning) []string {
	odds := tuning.Gifts.DailyRarity
	return []string{
		fmt.Sprintf("## 🎁 Gifts and stocking\n"+
			"**`/daily claim`**: 1 to %d gifts, once every %s. They are all of one rarity: Normal %d%%, Rare %d%%, Epic %d%%.\n"+
			"**`/look in-your-stocking`**: your gifts, Snowballs, Coal, shield and extra watering.",
			core.MaxDailyGifts, interval(core.GiftCooldown), odds.Normal, odds.Rare, odds.Epic),
		fmt.Sprintf("### 📬 `/send gift`\n"+
			"Send gifts or Coal to another player.\n"+
			"Good deeds for each gift: Normal +%d · Rare +%d · Epic +%d\n"+
			"Bad deeds for each %s Coal: +%d",
			core.NormalGiftPoints, core.RareGiftPoints, core.EpicGiftPoints, CoalEmoji, core.CoalPoints),
		fmt.Sprintf("### 🎀 `/open gift`\n"+
			"Spend one gift for a surprise.\n"+
			"**Normal**: 1 Snowball, 1 Coal or a Christmas quote\n"+
			"**Rare**: %d Snowballs or an extra watering\n"+
			"**Epic**: a Grinch shield or %d normal gifts\n"+
			"-# You can keep one shield and one extra watering at a time.",
			core.SurpriseSnowballCount, core.SurpriseGiftCount),
		"-# Keep some gifts: they can bring you or a friend back in a Grinch fight.",
	}
}

func guideDrops(tuning core.Tuning) []string {
	drops, rain, aura := tuning.Drops, tuning.Drops.GiftRain, tuning.Aura
	return []string{
		fmt.Sprintf("## 📦 Gift drops\nDrops appear while people chat. Press a button before they are gone: a drop stays open for %s.", span(drops.Lifetime)),
		fmt.Sprintf("### 🎁 Falling gifts\n"+
			"A 1 in %d chance with each message. Pick one of the %d boxes. One gift per player, and each box opens once.",
			drops.RandomActiveOneIn, core.GiftDropSlots),
		fmt.Sprintf("### 🌧️ Gift Rain\n"+
			"Comes after %d to %d messages. You can collect again every %s, until %d prizes are taken in total. "+
			"%d%% of the time you get %s Coal instead of a gift.",
			rain.MinMessages, rain.MaxMessages, interval(rain.ClaimCooldown), rain.ClaimsToEnd, rain.CoalPercent, CoalEmoji),
		fmt.Sprintf("-# 😇 The Nice list finds better falling gifts. 😈 The Naughty list waits only %s in a Gift Rain and has a %d%% chance of a double gift.",
			span(min(rain.ClaimCooldown, aura.Naughty.RainClaimCooldown)), aura.Naughty.RainDoublePercent),
	}
}

func guideAura(tuning core.Tuning) []string {
	aura, rain := tuning.Aura, tuning.Drops.GiftRain
	rules := fmt.Sprintf("## 😇 Aura 😈\nYour aura goes from %d (bad deeds) to +%d (good deeds). See it with **`/aura check`**.\n", core.AuraMin, core.AuraMax) +
		"**Up:** sending gifts, watering the tree\n" +
		"**Down:** sending Coal, throwing a 🟡 yellow snowball\n"
	if aura.SendCooldown > 0 {
		rules += fmt.Sprintf("-# Sends to the same player count once every %s in each direction.\n", interval(aura.SendCooldown))
	}
	if aura.DailyLimit > 0 {
		rules += fmt.Sprintf("-# It moves at most %d points up and %d down a day (UTC).\n", aura.DailyLimit, aura.DailyLimit)
	}
	rules += "-# Watering counts for your aura only from 0 up."
	return []string{
		rules,
		fmt.Sprintf("### 😇 Nice list (+%d or more)\n"+
			"• +%d gift on `/daily claim`\n"+
			"• Your watering adds %d%% more water\n"+
			"• Better odds on falling gifts\n"+
			"• +%d%% damage on the Grinch\n"+
			"• But a hit in the fight costs %d%% HP",
			aura.NiceAt, aura.Nice.DailyExtraGifts, aura.Nice.WaterExtraPoints, aura.Nice.BossDamageBonusPercent, aura.Nice.MissHPLossPercent),
		fmt.Sprintf("### 😈 Naughty list (%d or less)\n"+
			"• +%d 🟡 yellow snowball on `/daily claim`\n"+
			"• %d%% chance of a double Gift Rain gift, and only %s between them\n"+
			"• A hit in the fight costs just %d%% HP\n"+
			"• But your watering adds %d%% less water\n"+
			"• And %s when the Grinch wins",
			aura.NaughtyAt, aura.Naughty.DailyYellowSnowballs, aura.Naughty.RainDoublePercent,
			span(min(rain.ClaimCooldown, aura.Naughty.RainClaimCooldown)), aura.Naughty.MissHPLossPercent,
			aura.Naughty.WaterLessPoints, coalReward(aura.Naughty.LossCoal)),
		fmt.Sprintf("-# In between: a hit costs %d%% HP. With an aura of 0 or less, `/daily claim` also gives %d %s Coal.",
			tuning.Grinch.DefenseMissHPLoss, aura.DailyCoal, CoalEmoji),
	}
}

func guideSnowballs(tuning core.Tuning) []string {
	return []string{
		"## ❄️ Snowball war\n**`/throw snowball`** at another player. They are told in the channel, and the snowball goes to their stocking.",
		"❄️ **Normal snowball**: just for fun, your aura stays the same\n" +
			"🟡 **Yellow snowball**: a bad deed",
		fmt.Sprintf("### 🧺 Where to get them\n"+
			"❄️ `/open gift` (Normal or Rare) and the final phase of the Grinch fight\n"+
			"🟡 Losing to the Grinch, and the daily claim of the Naughty list\n"+
			"-# A lost fight gives %d 🟡 for each player who joined it.", tuning.Grinch.LossYellowPerGroupMember),
	}
}

func guideGrinch(tuning core.Tuning) []string {
	grinch := tuning.Grinch
	join := fmt.Sprintf("## 👽 Grinch fight\nHe shows up while people chat in his channel. Press **🎅 Start** to join: the fight begins %s after the first press, or at once with %d players.",
		span(grinch.RegistrationDuration), grinch.EarlyStartPlayers)
	phases := fmt.Sprintf("### ⚔️ Phases\n"+
		"**1. Snowballs**: press ☃️ and bring him down to %s HP within %s\n"+
		"**2. Defense**: press a 🎄 to hide. A ❄️ hits you\n"+
		"**3. Attack**: press ☃️, or 🧊 to hit harder\n"+
		"**4. Final**: ⚔️ attacks, 🎁 gives gift points, ❄️ gives a Snowball\n"+
		"-# After phase 1 he escapes if he is still standing after %s.",
		core.FormatHundredths(grinch.SnowballsBossHP), span(grinch.SnowballsDuration), span(grinch.FightDuration))
	health := fmt.Sprintf("### ❤️ Your HP\nYou start with 100%%. A ❄️ in the defense costs %d%% HP.", grinch.DefenseMissHPLoss)
	if grinch.CounterattackPercent > 0 {
		health += fmt.Sprintf(" When you attack, there is a %d%% chance he hits you back.", grinch.CounterattackPercent)
	}
	health += " Your aura changes these numbers.\n🛡️ A shield from an Epic gift blocks the next hit."
	if grinch.CanRevive() {
		var gifts []string
		for _, gift := range []struct {
			name string
			hp   int
		}{{"Normal", grinch.ReviveNormalHP}, {"Rare", grinch.ReviveRareHP}, {"Epic", grinch.ReviveEpicHP}} {
			if gift.hp > 0 {
				gifts = append(gifts, fmt.Sprintf("%s %d%% HP", gift.name, gift.hp))
			}
		}
		health += "\n💀 At 0 HP, spend a gift to get back up, or a friend can spend theirs on you: " + strings.Join(gifts, " · ") + "."
		if grinch.LastChance > 0 {
			health += fmt.Sprintf(" If everyone is down, you have %s.", span(grinch.LastChance))
		}
	}
	return []string{join, phases, health, fmt.Sprintf("### 🏆 Rewards\n"+
		"❄️ Snowballs you pick in the final phase\n"+
		"If he wins or escapes: %d %s Coal and %d 🟡 for each player who joined\n"+
		"-# The end shows the damage, gift and Snowball leaderboards.",
		grinch.LossCoal, CoalEmoji, grinch.LossYellowPerGroupMember)}
}

func guideTree(tuning core.Tuning) []string {
	tree := tuning.Tree
	water := fmt.Sprintf("## 🌲 Christmas tree\nOne tree for the whole server. It needs 100%% water and loses 1%% every %s.\n"+
		"**`/water`**: adds %d%% of water, once every %s.", interval(tree.DecayInterval), tree.WaterPoints, interval(tree.WaterCooldown))
	if tree.WaterAura > 0 {
		water += fmt.Sprintf(" From an aura of 0 up, it is also +%d good deed.", tree.WaterAura)
	}
	return []string{
		water,
		"💧 An extra watering from a Rare gift lets you skip the wait once.",
		fmt.Sprintf("-# The tree can be watered until %s UTC.", tree.EventEnd.UTC().Format("January 2, 2006 at 15:04")),
	}
}

func guideAdvent() []string {
	return []string{
		fmt.Sprintf("## 📅 Advent calendar\n**`/advent view`** shows your %d doors, in an order that is only yours.", core.AdventDays),
		"A door unlocks each day from December 1, at 00:00 UTC. Missed days can still be opened until the end of December 24 UTC.",
	}
}

func guideMore() []string {
	return []string{
		"## 📊 Stats and more\n" +
			"**`/stats`**: the server's numbers and the leaderboards\n" +
			"**`/help me-milky`**: the list of commands",
		"💬 Some channels share Christmas quotes and poems while people chat.",
	}
}

func coalReward(coal int64) string {
	if coal == 0 {
		return "no " + CoalEmoji + " Coal"
	}
	return fmt.Sprintf("%d %s Coal", coal, CoalEmoji)
}

// interval words a duration after "every": "hour", "24 hours".
func interval(d time.Duration) string {
	return strings.TrimPrefix(span(d), "1 ")
}

// span words a duration: "1 minute", "24 hours", "90 seconds".
func span(d time.Duration) string {
	unit, size := "second", time.Second
	switch {
	case d >= time.Hour && d%time.Hour == 0:
		unit, size = "hour", time.Hour
	case d >= time.Minute && d%time.Minute == 0:
		unit, size = "minute", time.Minute
	}
	count := int(d.Round(time.Second) / size)
	if count != 1 {
		unit += "s"
	}
	return fmt.Sprintf("%d %s", count, unit)
}
