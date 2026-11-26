## 1. History

- [x] 1.1 Keep finished fights and inventory changes in SQLite. Test transfers, a stale save, rewards, repeated saves of a finished fight, and a cancelled fight.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate record-scoreboard-history --strict`.
- [ ] 2.2 Discord test server: finish a fight and send a gift, then check both are in the database.
