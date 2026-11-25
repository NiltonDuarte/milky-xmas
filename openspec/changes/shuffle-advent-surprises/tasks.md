# Tasks

## 1. Calendar

- [x] 1.1 Shuffle buttons and surprises per player, show a surprise on each click, turn opened doors green and keep them clickable, remove `/advent open` from Discord and the gift total, and add the 24 content files. Test the shuffle, the button styles and order, repeat clicks, locked doors, and that every content fits.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate shuffle-advent-surprises --strict`.
- [ ] 2.2 Discord test server: open several doors, check the surprise messages, the green buttons, repeat clicks, and that another account sees a different order.
