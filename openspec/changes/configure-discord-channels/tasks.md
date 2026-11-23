# Tasks

## 1. Setup and routing

- [x] 1.1 Add durable guild settings and verify authorization, introduction retry, and restart behavior.
- [x] 1.2 Add Discord onboarding, setup commands, permission gates, and per-channel encounter scheduling with simulated interaction tests.

## 2. Verification

- [x] 2.1 Run formatting, vet, race, and strict OpenSpec validation and document acceptance.

## Acceptance evidence

- The simulated Discord gateway flow posted one onboarding prompt in the system channel, rejected a non-moderator, and accepted the channel settings from an unselected channel.
- The simulation allowed gameplay in a command-only channel, blocked gameplay in a boss-only channel, rejected manual summons in both cases, and restored the boss worker from saved settings. Disabling a role removed access.
- The simulated Discord request posted the exact five-line introduction in the selected boss channel. SQLite tests preserved settings across reopen, retried a failed message, and skipped re-announcing after success.
- Formatting, vet, the full race suite, and all strict OpenSpec checks passed. The race suite used a temporary Zig C compiler.
- Live onboarding and channel permissions still need acceptance in a Discord test server. The README contains the moderator flow.
