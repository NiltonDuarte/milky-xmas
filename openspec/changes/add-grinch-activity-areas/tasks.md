# Tasks

## 1. Events

- [x] 1.1 Record player and Grinch events in the core with a roll from the shell, keep the latest five, and clone them in the fixture store. Test every kind, the cap, the start events, and the defeat.
- [x] 1.2 Supply the roll in `Play`. Test that the shell passes it and stores the events.

## 2. Text and layout

- [x] 2.1 Write the pools and lines in `grinchtext`, including the player's hiding lines. Test the pools, the roll, and the text.
- [x] 2.2 Add the two areas to the Discord panel and the local chat. Test placeholders, events, the lobby and final messages, and the component count.

## 3. Verification

- [x] 3.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-grinch-activity-areas --strict`.
- [ ] 3.2 Discord test server: run a fight and check the lines for hits, misses, hiding behind the tree, gifts and Snowballs, and that the areas disappear in the final message.
