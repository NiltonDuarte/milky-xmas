# Tasks

## 1. Timer

- [x] 1.1 Give the snowball phase a deadline from `grinch.snowballs.duration`, with validation and the tuning file entry. Test the timeout, the on-time milestone, a tuned timer, and a refused late click.
- [x] 1.2 Update the snowball message text, the local chat test, the docs, and the README.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-snowball-phase-timer --strict`.
- [ ] 2.2 Discord test server: start a fight and wait out two minutes in the snowball phase, check the Grinch wins with Max's image and the loss rewards; start another and reach 75 HP in time.
