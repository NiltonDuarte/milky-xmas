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

Daily claims award one rarity per claim with a 70% chance of normal, 25% rare, and 5% epic. Daily gifts saved before this update become normal gifts. Aura is a single scale from -50 (bad deeds) to +50 (good deeds), and everyone starts neutral at 0. Each gift sent moves the sender 5, 10, or 30 points toward good, respectively. In `/send gift`, choose Coal to transfer Coal, which moves the sender 5 points per item toward bad. The aura never leaves the scale. `/aura check` draws the scale with a marker on its middle, a finger under your place, and your value beside your side. Upgrading from the earlier separate Bad and Good percentages resets everyone to 0.

In the local chat, run `/aura check`, then `/daily claim`. Use the rarity shown in the reply to send one gift, for example `/send gift friend normal 1`. Check that the confirmation and a new `/aura check` show the expected aura. Try sending more than you claimed and check that the aura stays the same. The browser plays as `fixture`, and `friend` is a simulated recipient. A claim after giving away your last gift still shows the original cooldown.

In the Discord test server, use two human accounts with application command permissions. With the first account, run `/aura check`, claim gifts, then select the second account, the awarded rarity, and quantity 1 under `/send gift`. Check that both command results are private. The first account's aura moves toward good and the second account's aura stays neutral. With the second account, send the same gift back to verify it was credited. Try selecting yourself or a bot to check that the transfer is rejected. Restart the bot with the same SQLite database and confirm that both accounts keep their aura and current balances.

## Socks inventory

Run `/look in-your-stocking` (Discord: `/look`, subcommand `in-your-stocking`) to see your normal, rare, and epic gifts, Snowballs, Yellow Snowballs, and Coal, including items with zero balance. Snowballs, Yellow Snowballs, and Coal start at zero.

In a fresh local chat, run `/look in-your-stocking` to check all six balances are zero. Run `/daily claim`, then `/look in-your-stocking` and check the quantity for the awarded rarity. Send one gift to `friend` and use `/look in-your-stocking` to check that balance decreased by one.

In the Discord test server, invoke `/look in-your-stocking` and check that only your account can see the reply, with the stocking image above the inventory. Send a gift to a second account and confirm each account sees its updated balance when it runs `/look in-your-stocking`.

## Throw a snowball

In the Discord test server, use `/throw snowball` and select another player. Then choose the type, Normal or Yellow Snowball; the type is required, like the player. Each type appears only when you own one. A throw moves one snowball of that type from your inventory to the recipient's, and a Yellow Snowball also moves your aura 5 points toward bad. The sender sees a private confirmation and the bot mentions the recipient in the same channel with their new Snowball total. The recipient can run `/look in-your-stocking` to check the balance. If you have none to throw, or the recipient's inventory is full, the throw is refused and nothing moves. If the bot cannot post in that channel, the inventory still changes and the sender sees a delivery warning.

In the local chat, give yourself Snowballs first (they come from the Grinch fight rewards), then run `/throw snowball friend`. The transcript shows the confirmation and simulated notification, and a throw without a Snowball is refused. Use `/throw snowball friend yellow` to throw a Yellow Snowball from your inventory. Use `/send gift friend coal 1` to transfer Coal.

## Advent calendar

Use `/advent view` to see your calendar: the image and 24 numbered buttons. Each player sees the buttons in their own order, and behind each door each player finds a different surprise, such as a recipe. Click a number to get its surprise in a private message. A door can be opened from midnight UTC on its December date until December 24 ends UTC; a missed door can be opened later in that time. An opened door turns green and can be clicked again to read its surprise. Doors that are not available are disabled.

The surprises are the files `src/internal/adapter/adventtext/content/01.md` to `24.md`. Edit them and rebuild the bot. The file number is not the door number, because each player's order differs.

The local chat has no buttons: `/advent view` lists the doors in your order, and `/advent open <day>` stands in for a click.

## Discord channel setup

When the bot joins a server, it asks moderators to configure it in the system channel or another text channel where it can post. Use `/setup channel` with a text channel and set `commands` and `boss` independently. Repeat for each channel. Both can be turned off to remove access. `/setup view` shows the configured channels. Setup requires **Manage Server** or **Administrator** permission.

Select a general channel with commands on and boss off. Select an event channel with boss on. The bot posts the Christmas introduction when that channel is first enabled for the boss. If it cannot post, grant **View Channel** and **Send Messages**, then repeat the setup. Check that `/hello` works in the general channel and is refused in an unselected channel. Restart the bot and check that `/setup view` still shows the same choices.

## Christmas quotes

Turn them on with `/setup channel` and `quotes: True` (the other two choices are set as usual; leaving `quotes` out keeps the current one). In those channels the bot posts a Christmas quote or poem after 3 player messages, at most one every 3 hours (the time of the last quote is saved, so a restart keeps the interval); both numbers are in the `quotes` section of `tuning.yaml`. A quiet channel gets nothing, and no quote repeats until all were shown. The texts are in `core/quotes.go` and must stay in the public domain. In the local chat, send `/message hi` until a quote appears.

## Grinch encounter

Use `/setup channel` to enable the boss in each Discord text channel where it may appear. Enable commands there too if moderators should be able to invoke `/grinch summon`. That command requires **Manage Messages**. Test it twice during an active fight and confirm that only one encounter starts. Use `/setup channel` for installations that previously selected a channel through `MILKY_GRINCH_CHANNEL_ID`.

The Grinch can also appear automatically. Each minute, the bot checks the designated channel for human activity within the last ten minutes and rolls a 10% spawn chance if there is no fight or automatic spawn cooldown. Automatic spawning cools down for 30 minutes after a fight. Manual summons skip that cooldown.

Click 🎅 **Start** in the announcement. The first click begins a one-minute countdown, and the fight starts at once when three players have pressed Start. There is no joining: everyone can take part, even after the fight has started, and a player is in the fight from their first accepted click on an attack button. If nobody presses it within one minute of the announcement, the encounter is cancelled. Only you see a confirmation of your Start, once. Clicking Start again does nothing. The fight is one panel (a Discord Components V2 message) with the Grinch image. Two small text areas of fixed height, in a box of their own, show the Grinch's lines (a random line from a pool, with the damage his snowballs did to players and the gift points and Snowballs players took back from him) and the players' attacks (who attacked and the damage they did). When a player hides behind the tree in the defense phase, they say a line too. The attack buttons of each phase sit in that box, below the two text areas, with the players' cards below the buttons in the same box. While players register it lists who joined. During the fight it shows the boss HP and a card for each of the eight players who dealt the most damage, with their HP and damage, and one line for the others. At the end it lists the damage, gift, and Snowball leaderboards, ten players each, and attaches the full standings when more players scored. Players with zero points are not listed, and a leaderboard nobody scored on is left out. Throw snowballs until the Grinch reaches 75% HP. The group has two minutes to do it. If time runs out, the Grinch wins. The shared one-minute countdown starts at 75% HP. The group needs one safe defense per player, but never more than six, and together they take him down to 45%. Choose 🎄 to defend. The defense emojis reshuffle after each accepted click. At 45%, the 🧊 and ☃️ attack grid appears. Six hits are enough to clear each phase, however many players there are. At 5%, each accepted ⚔️ deals 0.84 HP and awards 2 points, and 🎁 awards 1–5 gift points. Only the final attack awards points. The points are kept for each player for later use, and the damage leaderboard shows only the damage dealt. Pick ❄️ for one Snowball in `/look in-your-stocking`. The four attack and gift buttons are rerolled after each accepted choice. The countdown continues through defense, the attack grid, and the final attack. The final message shows separate damage and gift rankings for everyone who joined. Give the bot **Attach Files** permission to include full standings when they exceed Discord's message limit. One player can fight and win alone. If all registered players reach zero HP, the Grinch wins immediately. If the Grinch wins or escapes, every registered player receives one Coal and a number of Yellow Snowballs equal to the group size. A group of four receives four Yellow Snowballs per player.

In the local chat, use `/grinch summon`, then `/grinch join` to press Start. Any player can also just `/grinch hit` once the fight is running. Use `/grinch view` to see when combat has started. Use `/grinch hit 1` for snowballs, or `/grinch hit <tile>` for the numbered boards. The local fixture needs no Discord credentials.

## Moderator gifts and reply formats

Moderators with Manage Messages or Administrator can run `/give gift <recipient> <amount> <rarity>` to add normal, rare, or epic gifts to a player. It does not use the moderator's inventory and does not change aura or cooldowns. Reply is private. In the local chat, run `/give gift friend 3 rare`; the browser acts as the moderator.

`/daily claim` uses a heading and code block, `/look in-your-stocking` shows the contents in three sections divided by lines, and `/aura check` uses a code block. In Discord, `/aura check` shows the bad good image above the aura scale, both in one box. The local chat shows the text only.

In the Discord test server, check that a member without Manage Messages cannot see `/give` or is refused, that a moderator's grant appears in the recipient's `/look in-your-stocking`, and that the aura image appears above the aura line.

The snowball notification reads `<recipient>❕  <sender> has thrown a <type> at you. 🥶`. In Discord it mentions the recipient and shows the snowball image below the text. Check that after `/throw snowball`.

The Grinch announcement shows the Grinch image below the registration text. In the Discord test server, summon the Grinch with `/grinch summon` and check the image, then press Start and confirm the image stays while the fight updates.

If the Grinch escapes or defeats all players, the final battle message shows only Max's image, with no Grinch image. A victory shows only the Grinch. Check both endings in the Discord test server.

Run `/help` or `/helpmemilky` to list the commands you can use. In Discord, moderator commands appear only for members with the permission: Manage Messages for `/give gift` and `/grinch summon`, Manage Server for `/setup`. The reply is private. The list is written by hand in `gifttext/help.go`, so update it when a command is added or renamed. In the Discord test server, run it with a regular account and with a moderator account and compare the lists.

## Activity drop updates

The Random Active Drop shows the Grinch and Max image and the Gift Rain shows the gift rain image. A player with an aura of +25 or more has better odds of rare and epic gifts from the Random Active Drop. The button a player opens is disabled. When a drop's ten minutes run out, its message says that the Grinch stole all the presents. A Gift Rain ends after ten collections and each has a 25% chance of Coal. When a drop ends or expires after ten minutes, its message is edited to say so and to list who collected what. Drops are kept in memory, so a restart leaves an open drop's message as it was.

In the Discord test server, enable a channel and chat until a drop appears. Check the image, that your clicked button is disabled, that a second account cannot open the same button, and that the final list appears after ten minutes. For a Gift Rain, collect ten times across accounts (one collection per minute each) and check the end and the Coal prizes. In the local chat, use `/message <text>` to trigger drops, `/drop collect <id>`, and `/drop expire <id>`.

## Advent calendar acceptance

In the Discord test server, run `/advent view`, click today's number, and check that the surprise arrives in a private message, the button turns green, and clicking it again shows the same surprise. Check that a later day's button is disabled, and that a second account sees a different button order.

**Temporary test window:** until November 29, 2026 (UTC), every Advent door is available, so the calendar can be tested before December. Doors opened in the window are stored separately and do not count in December. Remove the window from `core/advent.go` afterwards.

## Christmas tree

Run `/water` to water the server's Christmas tree. It starts withered at 10% water, loses 1% every hour, and gains 2% per watering. It needs 100% to be full and alive. Each player can water once every hour. The event ends on January 5, 2027 at 01:00 UTC. These numbers are the defaults of the `tree` section in `tuning.yaml` (`start_level`, `water_points`, `water_aura`, `decay_interval`, `water_cooldown`, `event_end`); the 100% maximum is fixed. After that the tree keeps its final level and `/water` only shows it privately. Each watering that counts also moves the player's aura 1 point toward good deeds (`water_aura`; 0 turns this off), up to +50. In Discord a watering is posted to the channel as one box with the tree picture, and a ten-square water bar: the withered tree up to 50% and the decorated tree at 100%; from 51% to 99% there is no picture yet. A refusal, such as watering too soon or watering a full tree, is private; for a full tree the player also sees the decorated tree. The local chat shows the same text without the image.

In the Discord test server, run `/water` from two accounts and check the public box, the picture, and `/aura check`, then run it again right away to see the private refusal. Restart the bot with the same database to check the level continues.

## Christmas spirit

When the event ends, a moderator (Manage Messages) runs `/spirit` to post the Christmas spirit of the server: every player's good deeds and bad deeds added up from their current aura, with the winning side. The same picture is shown for either side until one exists for bad deeds; the current one is a watermarked preview to be replaced.

## Stats and leaderboards

Anyone can run `/stats` to open a private panel. Pick **Server stats** (tree level, waterings, gifts and Coal sent, snowballs thrown, drops collected, and Grinch fights) or **Leaderboards**, and then a category: Bad eggs, Good eggs, Throwers, Gifters, or Plant moms and dads. Each lists the top ten. Bad and Good eggs add up deed points without one side lowering the other and without the aura's limit. In the local chat use `/stats server`, `/stats bad`, `good`, `throwers`, `gifters`, or `plants`.

## Event history

For the scoreboards at the end of the event, the bot's database keeps every finished Grinch fight (`grinch_history`) and every inventory or aura change with its reason (`player_history`). Waterings stay counted in the tree and opened Advent doors in `advent_openings`. `/stats` is built from these records.

## Tuning the numbers

The balance numbers in [docs/activity-drops-numbers.md](docs/activity-drops-numbers.md) and [docs/grinch-fight-numbers.md](docs/grinch-fight-numbers.md) live in [tuning.yaml](tuning.yaml): drop odds and lifetimes, Gift Rain size and Coal chance, gift rarity odds, and every Grinch timer, HP milestone, damage value, and reward. The file holds the defaults with a comment per value. Without it the bot uses the same built-in numbers.

Point the bot at the file before starting it. Both the local chat and the Discord bot read it once at startup, so restart to apply a change.

```sh
MILKY_TUNING_CONFIG=tuning.yaml go run ./src/cmd/web
MILKY_TUNING_CONFIG=tuning.yaml go run ./src/cmd/discord
```

Keys you leave out keep their default. An unknown key, a bad unit, or a value outside its range stops the bot at startup and the message lists every problem with its key path. Durations are strings such as `90s` or `10m`, and boss HP and damage are percentages with up to two decimals.

In the local chat, copy `tuning.yaml`, set `drops.gift_rain.min_messages` and `max_messages` to `2`, `claims_to_end` to `2`, and `claim_cooldown` to `0s`, then start the chat with the copy. Send `/message hi` twice to start a Gift Rain and `/drop collect 1` twice. The second collection ends the drop. For the Grinch, set `grinch.registration.early_start_players` to `1` and `grinch.snowballs.boss_hp_after` to `90`. Run `/grinch summon` and `/grinch join`, then ten `/grinch hit 1`. The fight reaches the defense phase at 90 HP. In the Discord test server, start the bot with the same file and check that the announcement shows `0/1 to start early` and `reaches **90 HP**`.

Make an invalid change, such as `gifts.daily_rarity.epic: 10`, to check the startup error that names `gifts.daily_rarity`.

## Log file

The Discord bot writes its log to standard error and, when `MILKY_LOG_FILE` is set, also appends it to that file (owner-only permissions; a file above 10 MB is moved to `<file>.1` at startup). A failed response to Discord logs how old the interaction was, since Discord wants the first response within three seconds of its creation.
