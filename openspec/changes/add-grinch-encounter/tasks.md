# Tasks

## 1. Encounter behavior

- [ ] 1.1 Implement shared spawn and combat decisions with explicit time and random inputs. Test inactivity, missed rolls, cooldown boundaries, shared HP, rapid snowball attacks, late joiners, the exact phase deadline, the 75% health floor, immediate phase transition, randomized defense board, per-player defense damage, and escape.
- [ ] 1.2 Implement encounter orchestration and SQLite recovery. Test the 75% health floor and single defense transition under concurrent attacks, one announcement per encounter, restart, stale buttons, and failed send or edit handling.

## 2. Interfaces

- [ ] 2.1 Wire the configured Discord channel, message activity, scheduler, and snowball and defense button interactions. Test transport mapping and document channel permissions and acceptance in the real server.
- [ ] 2.2 Add local fixture combat and test HP, the transition at 75%, the shared phase deadline, and private defense damage without Discord credentials.

## 3. Verification

- [ ] 3.1 Run Go formatting, vet, race tests, and strict OpenSpec validation. Record local and Discord acceptance evidence.
