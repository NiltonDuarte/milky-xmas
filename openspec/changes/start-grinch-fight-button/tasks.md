# Tasks

## 1. Core

- [x] 1.1 Record the first Start press and derive the registration deadline from it, with `registration.start_window` for the wait before it. Test the countdown from the press, the idle cancel, the early start, the late join, and a fight stored before this change.

## 2. Tuning and Discord

- [x] 2.1 Add `grinch.registration.start_window` to the tuning file, its validation, and the docs.
- [x] 2.2 Show Start in the lobby and Join in a running fight, with the announcement text for before and after the first press.

## 3. Verification

- [x] 3.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate start-grinch-fight-button --strict`.
- [ ] 3.2 Discord test server: press Start from one account and check the countdown message and the start a minute later; press it from three accounts and check the fight starts at once; check the combat screen has no Join button; leave an announcement alone and check it is cancelled.
