# Tasks

## 1. Shared claim behavior

- [x] 1.1 Implement the pure eligibility decision and shared claim service. Test the first claim, early retry, exact 24-hour boundary, independent members, and a failed save with an injected clock and deterministic state.
- [x] 1.2 Implement the SQLite claim store and production configuration. Test concurrent attempts, uniqueness by server and member, and cooldown and total after reopening the database.

## 2. Discord and local chat

- [x] 2.1 Register and handle the Discord subcommand while retaining `/hello`. Mock the interaction transport to verify subcommand and identity mapping, ephemeral deferral, and the final response. Document startup with persistent storage and acceptance in the configured Discord test server.
- [x] 2.2 Add `/daily claim` to the local chat with fixture storage. Exercise first claim, cooldown, and recovery after 24 hours through the HTTP adapter with a controlled clock. Document and run the browser steps for a first claim and immediate retry.

## 3. Integration checks

- [x] 3.1 Run formatting, vet, race tests, and strict OpenSpec validation. Record observed acceptance results and any Discord server checks that still need to run.

## 4. Random daily rewards

- [x] 4.1 Award 1–10 gifts using injected randomness, persist the award with the cooldown, and render the requested success reply in both interfaces. Verify the endpoints, accumulated totals, retries, and the exact reply with deterministic tests. Update local and Discord acceptance steps.

## Verification

- The original fixed-reward flow passed a headless Chrome check for a successful claim and two immediate retries.
- Core and HTTP tests used a controlled clock to award the next gift at exactly 24 hours. SQLite tests used two connections to verify one award per round and reopened the database to verify retained totals and cooldown.
- Discord transport tests verified guild registration, member identity, private deferral, and the final reply. The executable reported missing or invalid database configuration at startup.
- Formatting, vet, race tests, and strict OpenSpec validation passed.
- Live Discord acceptance remains pending. Use the configured test server to verify registration, permissions, private replies, and cooldown after restart.

- The cooldown copy update passed Discord and HTTP adapter tests for hours, partial hours, and the last minute before eligibility.
- For random rewards, deterministic tests confirmed awards of 1 and 10, the accumulated total, the 24-hour boundary, recovery after a failed save, concurrent claims, and restart persistence. HTTP and mocked Discord tests checked the exact success reply for both endpoints of the reward range.
- In a fresh local chat in headless Chrome, `/daily claim` showed “🎅🏻 Christmas won’t be Christmas without any presents. You collected 2 gifts 🎁.” An immediate retry reported 24 hours until the next claim.
