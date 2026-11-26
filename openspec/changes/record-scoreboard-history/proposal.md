## Why

At the end of the event the developer wants scoreboards for everything. Only the latest Grinch fight of each channel is stored, and inventory changes keep no trace of what caused them, so that information is lost as the event goes on.

## What Changes

- The bot's database keeps every finished Grinch fight.
- It keeps each inventory or aura change with its reason and, for transfers, the other player.
- Nothing new is shown. Scoreboards are a later change.

## Capabilities

### New Capabilities

- `scoreboard-history`: the records the final scoreboards will be built from.

### Modified Capabilities

None.

## Impact

The shared gift use cases name why they change an inventory, and the SQLite store adds two tables.
