# Proposal

## Why

Players can see their Snowball balance in `/socks`. Throwing a snowball gives players a way to add to another player's inventory.

## What Changes

- Add `/thrown snowball` with a recipient selector.
- Credit one Snowball to the recipient and tag the recipient in the channel where the throw happened.

## Capabilities

### New Capabilities

- `snowball-throw` covers crediting a snowball and notifying the recipient.

### Modified Capabilities

## Impact

The shared application service and both adapters will change. Snowballs use the existing inventory storage.
