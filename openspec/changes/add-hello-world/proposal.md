# Proposal

## Why

Milky Xmas needs a first runnable bot command to verify that Discord and the local test interface can exercise the same Go behavior.

## What Changes

- Add a guild-scoped Discord slash command, `/hello`, that replies with `Hello, world!`.
- Add a loopback-only chat with a composer and transcript for testing `/hello` without credentials.
- Document startup from the Go toolchain and acceptance steps for both interfaces.

## Capabilities

### New Capabilities

- `hello-command` provides a greeting on Discord and in the local test chat.

### Modified Capabilities

## Impact

Adds Go entry points, a shared greeting use case and rule, a DiscordGo adapter, and a local HTTP adapter.
