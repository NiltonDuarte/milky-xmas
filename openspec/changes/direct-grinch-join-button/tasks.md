# Tasks

## 1. Join button

- [x] 1.1 Replace "Open registration" with a Join button on the fight message, and update the announcement text.
- [x] 1.2 Register the player in one click: acknowledge silently, run the lobby move, send one private confirmation only when the player joins now.
- [x] 1.3 Remove the private registration panel and its state. Keep old panels working.

## 2. Verification

- [x] 2.1 Test first click, repeat click, a second player, and the distinct players registered by the fight. Test that a store failure sends a private follow-up after the acknowledgement.
- [x] 2.2 Update the README. Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate direct-grinch-join-button --strict`.
- [ ] 2.3 Discord test server: press Join from two accounts, press it again, and check only the first click shows a private message.
