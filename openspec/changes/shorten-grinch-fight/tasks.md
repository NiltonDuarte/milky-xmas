# Tasks

## 1. Changes

- [x] 1.1 Set the damage numbers, cap the defenses, and add the two drop end texts. Test six hits per phase for one, three, and twenty players.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate shorten-grinch-fight --strict`.
- [ ] 2.2 Discord test server: finish a fight and check six hits clear each phase; let a drop run out and read its message.
