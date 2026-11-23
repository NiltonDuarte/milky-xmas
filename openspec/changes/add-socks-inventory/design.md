# Design

## Context

Both adapters share a player record with per-rarity gift balances. Grinch snowball actions and Advent rewards have their own state.

## Decisions

- Load the inventory through the existing player store. Both adapters ask the shared service for the invoking player's inventory and use the same list format.
- Add Snowball and Coal balances to the player inventory and preserve them in claims and transfers. The new balances default to zero. Item acquisition and consumption stay with their respective gameplay changes.
- Add the columns in the existing SQLite startup transaction. Existing balances and timestamps carry forward. Discord reuses deferred private replies.
