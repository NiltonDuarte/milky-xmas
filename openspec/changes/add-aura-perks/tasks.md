## 1. Aura perks

- [x] 1.1 Core lists and perks for claim, water, rain, and fight, with tests.
- [x] 1.2 Count one send per recipient, direction, and day, kept in SQLite across restarts.
- [x] 1.3 Move the numbers to the `aura` tuning section and update replies, README, and the numbers document.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-aura-perks --strict`.
- [ ] 2.2 Discord test server: check each perk with a Nice and a Naughty account, and a repeated send.
