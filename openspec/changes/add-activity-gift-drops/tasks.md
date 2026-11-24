# Tasks

## 1. Shared drop behavior

- [x] 1.1 Implement deterministic core rules for Random Active Drop and Gift Rain eligibility, activity thresholds, and per-player limits, with table-driven unit tests.
- [x] 1.2 Implement application orchestration for activity observation and compare-and-save gift awards, with concurrency tests proving a duplicate claim awards once.

## 2. Discord presentation

- [x] 2.1 Observe eligible human messages, announce the two requested drop messages, and render two rows of four gift controls, verified by Discord adapter tests.
- [x] 2.2 Route drop button interactions through the shared claim behavior and verify one-per-drop and hidden Gift Rain cooldown handling in adapter tests.

## 3. Verification

- [x] 3.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-activity-gift-drops --strict`.
- [x] 3.2 Record repeatable Discord acceptance steps and observed results for enabled-channel drops, first claim, duplicate claim, and Gift Rain cooldown behavior.

## Acceptance evidence

- Automated Discord adapter acceptance configured channel `400` and left `401` disabled. Six messages in `401` produced no card. The fifth message in `400` posted `🎁 It's raining gifts! 🎁 Hallelujah` with two rows of four gift buttons.
- The simulated player collected a normal gift from a Gift Rain then clicked again immediately. The second private response was `You cannot collect another gift yet.` and contained no remaining-time value.
- Local browser acceptance sent `/message festive chat` five times, collected `/drop collect 1`, and confirmed the same hidden cooldown response.
- Repeat in the Discord test server after deployment: enable a command channel with `/setup channel`, send messages until either drop appears, collect one gift, repeat the click, then verify `/socks` increased once. Use a second player to confirm each player has independent eligibility.

## Workflow follow-up

- Archive the change after review and acceptance are complete.
