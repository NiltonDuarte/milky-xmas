# Tasks

## 1. Core tuning

- [x] 1.1 Add `core.Tuning`, `DefaultTuning()`, and `Validate()` with table-driven tests for every range and the HP ordering.
- [x] 1.2 Make the rarity, drop, and Gift Rain rules read the tuning. Keep the existing tests passing with the defaults and add cases with changed values.
- [x] 1.3 Make the Grinch rules read the tuning (spawn, registration, duration, milestones, damage, defense miss, final attack, rewards, scores). Add tests with changed values and with the defaults.

## 2. Application layer

- [x] 2.1 Carry the tuning in `Service`, `ActivityDrops`, and `GrinchEncounter`, with defaults when none is given. Test the shell with a changed tuning, including the spawn roll and loss rewards.

## 3. File and executables

- [x] 3.1 Add the `tuningfile` adapter with YAML parsing, defaults overlay, unit conversion, and error messages that list all key paths. Test partial files, unknown keys, bad units, bad ranges, several errors, and the shipped `tuning.yaml` against the defaults.
- [x] 3.2 Load `MILKY_TUNING_CONFIG` in `src/cmd/discord` and `src/cmd/web`. Ship `tuning.yaml` with every default and a comment per value.

## 4. Adapters and docs

- [x] 4.1 Render timers and number-bearing messages from the tuning in Discord and the local chat. Test default output is unchanged and a changed tuning shows the new numbers.
- [x] 4.2 Update both docs with the YAML key of each number, and the README with the tuning file and acceptance steps.

## 5. Verification

- [x] 5.1 Run `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, and `openspec validate add-yaml-tuning-config --strict`. Record acceptance evidence.

## Acceptance evidence

- Automated: core tests cover every validation range, the rarity bands, the Gift Rain range and Coal chance, a full Grinch fight with every number changed (early start, milestones, damage, defense miss, gift points, rewards, messages), the spawn rules, and the loss rewards. App tests cover a tuned Gift Rain (range, end, cooldown), drop lifetime, Coal chance, and boosted odds. The `tuningfile` tests cover partial files, unit conversion, unknown keys, bad units, a decimal in a whole-number key, three decimals, several problems at once, a missing file, and that the shipped `tuning.yaml` equals the defaults. Adapter tests cover the Discord message numbers, the drop lifetime, the SQLite-backed spawn chance and loss rewards, and the local chat with a changed file. All existing tests pass unchanged in behavior with the defaults.
- Observed in the local chat (`MILKY_TUNING_CONFIG` pointing at a copy with `claims_to_end: 2`, rain range 2–2, `early_start_players: 1`, `snowballs.boss_hp_after: 90`): two `/message` start the rain, the second `/drop collect 1` ends it and lists both collections, a third is refused, one `/grinch join` starts combat, and ten `/grinch hit 1` reach the defense phase at 90 HP. A file with an unknown key and a missing file stop the executable at startup with a message naming the problem.
- Pending: Discord test server check of the announcement text (`0/1 to start early`, `reaches **90 HP**`), the timer interval, and a Gift Rain that ends on the configured collection. Steps are in the README under "Tuning the numbers".
