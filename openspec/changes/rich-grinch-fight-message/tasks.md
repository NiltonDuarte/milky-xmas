# Tasks

## 1. Core

- [x] 1.1 Count the Snowballs each player collects and add a Snowball leaderboard, with tests including old stored fights.

## 2. Discord layout

- [x] 2.1 Build the Components V2 layout: image, text, the joined list, the buttons, cards, summary, and final leaderboards. Test each phase, the 8-card cap, the component count, and the top-ten cut.
- [x] 2.2 Send and edit the fight message as V2 with no mentions. Keep the image and standings attachment behaviour.

## 3. Local chat and docs

- [x] 3.1 Add the Snowball leaderboard to the local chat and update the README.

## 4. Verification

- [x] 4.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate rich-grinch-fight-message --strict`.
- [ ] 4.2 Discord test server: run a fight with two or more accounts and check the image, the joined list, live HP and cards, the final leaderboards, and that a fight already running at the update does not break.
