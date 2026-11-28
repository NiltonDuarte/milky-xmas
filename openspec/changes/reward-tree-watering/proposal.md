## Why

Watering helps the whole community but gives the player nothing, the water bar looks plain, the reply is split in two embeds, and every level shows the withered tree.

## What Changes

- Each watering that counts moves the player's aura 1 point toward good deeds (`tree.water_aura`).
- The water level is a row of ten emoji squares.
- The tree reply is one box with the picture and the text. The top waterers list is removed; waterings are still counted.
- The picture follows the level, with four pictures: up to 25%, 26% to 50%, 51% to 75%, and 76% to 100%.
- Discord acknowledges `/water` before doing any work.

## Capabilities

### New Capabilities

- `tree-watering-reward`: the good deed for watering and how the tree is shown.

### Modified Capabilities

None. The `christmas-tree` capability is not archived yet, so its presentation requirement is edited in `add-christmas-tree`.

## Impact

Core tree and tuning rules, the tree use case, both adapters, `tuning.yaml`, and two new embedded images.
