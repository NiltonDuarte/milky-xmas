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

The starter suite checks dependency direction across the FCIS layers. Add scenario tests alongside the first bot behavior.

## Local acceptance testing

The proposed test interface is a small chat page served by Go on localhost, with a transcript and a composer. The web and Discord adapters translate their inputs into the same application requests and render the results for their respective interfaces. The shared Go service runs authorization, loads state, calls the functional core, and performs writes or sends messages.

- **Fixture mode by default.** Wire fixture identities, state, and message delivery into the shared service. Run an event action from the chat and show the result or error in the transcript. Use this mode for repeatable acceptance without Discord credentials.
- **Opt-in Discord mode.** The Go shell can use the test bot to fetch recent messages from a configured test channel and display them in the transcript, or post a message there as the bot. Use Discord channel permissions to scope access. Keep the token server-side, bind the web server to loopback, and check that browser writes originate from the local app. [Discord message endpoints](https://docs.discord.com/developers/resources/message#get-channel-messages)

Start with fixture mode alongside the first feature that needs browser acceptance. Use the Discord client in a test server for native slash commands, permissions, acknowledgements, and message presentation. Sending a bot message and invoking a Discord slash command are separate API operations. [Discord application commands](https://docs.discord.com/developers/docs/interactions/slash-commands)
