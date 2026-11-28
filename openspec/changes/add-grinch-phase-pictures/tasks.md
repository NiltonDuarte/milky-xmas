## 1. Pictures

- [x] 1.1 Show the picture of each phase and upload it only when it changes. Test every phase change and a repeated update.
- [x] 1.2 Replace the final attack picture with the escape picture (`grinchfinal.jpg`).

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-grinch-phase-pictures --strict`.
- [x] 2.2 Discord test server: play a fight and check the animation in the defense phase and the heart picture at 5%.
- [ ] 2.3 Discord test server: check the escape picture at 5%.
