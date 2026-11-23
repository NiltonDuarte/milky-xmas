# Proposal

## Why

Milky Xmas needs a spontaneous cooperative event that gives members something to do together when their Discord channel is active.

## What Changes

- Let the Grinch appear at random in one configured Discord channel following recent human activity.
- Give moderators a `/grinch summon` command to start an encounter manually in that channel.
- Announce the upcoming encounter with a 🎅 join button and start combat after three members join or after one minute, and keep joining open during the fight.
- Start the fight with an untimed snowball phase shared by everyone in the channel. Each eligible press of the snowball button deals one damage to shared health, down to 75% of maximum HP. Switch immediately to the defense phase when shared health reaches 75%.
- Show eight randomized defense buttons in two rows of four, and reshuffle after each accepted defense choice. A Christmas tree protects the player and damages the boss for 30 percentage points divided by the number of players who joined with 🎅, down to a minimum of 45% of maximum HP. A snowflake deals 10% damage to the player who selects it. Switch to phase three at 45% HP. Phase three starts with a Grinch shout and a randomized grid of eight attack buttons. ☃️ deals base damage and 🧊 deals 5% more. Attacks can reduce the boss to 5% HP.
- At 5% HP, replace the grid with a randomized row of attack and gift buttons until the Grinch is defeated. Allow one player or a group to win. End the fight in defeat if every registered player reaches zero HP, or let the Grinch escape if the one-minute countdown expires.
- Show separate damage and gift leaderboards on the final result.
- Provide a deterministic local encounter fixture to test the same combat rules.

## Capabilities

### New Capabilities

- `grinch-encounter` covers activity-based appearances and cooperative combat.

### Modified Capabilities

## Impact

Adds a shared encounter service, an encounter store and scheduler, Discord message and button handling, and local fight controls.
