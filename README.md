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

Open [http://127.0.0.1:8080](http://127.0.0.1:8080) and send `/hello`. The transcript should show your command and the bot's reply, `Hello, world!`. Send `/unknown` to see an error that points to `/hello`, then send `/hello` again to verify recovery.

The chat binds to loopback and works with fixture behavior. It requires no Discord credentials. The transcript lives in the current page.

The planned opt-in Discord mode for the local chat can read or post bot messages in a configured test channel. That mode keeps the token server-side and uses Discord channel permissions to scope access. [Discord message endpoints](https://docs.discord.com/developers/resources/message#get-channel-messages)

## Discord test server

Create a bot in the [Discord Developer Portal](https://discord.com/developers/applications) and install it in a test server with the bot and application commands scopes. Use a channel where your account can send messages and use application commands. Leave the application's Interactions Endpoint URL blank so it receives interactions through the Gateway. [Discord interaction delivery](https://docs.discord.com/developers/interactions/receiving-and-responding)

Set `DISCORD_BOT_TOKEN` to the bot token in your local environment and set `DISCORD_TEST_GUILD_ID` to the test server ID. Then start the bot.

```sh
go run ./src/cmd/discord
```

Wait for the startup log to report that `/hello` is ready in the test server. In Discord, choose the Milky Xmas `/hello` command and send it. The expected result is one public reply, `Hello, world!`. Check that the command is available in the configured server and can be used by a member with application command permissions.
