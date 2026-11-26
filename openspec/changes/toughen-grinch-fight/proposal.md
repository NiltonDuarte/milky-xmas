## Why

The fight ends in a few clicks, and the safe trees are shown, so a player's HP almost never matters.

## What Changes

- Attack phases take eight hits per player, never under fifteen.
- Each player waits two seconds between actions.
- The Grinch has a 10% chance of hitting back at an attacker.
- A fallen player gets back up by spending one of their own gifts: 25%, 50%, or 100% HP by rarity. When everybody is down, the Grinch waits 20 seconds before winning.
- The fight clock after the snowball phase is three minutes instead of one.

## Capabilities

### New Capabilities

- `grinch-survival`: how long the fight lasts and how players fall and get up.

### Modified Capabilities

None. The fight changes are not archived yet; these requirements replace their six-hit and one-minute rules.

## Impact

The fight panel gains revival buttons, the local chat `/grinch revive <rarity>`, and the tuning file `grinch.hits`, `grinch.click_cooldown`, `grinch.counterattack_percent`, and `grinch.revive`.
