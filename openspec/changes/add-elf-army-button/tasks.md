# Tasks

## 1. Button and role

- [x] 1.1 Add the button to the Christmas introduction, create the `elfs` role at setup, and handle the click with a deferred private reply. Test the setup message, the order of acknowledgement and role calls, a repeated click, and a refused role change.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-elf-army-button --strict`.
- [ ] 2.2 Discord test server: run `/setup channel` with the boss on for a new channel, check the introduction shows the button, click it, check the role appears, click again, and check the second reply says you are already in.
