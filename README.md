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

Wait for the startup log to report that the bot is ready. In Discord, enable commands in a text channel with `/setup channel`, then send the Milky Xmas `/hello` command there. The expected result is one public reply, `Hello, world!`. Check that the command is available in the configured server and can be used by a member with application command permissions.

In the test server, invoke `/daily claim`. Check that you see the same success message as in the local chat with an amount from 1 to 10 in a private reply. Invoke it again immediately and check that it shows the remaining cooldown. Restart the bot with the same database and try again to check that the original cooldown persists.

## Gift transfers and aura

Daily claims award one rarity per claim with a 70% chance of normal, 25% rare, and 5% epic. Daily gifts saved before this update become normal gifts. Each gift sent adds 5, 10, or 30 percentage points to the sender's Good deeds, respectively, up to 100%. In `/send gift`, choose Coal to transfer Coal and increase the sender's Bad deeds by five percentage points per item, capped at 100%. Bad deeds appears on the left and Good deeds on the right. Both start at 0%.

In the local chat, run `/aura check`, then `/daily claim`. Use the rarity shown in the reply to send one gift, for example `/send gift friend normal 1`. Check that the confirmation and a new `/aura check` show the expected Good deeds value. Try sending more than you claimed and check that the aura stays the same. The browser plays as `fixture`, and `friend` is a simulated recipient. A claim after giving away your last gift still shows the original cooldown.

In the Discord test server, use two human accounts with application command permissions. With the first account, run `/aura check`, claim gifts, then select the second account, the awarded rarity, and quantity 1 under `/send gift`. Check that both command results are private. The first account's Good deeds increases and the second account's aura stays neutral. With the second account, send the same gift back to verify it was credited. Try selecting yourself or a bot to check that the transfer is rejected. Restart the bot with the same SQLite database and confirm that both accounts keep their aura and current balances.

## Socks inventory

Run `/look in-your-stocking` (Discord: `/look`, subcommand `in-your-stocking`) to see your normal, rare, and epic gifts, Snowballs, Yellow Snowballs, and Coal, including items with zero balance. Snowballs, Yellow Snowballs, and Coal start at zero. Advent gifts appear in the Advent calendar.

In a fresh local chat, run `/look in-your-stocking` to check all six balances are zero. Run `/daily claim`, then `/look in-your-stocking` and check the quantity for the awarded rarity. Send one gift to `friend` and use `/look in-your-stocking` to check that balance decreased by one.

In the Discord test server, invoke `/look in-your-stocking` and check that only your account can see the reply, with the stocking image above the inventory. Send a gift to a second account and confirm each account sees its updated balance when it runs `/look in-your-stocking`.

## Throw a snowball

In the Discord test server, use `/thrown snowball` and select another player. Use the optional type field to choose Normal or Yellow Snowball. Yellow appears when you have one. Throwing it transfers one to the recipient and adds five percentage points to your Bad deeds, up to 100%. The sender sees a private confirmation and the bot mentions the recipient in the same channel with their new Snowball total. The recipient can run `/look in-your-stocking` to check the balance. If the bot cannot post in that channel, the inventory still increases and the sender sees a delivery warning.

In the local chat, run `/thrown snowball friend` twice. The transcript shows the confirmation and simulated notification. The notification reports one Snowball on the first throw and two on the second. Use `/thrown snowball friend yellow` to throw a Yellow Snowball from your inventory. Use `/send gift friend coal 1` to transfer Coal.

## Advent calendar

Use `/advent view` to see the 24 doors and your Advent gift total. Use `/advent open 1` to open door 1 in the local chat. In Discord, select the `open` subcommand and enter the door number in the `day` field.

Each door unlocks at midnight UTC on its December date and stays available until December 24 ends UTC. An opened door counts as one Advent gift for that year. The daily gift total and cooldown are tracked separately.

During December 1–24 UTC, open a door whose date has passed and confirm the Advent total increases. Try the same door again and confirm the total stays the same. On Discord, check that replies are private and the door remains open after a restart with the same database. Outside this window, verify that unopened doors cannot be claimed. The deterministic adapter tests cover the active window with a fixed December clock.

## Discord channel setup

When the bot joins a server, it asks moderators to configure it in the system channel or another text channel where it can post. Use `/setup channel` with a text channel and set `commands` and `boss` independently. Repeat for each channel. Both can be turned off to remove access. `/setup view` shows the configured channels. Setup requires **Manage Server** or **Administrator** permission.

Select a general channel with commands on and boss off. Select an event channel with boss on. The bot posts the Christmas introduction when that channel is first enabled for the boss. If it cannot post, grant **View Channel** and **Send Messages**, then repeat the setup. Check that `/hello` works in the general channel and is refused in an unselected channel. Restart the bot and check that `/setup view` still shows the same choices.

## Grinch encounter

Use `/setup channel` to enable the boss in each Discord text channel where it may appear. Enable commands there too if moderators should be able to invoke `/grinch summon`. That command requires **Manage Messages**. Test it twice during an active fight and confirm that only one encounter starts. Use `/setup channel` for installations that previously selected a channel through `MILKY_GRINCH_CHANNEL_ID`.

The Grinch can also appear automatically. Each minute, the bot checks the designated channel for human activity within the last ten minutes and rolls a 10% spawn chance if there is no fight or automatic spawn cooldown. Automatic spawning cools down for 30 minutes after a fight. Manual summons skip that cooldown.

Click 🎅 **Open registration** in the announcement, then **Join** in your private panel. The Join button disappears from your panel after you register. Combat starts as soon as three players join or after one minute with one or two players. If nobody joins, the encounter is cancelled. New players can still join during combat. Throw snowballs until the Grinch reaches 75% HP. This first phase has no time limit. The shared one-minute countdown starts at 75% HP. Choose 🎄 to defend and damage him by 30 percentage points divided by the number who joined with 🎅 before defense begins, down to 45%. The defense emojis reshuffle after each accepted click. At 45%, the 🧊 and ☃️ attack grid appears. At 5%, each accepted ⚔️ deals 1 HP and awards 2 damage points, and 🎁 awards 1–5 gift points. Pick ❄️ for one Snowball in `/look in-your-stocking`. The four attack and gift buttons are rerolled after each accepted choice. The countdown continues through defense, the attack grid, and the final attack. The final message shows separate damage and gift rankings for everyone who joined. Give the bot **Attach Files** permission to include full standings when they exceed Discord's message limit. One player can fight and win alone. If all registered players reach zero HP, the Grinch wins immediately. If the Grinch wins or escapes, every registered player receives one Coal and a number of Yellow Snowballs equal to the group size. A group of four receives four Yellow Snowballs per player.

In the local chat, use `/grinch summon`, then `/grinch join` within one minute. Use `/grinch view` to see when combat has started. Use `/grinch hit 1` for snowballs, or `/grinch hit <tile>` for the numbered boards. The local fixture needs no Discord credentials.

## Moderator gifts and reply formats

Moderators with Manage Messages or Administrator can run `/give gift <recipient> <amount> <rarity>` to add normal, rare, or epic gifts to a player. It does not use the moderator's inventory and does not change aura or cooldowns. Reply is private. In the local chat, run `/give gift friend 3 rare`; the browser acts as the moderator.

`/daily claim`, `/look in-your-stocking`, and `/aura check` use a heading and code block. In Discord, `/aura check` shows the bad good image above `😈 Bad n%  |  n% Good 😇`. The local chat shows the text only.

In the Discord test server, check that a member without Manage Messages cannot see `/give` or is refused, that a moderator's grant appears in the recipient's `/look in-your-stocking`, and that the aura image appears above the aura line.

The snowball notification reads `<recipient>❕  <sender> has thrown a <type> at you. 🥶`. In Discord it mentions the recipient and shows the snowball image below the text. Check that after `/thrown snowball`.

The Grinch announcement shows the Grinch image below the registration text. In the Discord test server, summon the Grinch with `/grinch summon` and check the image, then join and confirm the image stays while the fight updates.

If the Grinch escapes or defeats all players, the final battle message also shows Max's image next to the Grinch image. A victory shows only the Grinch. Check both endings in the Discord test server.

Run `/helpmemilky` to list the commands you can use. In Discord, moderator commands appear only for members with the permission: Manage Messages for `/give gift` and `/grinch summon`, Manage Server for `/setup`. The reply is private. The list is written by hand in `gifttext/help.go`, so update it when a command is added or renamed. In the Discord test server, run it with a regular account and with a moderator account and compare the lists.

## Activity drop updates

The Random Active Drop shows the Grinch and Max image and the Gift Rain shows the gift rain image. A player with Good deeds of 50% or more has better odds of rare and epic gifts from the Random Active Drop. The button a player opens is disabled. A Gift Rain ends after ten collections and each has a 25% chance of Coal. When a drop ends or expires after ten minutes, its message is edited to say so and to list who collected what. Drops are kept in memory, so a restart leaves an open drop's message as it was.

In the Discord test server, enable a channel and chat until a drop appears. Check the image, that your clicked button is disabled, that a second account cannot open the same button, and that the final list appears after ten minutes. For a Gift Rain, collect ten times across accounts (one collection per minute each) and check the end and the Coal prizes. In the local chat, use `/message <text>` to trigger drops, `/drop collect <id>`, and `/drop expire <id>`.

## Advent calendar buttons

In Discord, `/advent view` shows the calendar image with 24 numbered buttons. Click a number to open that door once. The button becomes ❌ and stays disabled. Doors that are not available yet, or after December 24 UTC, are disabled. `/advent open <day>` still works, and the local chat still lists the doors in text.

In the Discord test server, run `/advent view` and click today's number, then check the ❌, that the gift total went up, and that the button no longer responds. Run `/advent view` again to see the ❌ persist, and try a number from a later day to confirm it is disabled.

**Temporary test window:** until November 29, 2026 (UTC), every Advent door is available, so the calendar can be tested before December. Doors opened in the window are stored separately and do not count in December, but gifts awarded during the test stay in inventories. Remove the window from `core/advent.go` afterwards.
