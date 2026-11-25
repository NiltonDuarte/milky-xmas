## Why

The fight needed about forty clicks, too many for a small group within its timers. A drop that ran out only said it expired.

## What Changes

- Each damage phase takes at most six hits from the group, however many players there are.
- The defense phase needs one safe defense per player, at most six (`grinch.defense.max_defenses`).
- A drop whose time runs out says the Grinch stole all the presents. A Gift Rain used up by ten collections says all the gifts were collected.

## Impact

- Core: new default damage numbers and a cap on the required defenses. `tuning.yaml` follows.
- Both adapters: the drop end texts.
