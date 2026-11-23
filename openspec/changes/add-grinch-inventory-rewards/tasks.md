# Tasks

## 1. Implementation

- [x] 1.1 Add core Snowball and loss rules. Test valid picks, stale actions, and participant-based rewards.
- [x] 1.2 Persist encounter rewards and inventory atomically. Verify retry, restart, and concurrent writes with SQLite and fixtures.
- [x] 1.3 Render the new action and inventory balance and verify both adapter flows.

## 2. Verification

- [x] 2.1 Run formatting, vet, race tests, and strict validation. Record acceptance evidence.

## Acceptance evidence

- The local chat HTTP test reached 5% HP, accepted two snowflake picks, and showed two Snowballs in `/socks`. A victory awarded no loss items. A later solo escape credited one Coal and one Yellow Snowball.
- The SQLite test ran the same snowflake request against two connections and verified one credit. It then confirmed that a failed participant write rolls back the ending and all rewards. Retrying and reopening left all four participants with exactly one Coal and four Yellow Snowballs each.
- Discord component tests verified the snowflake control on both current and legacy final rows. Live Discord acceptance awaits deployment.
- Formatting, vet, the full race suite, and all strict OpenSpec checks passed. The race suite used a temporary Zig C compiler.
