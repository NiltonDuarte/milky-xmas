# Proposal

## Why

The Milky Christmas event needs a reward that members can return to collect. A daily gift gives members a reason to come back and starts their personal collection.

## What Changes

- Add `/daily claim` to Discord and the local test chat.
- Award one gift per successful claim and report the member's accumulated total.
- Make the next gift available 24 hours after the last successful claim. Early attempts show when the next gift is available.
- Keep production claims and totals across bot restarts.

## Capabilities

### New Capabilities

- `daily-gift-claim` governs gift collection, the cooldown, and member totals.

### Modified Capabilities

## Impact

Adds a stateful shared Go use case, production storage, a Discord slash subcommand, and a fixture-backed chat path.
