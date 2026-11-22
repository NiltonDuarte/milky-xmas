# Design

## Context

The repository defines FCIS package boundaries. This change provides the first shared request path through them.

## Goals and scope

The Discord bot and local web application have distinct entry points at `src/cmd/discord` and `src/cmd/web`, with separate adapters under `src/internal/adapter`. Both entry points use the stateless application service and core under `src/internal`. The local chat uses session-local transcript rendering and runs independently of the Discord executable.

## Decisions

- Both adapters give a command name to the application service, which dispatches `hello` to a pure greeting rule and returns a reply or an unknown-command error. The core has no inputs or effects for this greeting.
- The Discord adapter uses DiscordGo for the Gateway connection, guild command registration, and interaction responses. Guild registration makes the command available in the configured test server. Registration upserts the one command so repeated starts preserve other application commands. The reply is the initial interaction acknowledgement.
- The web adapter serves an embedded page and accepts local form submissions through a Go HTTP handler. Browser code renders user and bot entries as text. The HTTP listener binds to loopback and the handler uses Go's cross-origin protection for writes.

## Risks and trade-offs

Automated tests exercise the Discord request and response mapping with a fake HTTP transport. Registration, permissions, timing, and presentation still need an installed test bot for acceptance in the Discord client.
