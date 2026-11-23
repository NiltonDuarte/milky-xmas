# milky-xmas

Go Discord bot for the Milky community Christmas event.

## Develop

Use Go 1.26 or newer and OpenSpec. For a new feature, start with `$openspec-propose "your idea"` in Codex. Review the behavior scenarios and design, then implement the selected change with `$openspec-apply-change <change-id>`. See [AGENTS.md](AGENTS.md) for the complete development and review process.

Run the checks from the repo root.

```sh
go fmt ./...
go vet ./...
go test -race -count=1 ./...
openspec validate --all --strict --no-interactive
openspec validate --archived --no-interactive
```

The test suite checks FCIS dependency direction, the greeting rule, and command handling through both adapters.

## Local acceptance testing

Start the local chat from the repository root.

```sh
go run ./src/cmd/web
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080) and send `/hello`. The transcript should show your command and the bot's reply, `Hello, world!`. Send `/unknown` to see an error that lists the supported commands, then send `/hello` again to verify recovery.

Send `/daily claim` to collect a random 1–10 gifts. The transcript shows “🎅🏻 Christmas won’t be Christmas without any presents. You collected X gifts 🎁.” with the amount awarded in place of X, followed by the rarity. Send it again immediately to see how long remains until the next claim. The local fixture shares one test identity across tabs and resets when the web process restarts.

The chat binds to loopback and works with fixture behavior. It requires no Discord credentials. The transcript lives in the current page.

The planned opt-in Discord mode for the local chat can read or post bot messages in a configured test channel. That mode keeps the token server-side and uses Discord channel permissions to scope access. [Discord message endpoints](https://docs.discord.com/developers/resources/message#get-channel-messages)

## Discord test server

Create a bot in the [Discord Developer Portal](https://discord.com/developers/applications) and install it in a test server with the bot and application commands scopes. Use a channel where your account can send messages and use application commands. Leave the application's Interactions Endpoint URL blank so it receives interactions through the Gateway. [Discord interaction delivery](https://docs.discord.com/developers/interactions/receiving-and-responding)

Set `DISCORD_BOT_TOKEN` to the bot token in your local environment and set `DISCORD_TEST_GUILD_ID` to the test server ID. Set `MILKY_GIFT_DB` to a SQLite file on a persistent writable volume. The bot creates the file and schema on startup. Keep that file across restarts to retain gift inventories, aura, and cooldowns.

Start the bot:

```sh
go run ./src/cmd/discord
```

Wait for the startup log to report that `/hello` is ready in the test server. In Discord, choose the Milky Xmas `/hello` command and send it. The expected result is one public reply, `Hello, world!`. Check that the command is available in the configured server and can be used by a member with application command permissions.

In the test server, invoke `/daily claim`. Check that you see the same success message as in the local chat with an amount from 1 to 10 in a private reply. Invoke it again immediately and check that it shows the remaining cooldown. Restart the bot with the same database and try again to check that the original cooldown persists.

## Gift transfers and aura

Daily claims award one rarity per claim with a 70% chance of normal, 25% rare, and 5% epic. Daily gifts saved before this update become normal gifts. Each gift sent adds 5, 10, or 30 percentage points to the sender's Good deeds, respectively, up to 100%. Bad deeds appears on the left and Good deeds on the right. Both start at 0%.

In the local chat, run `/aura check`, then `/daily claim`. Use the rarity shown in the reply to send one gift, for example `/send gift friend normal 1`. Check that the confirmation and a new `/aura check` show the expected Good deeds value. Try sending more than you claimed and check that the aura stays the same. The browser plays as `fixture`, and `friend` is a simulated recipient. A claim after giving away your last gift still shows the original cooldown.

In the Discord test server, use two human accounts with application command permissions. With the first account, run `/aura check`, claim gifts, then select the second account, the awarded rarity, and quantity 1 under `/send gift`. Check that both command results are private. The first account's Good deeds increases and the second account's aura stays neutral. With the second account, send the same gift back to verify it was credited. Try selecting yourself or a bot to check that the transfer is rejected. Restart the bot with the same SQLite database and confirm that both accounts keep their aura and current balances.

## Socks inventory

Run `/socks` to see your normal, rare, and epic gifts, Snowballs, and Coal, including items with zero balance. Snowballs and Coal start at zero. Advent gifts appear in the Advent calendar.

In a fresh local chat, run `/socks` to check all five balances are zero. Run `/daily claim`, then `/socks` and check the quantity for the awarded rarity. Send one gift to `friend` and use `/socks` to check that balance decreased by one.

In the Discord test server, invoke `/socks` and check that only your account can see the reply. Send a gift to a second account and confirm each account sees its updated balance when it runs `/socks`.

## Advent calendar

Use `/advent view` to see the 24 doors and your Advent gift total. Use `/advent open 1` to open door 1 in the local chat. In Discord, select the `open` subcommand and enter the door number in the `day` field.

Each door unlocks at midnight UTC on its December date and stays available until December 24 ends UTC. An opened door counts as one Advent gift for that year. The daily gift total and cooldown are tracked separately.

During December 1–24 UTC, open a door whose date has passed and confirm the Advent total increases. Try the same door again and confirm the total stays the same. On Discord, check that replies are private and the door remains open after a restart with the same database. Outside this window, verify that unopened doors cannot be claimed. The deterministic adapter tests cover the active window with a fixed December clock.

## Grinch encounter

Set `MILKY_GRINCH_CHANNEL_ID` to the ID of a Discord text channel in the configured server, then restart the bot. Allow the bot to view the channel and send messages there. A member with **Manage Messages** can invoke `/grinch summon` in that channel. Try the command again during a fight to verify it creates only one announcement. With a regular member or in another channel, verify the summon is unavailable.

The Grinch can also appear automatically. Each minute, the bot checks the designated channel for human activity within the last ten minutes and rolls a 10% spawn chance if there is no fight or automatic spawn cooldown. Automatic spawning cools down for 30 minutes after a fight. Manual summons skip that cooldown.

Click 🎅 **Join** in the announcement. Combat starts as soon as three players join or after one minute with one or two players. If nobody joins, the encounter is cancelled. New players can still join during combat. Throw snowballs until the Grinch reaches 75% HP. This first phase has no time limit. The shared one-minute countdown starts at 75% HP. Choose 🎄 to defend and damage him by 30 percentage points divided by the number who joined with 🎅 before defense begins, down to 45%. The defense emojis reshuffle after each accepted click. At 45%, the 🧊 and ☃️ attack grid appears. At 5%, each accepted ⚔️ deals 1 HP and awards 2 damage points, and 🎁 awards 1–5 gift points. The four buttons are rerolled after each accepted choice. The countdown continues through defense, the attack grid, and the final attack. The final message shows separate damage and gift rankings for everyone who joined. Give the bot **Attach Files** permission to include full standings when they exceed Discord's message limit. One player can fight and win alone. If all registered players reach zero HP, the Grinch wins immediately.

In the local chat, use `/grinch summon`, then `/grinch join` within one minute. Use `/grinch view` to see when combat has started. Use `/grinch hit 1` for snowballs, or `/grinch hit <tile>` for the numbered boards. The local fixture needs no Discord credentials.
