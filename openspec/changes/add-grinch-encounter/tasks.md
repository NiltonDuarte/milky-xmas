# Tasks

## 1. Encounter behavior

- [x] 1.1 Implement shared spawn and combat decisions with explicit time and random inputs. Test inactivity, missed rolls, cooldown boundaries, shared HP, rapid snowball attacks, untimed snowballs with one player, the 60-second countdown from 75% HP, the 75% health floor, immediate phase transition, randomized defense board, per-player defense damage, boss damage from successful defenses at fixed player counts of 3, 6, and 7, the transition at 45%, the phase-three attack board, both damage values, the 5% transition, final attack, victory, and the one-minute encounter deadline.
- [x] 1.2 Implement encounter orchestration and SQLite recovery. Test the 75% health floor and single defense transition under concurrent attacks, the 45% floor and single phase-three transition under concurrent defenses, one announcement per encounter, restart, stale buttons, and failed send or edit handling.

## 2. Interfaces

- [x] 2.1 Wire the configured Discord channel, message activity, scheduler, and snowball, defense, phase-three grid, and final attack button interactions. Test transport mapping and document channel permissions and acceptance in the real server.
- [x] 2.2 Add local fixture combat and test HP, the transition at 75%, the countdown from 75% HP, and defense damage, the final attack, and the encounter deadline without Discord credentials.

- [x] 2.3 Register `/grinch summon` and verify moderator permission checks, channel restrictions, automatic eligibility overrides, and competing manual and automatic summons. Include a moderator acceptance step in the Discord test server.

- [x] 2.4 Add the 🎅 registration phase in Discord and equivalent local chat commands. Test repeated joins, the deadline, an empty roster, and rejection of unregistered attacks.

## 3. Verification

- [x] 3.1 Run Go formatting, vet, race tests, and strict OpenSpec validation. Record local and Discord acceptance evidence.

## Verification

- Go formatting, vet, and race tests passed. Strict OpenSpec validation passed.
- Fixed-clock tests exercised all phases, the shared ten-second window, the one-minute deadline, defense scaling, and victory. Two SQLite connections tested competing summons and concurrent attacks and defenses. Reopening the database retained HP and message identity, and a failed edit was retried.
- Discord transport tests verified moderator permissions, channel scoping, the duplicate summon guard, the defense grid, and button acknowledgement. The Discord API confirmed that the live bot registered the summon command in the configured guild with Manage Messages as default permission.
- The running local chat completed a fight through 100, 75, 45, 5, and 0 HP. A moderator's in-channel summon and full live combat acceptance remain pending.
- Registration update passed vet, race tests, and strict OpenSpec validation. The live local chat confirmed a single registration after repeated joins and accepted a snowball after the countdown. The Discord bot restarted in the configured channel. Clicking 🎅 on a real Discord announcement remains to be verified.
- The roster-based defense update passed vet, race tests, and strict validation. Core tests use 3, 6, and 7 registered members with only one snowball attacker and verify that all registered members count.
- The countdown update passed core and HTTP tests. A single registered player can advance after more than one minute in phase one, and the fight expires 60 seconds after reaching 75 HP.
