# Tasks

## 1. Aura scale

- [x] 1.1 Replace the two counters with one aura that stays between -50 and +50, move the drop threshold to +25, add the storage column, and draw the scale. Test the move rule at both ends, each action, the upgrade from the old columns, the drop boundary, and the drawing.
- [x] 1.2 Shorten the `/send gift` messages and update the README.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate single-aura-scale --strict`.
- [ ] 2.2 Discord test server: run `/aura check` (expect 0), send a gift and Coal, and check the marker and number move and line up on desktop and phone.
