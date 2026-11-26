## 1. Watering reward and presentation

- [x] 1.1 Add `tree.water_aura` and move the player's aura after a counted watering. Test the move, the refusal, the top of the scale, and a failed save.
- [x] 1.2 Draw the water bar with squares, box the tree reply, pick the picture by level, and defer `/water`. Test the call order, each picture boundary, and the failed channel post.

## 2. Verification

- [x] 2.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate reward-tree-watering --strict`.
- [ ] 2.2 Discord test server: run `/water`, check the box, the bar on desktop and phone, `/aura check`, and the private refusal.
- [ ] 2.3 Add a usable picture for 51% to 99%.
