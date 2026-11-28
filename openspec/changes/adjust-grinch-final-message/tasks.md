# Tasks

## 1. Final message

- [x] 1.1 Show only Max's image when the Grinch wins and handle the missing attach permission.
- [x] 1.2 List only players with points on the leaderboards in Discord, the standings file, and the local chat.
- [x] 1.3 Show damage, not points, on the damage leaderboard and keep final-attack points apart. Only the final attack awards points.
- [x] 1.4 Show the picture of the defeated Grinch when the group wins.

## 2. Verification

- [x] 2.1 Test the pictures for each ending, the retry without permission, the leaderboard filter, an empty leaderboard, and more than ten scorers.
- [x] 2.2 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate adjust-grinch-final-message --strict`.
- [ ] 2.3 Discord test server: end a fight by escape and check only Max's image is shown; end one with a win and check only the animated defeated Grinch; check the leaderboards list only scorers.
