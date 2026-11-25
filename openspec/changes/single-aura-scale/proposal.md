## Why

Aura was two independent counters, so a player could be 100% bad and 100% good at once. The community wants one position between bad and good that every action moves.

## What Changes

- Aura is one number from -50 (bad deeds) to +50 (good deeds). Everyone starts at 0.
- Sending gifts moves it toward good by 5, 10, or 30 points per normal, rare, or epic gift. Coal and the yellow snowball move it toward bad by 5 points each.
- `/aura check` draws the scale with a marker and the number under it.
- The Random Active Drop bonus applies from +25 (`gifts.good_deeds_boost.threshold`).
- Existing players are reset to 0.
- `/send gift` no longer reports percentages: gifts say Good deeds are going up, Coal warns that Bad deeds are increasing, and there is a space after the gift emoji.

## Impact

- Core: one aura type with a single move rule.
- Storage: a new `aura` column. The old `bad` and `good` columns stay unused in existing databases.
- Replaces the percentage rules of `add-gift-transfer-and-aura`, `add-item-transfer-choices`, `transfer-normal-snowball`, and the Good deeds threshold of `update-activity-drops`.
