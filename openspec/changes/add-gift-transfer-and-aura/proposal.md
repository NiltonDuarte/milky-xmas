# Proposal

## Why

Players can collect Christmas gifts but cannot give them to each other. Gift transfers give players a way to share rewards and build a personal aura.

## What Changes

- Add `/send gift` with recipient, rarity, and quantity.
- Add `/aura check` with Bad deeds on the left and Good deeds on the right.
- Give daily claim rewards a rarity and preserve existing gifts as normal gifts.

## Capabilities

### New Capabilities

- `gift-transfers` covers transferable gift inventory, transfers, and player aura.

### Modified Capabilities

## Impact

The shared Go service, SQLite storage, local fixtures, and both command adapters will change. Advent gifts and Grinch gift points remain separate event rewards.
