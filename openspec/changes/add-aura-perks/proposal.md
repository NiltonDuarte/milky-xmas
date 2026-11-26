## Why

The aura barely matters: only the good side gets anything, and it is a Random Active Drop bonus players hardly notice. Two players can also push both auras to the top by passing the same gift back and forth.

## What Changes

- Players at +25 or more are on the Nice list and players at -25 or less on the Naughty list, each with perks.
- A send moves the sender's aura once a day per recipient and per direction.
- The aura rises and falls at most 15 points a day.
- The daily claim adds Coal for players at aura zero or below, and watering is a good deed only from zero up, so the Naughty list is reachable.
- The numbers move to an `aura` section of the tuning file. **BREAKING**: `gifts.good_deeds_boost.threshold` is replaced by `aura.nice_at`.

## Capabilities

### New Capabilities

- `aura-perks`: the two lists, their perks, and when a send counts for the aura.

### Modified Capabilities

None. `aura-scale` and the drop, tree, claim, and fight changes are not archived yet; the requirements here add to them.

## Impact

`/daily claim`, `/water`, `/send gift`, `/throw snowball`, activity drops, the Grinch fight, the tuning file, and a new `aura_sends` table.
