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

Send `/daily claim` to collect one gift and see the total. Send it again immediately to see the next eligible time in UTC. The local fixture shares one test identity across tabs and resets when the web process restarts.

The chat binds to loopback and works with fixture behavior. It requires no Discord credentials. The transcript lives in the current page.

The planned opt-in Discord mode for the local chat can read or post bot messages in a configured test channel. That mode keeps the token server-side and uses Discord channel permissions to scope access. [Discord message endpoints](https://docs.discord.com/developers/resources/message#get-channel-messages)

## Discord test server

Create a bot in the [Discord Developer Portal](https://discord.com/developers/applications) and install it in a test server with the bot and application commands scopes. Use a channel where your account can send messages and use application commands. Leave the application's Interactions Endpoint URL blank so it receives interactions through the Gateway. [Discord interaction delivery](https://docs.discord.com/developers/interactions/receiving-and-responding)

Set `DISCORD_BOT_TOKEN` to the bot token in your local environment and set `DISCORD_TEST_GUILD_ID` to the test server ID. Set `MILKY_GIFT_DB` to a SQLite file on a persistent writable volume. The bot creates the file and schema on startup. Keep that file across restarts to retain gift totals and cooldowns.

Start the bot:

```sh
go run ./src/cmd/discord
```

Wait for the startup log to report that `/hello` is ready in the test server. In Discord, choose the Milky Xmas `/hello` command and send it. The expected result is one public reply, `Hello, world!`. Check that the command is available in the configured server and can be used by a member with application command permissions.

In the test server, invoke `/daily claim`. The first claim should give one gift and show the total in a private reply. Invoke it again immediately and check that it shows the next eligible time. Restart the bot with the same database and try again to check that the original cooldown persists.

## Advent calendar

Use `/advent view` to see the 24 doors and your Advent gift total. Use `/advent open 1` to open door 1 in the local chat. In Discord, select the `open` subcommand and enter the door number in the `day` field.

Each door unlocks at midnight UTC on its December date and stays available until December 24 ends UTC. An opened door counts as one Advent gift for that year. The daily gift total and cooldown are tracked separately.

During December 1–24 UTC, open a door whose date has passed and confirm the Advent total increases. Try the same door again and confirm the total stays the same. On Discord, check that replies are private and the door remains open after a restart with the same database. Outside this window, verify that unopened doors cannot be claimed. The deterministic adapter tests cover the active window with a fixed December clock.
