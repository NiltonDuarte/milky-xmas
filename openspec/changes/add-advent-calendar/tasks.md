# Tasks

## 1. Calendar behavior

- [x] 1.1 Implement shared Advent rules and persistence. Test catch-up, December boundaries, duplicate opens, year and member isolation, persistence, and daily balance separation.

## 2. Interfaces

- [x] 2.1 Register and handle Discord Advent commands and map typed equivalents in the test chat. Test both adapter paths and document acceptance steps.

## 3. Verification

- [x] 3.1 Run formatting, vet, race tests, and strict OpenSpec validation. Record local acceptance and pending Discord checks.

## Verification

- Headless Chrome showed the 24 doors and rejected opening door 1 before December.
- Fixed-clock tests exercised catch-up, UTC boundaries, duplicate opens, and daily gift independence. SQLite tests verified concurrent requests and restart persistence. Discord transport tests verified command registration and private responses.
- Formatting, vet, race tests, and strict OpenSpec validation passed. Acceptance in a configured Discord server remains pending.
