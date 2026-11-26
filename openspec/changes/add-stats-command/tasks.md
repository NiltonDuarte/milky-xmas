## 1. Stats

- [x] 1.1 Count each player's activity from the records and rank it. Test points, ties, receiving, other servers, and fights.
- [x] 1.2 Add `/stats` with drop-downs in Discord and arguments in the local chat. Test the call order, each view, and a failed read.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-stats-command --strict`.
- [ ] 2.2 Discord test server: open `/stats`, switch between views and each leaderboard, on desktop and phone.
- [ ] 2.3 Decide how to show the Grinch fight leaderboards.
