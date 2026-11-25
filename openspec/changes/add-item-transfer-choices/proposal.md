# Proposal

## Why

Players need a way to use Yellow Snowballs and send Coal from their inventory.

## What Changes

- Offer Yellow Snowball in `/throw snowball` when the sender owns one.
- Offer Coal in `/send gift` and increase the sender's Bad deeds for each Coal sent.

## Capabilities

### New Capabilities

- `item-transfer-choices` covers the new item choices and their effects.

### Modified Capabilities

## Impact

The shared inventory service and both command adapters will change. Both items already have persistent balances.
