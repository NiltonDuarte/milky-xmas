# Proposal

## Why

Active event channels need lightweight, shared rewards that make ordinary chat participation feel festive without allowing one player to exhaust a drop.

## What Changes

- Add Random Active Drops that can appear from eligible channel activity, render two rows of four collectible gift buttons, and let each player claim only one gift from a drop.
- Add Gift Rain Drops after a random five to one hundred qualifying chat messages, with two rows of four buttons and one claim per player per minute while the drop remains available.
- Award a random gift rarity to each accepted claim and make the active-drop rules available through the shared application layer.

## Capabilities

### New Capabilities

- `activity-gift-drops`: shared reward eligibility, claiming, and Discord presentation for activity-driven gift drops.

### Modified Capabilities

None.

## Impact

The core, shared application service, Discord message/event adapter, SQLite and fixture storage, and local acceptance adapter will gain activity-drop behavior and tests.
