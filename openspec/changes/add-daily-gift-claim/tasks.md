# Tasks

## 1. Shared claim behavior

- [x] 1.1 Implement the pure eligibility decision and shared claim service. Test the first claim, early retry, exact 24-hour boundary, independent members, and a failed save with an injected clock and deterministic state.
- [x] 1.2 Implement the SQLite claim store and production configuration. Test concurrent attempts, uniqueness by server and member, and cooldown and total after reopening the database.

## 2. Discord and local chat

- [x] 2.1 Register and handle the Discord subcommand while retaining `/hello`. Mock the interaction transport to verify subcommand and identity mapping, ephemeral deferral, and the final response. Document startup with persistent storage and acceptance in the configured Discord test server.
- [x] 2.2 Add `/daily claim` to the local chat with fixture storage. Exercise first claim, cooldown, and recovery after 24 hours through the HTTP adapter with a controlled clock. Document and run the browser steps for a first claim and immediate retry.

## 3. Integration checks

- [x] 3.1 Run formatting, vet, race tests, and strict OpenSpec validation. Record observed acceptance results and any Discord server checks that still need to run.

## Verification

- In headless Chrome, `/daily claim` awarded one gift and displayed a total of one. Two immediate retries showed the same next eligible time in UTC. The local server was restarted afterward to reset the fixture.
- Core and HTTP tests used a controlled clock to award the next gift at exactly 24 hours. SQLite tests used two connections to verify one award per round and reopened the database to verify retained totals and cooldown.
- Discord transport tests verified guild registration, member identity, private deferral, and the final reply. The executable reported missing or invalid database configuration at startup.
- Formatting, vet, race tests, and strict OpenSpec validation passed.
- Live Discord acceptance remains pending. Use the configured test server to verify registration, permissions, private replies, and cooldown after restart.
