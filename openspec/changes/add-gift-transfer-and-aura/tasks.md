# Tasks

## 1. Shared rules and persistence

- [x] 1.1 Implement rarity, aura arithmetic, and inventory transfer in the core with deterministic tests for quantities, rates, cooldown, and rejected transfers.
- [x] 1.2 Implement shared commands and atomic fixture and SQLite storage. Verify contention, rollback, migration, cooldown, and restart with integration tests.

## 2. Command adapters

- [x] 2.1 Register and handle both Discord commands and verify options, private replies, and bot rejection with simulated interactions.
- [x] 2.2 Route both commands through the browser chat. Verify the fixture flow using HTTP and document reproducible acceptance for both interfaces.

## 3. Integration

- [x] 3.1 Run formatting, vet, race tests, and strict OpenSpec validation. Review the changed command paths and record observed and pending acceptance.

## Acceptance evidence

- Local web executable over loopback HTTP. `/aura check` showed 0% on both sides. `/daily claim` awarded seven rare gifts. Sending all seven to `friend` showed 70% Good deeds. A second send was rejected, the aura stayed at 70%, and another claim reported the original cooldown.
- Deterministic SQLite tests verified that legacy gifts migrate as normal gifts, both inventories and aura survive reopening, a failed recipient write rolls back the sender's changes, and only one competing transfer spends the final gift.
- Simulated Discord interactions verified the registered options, bot rejection, a private deferred acknowledgement followed by a private result, and independent aura for the sender and recipient.
- Live Discord acceptance remains pending because this environment has no test server credentials. The README lists the two-account steps for checking command permissions, presentation, and sending a received gift back.
- Formatting, vet, strict OpenSpec validation, and the full race suite passed. The race suite used a temporary Zig C compiler with CGO enabled.
