# Tasks

## 1. Inventory

- [x] 1.1 Extend the shared player record and SQLite storage. Test migration, nonzero item balances, transfers, and reopening.
- [x] 1.2 Register and route `/socks` through both adapters. Test inventory display, private Discord acknowledgement, and read failures. Document acceptance steps.

## 2. Verification

- [x] 2.1 Run format, vet, race tests, and strict validation. Record local and Discord acceptance evidence.

## Acceptance evidence

- Local executable over loopback HTTP showed five zero balances for a new player. A daily claim awarded six normal gifts. After sending one gift to `friend`, `/socks` showed five normal gifts and zero for the other items.
- SQLite tests verified the migration preserves existing balances, aura, and cooldown. Nonzero Snowballs and Coal survived a gift transfer and reopening the database.
- Simulated Discord interactions verified that the standalone command registers without options, uses a deferred private acknowledgement, and shows each participant their own balance after a transfer.
- Live Discord acceptance is pending. The README describes the two-account check.
- Formatting, vet, the full race suite, and strict OpenSpec validation passed. The race suite used a temporary Zig C compiler.
