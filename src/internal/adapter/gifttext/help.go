package gifttext

import "strings"

const HelpCommand = "helpmemilky"

// HelpAccess says which restricted command groups the player may use.
type HelpAccess struct {
	Moderator bool // Manage Messages
	Manager   bool // Manage Server
}

// Help lists the commands available to the player.
func Help(access HelpAccess) string {
	var text strings.Builder
	text.WriteString("## 🎅 Milky Xmas commands\n```\n")
	for _, line := range []string{
		"/hello                          Get a greeting",
		"/daily claim                    Claim your daily gifts",
		"/look in-your-stocking          See your gifts and items",
		"/send gift <player> <item> <n>  Send gifts or Coal",
		"/aura check                     See your Bad and Good deeds",
		"/throw snowball <player>        Throw a snowball",
		"/advent view                    See your Advent doors",
		"/water                          Water the Christmas tree",
		"/helpmemilky                    Show this list",
	} {
		text.WriteString(line + "\n")
	}
	text.WriteString("```")
	if access.Moderator || access.Manager {
		text.WriteString("\n**Moderator commands**\n```\n")
		if access.Moderator {
			text.WriteString("/give gift <player> <amount> <rarity>  Give gifts to a player\n")
			text.WriteString("/grinch summon                         Summon the Grinch\n")
		}
		if access.Manager {
			text.WriteString("/setup view                            See the configured channels\n")
			text.WriteString("/setup channel                         Set a channel's permissions\n")
		}
		text.WriteString("```")
	}
	return text.String()
}
