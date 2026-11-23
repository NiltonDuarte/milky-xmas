# Proposal

## Why

Players need a single command to check which gifts and event items they own.

## What Changes

- Add `/socks` to display the invoking player's normal, rare, and epic gifts, Snowballs, and Coal.
- Store Snowball and Coal balances, starting at zero for existing and new inventories.

## Capabilities

### New Capabilities

- `socks-inventory` covers the player's inventory view.

### Modified Capabilities

## Impact

Both adapters will call the shared application service. SQLite will add two balances to the player record.
