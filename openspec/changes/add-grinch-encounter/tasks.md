# Tasks

## 1. Encounter behavior

- [ ] 1.1 Implement shared spawn and combat decisions with explicit time and random inputs. Test inactivity, missed rolls, cooldown boundaries, shared HP, rapid snowball attacks, late joiners, the exact phase deadline, the 75% health floor, immediate phase transition, randomized defense board, per-player defense damage, boss damage from successful defenses at fixed player counts of 3, 6, and 7, the transition at 45%, the phase-three attack board, both damage values, the 5% transition, final attack, victory, and the one-minute encounter deadline.
- [ ] 1.2 Implement encounter orchestration and SQLite recovery. Test the 75% health floor and single defense transition under concurrent attacks, the 45% floor and single phase-three transition under concurrent defenses, one announcement per encounter, restart, stale buttons, and failed send or edit handling.

## 2. Interfaces

- [ ] 2.1 Wire the configured Discord channel, message activity, scheduler, and snowball, defense, phase-three grid, and final attack button interactions. Test transport mapping and document channel permissions and acceptance in the real server.
- [ ] 2.2 Add local fixture combat and test HP, the transition at 75%, the shared phase deadline, and defense damage, the final attack, and the encounter deadline without Discord credentials.

- [ ] 2.3 Register `/grinch summon` and verify moderator permission checks, channel restrictions, automatic eligibility overrides, and competing manual and automatic summons. Include a moderator acceptance step in the Discord test server.

## 3. Verification

- [ ] 3.1 Run Go formatting, vet, race tests, and strict OpenSpec validation. Record local and Discord acceptance evidence.
