# Proposal

## Why

Milky Xmas needs a spontaneous cooperative event that gives members something to do together when their Discord channel is active.

## What Changes

- Let the Grinch appear at random in one configured Discord channel following recent human activity.
- Start the fight with a 10-second snowball phase shared by everyone in the channel. Each eligible press of the snowball button deals one damage to shared health, down to 75% of maximum HP. Switch immediately to the defense phase when shared health reaches 75%.
- Show eight randomized defense buttons in two rows of four. A Christmas tree is safe and a snowflake deals 10% damage to the player who selects it. The encounter is planned for three phases.
- Provide a deterministic local encounter fixture to test the same combat rules.

## Capabilities

### New Capabilities

- `grinch-encounter` covers activity-based appearances and cooperative combat.

### Modified Capabilities

## Impact

Adds a shared encounter service, an encounter store and scheduler, Discord message and button handling, and local fight controls.
