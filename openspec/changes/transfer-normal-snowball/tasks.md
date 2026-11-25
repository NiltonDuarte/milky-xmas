# Tasks

## 1. Transfer

- [x] 1.1 Add `TransferSnowball`, route normal throws through the shared transfer, add the refusal text, and offer the normal type only when owned. Test the rule, the app flow (two throws, none left, full recipient, failed save), the Discord interaction and suggestions, and the local chat.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate transfer-normal-snowball --strict`.
- [ ] 2.2 Discord test server: with one Snowball, throw it at another player and check both inventories with `/look in-your-stocking`; then try again with none and check the refusal.
