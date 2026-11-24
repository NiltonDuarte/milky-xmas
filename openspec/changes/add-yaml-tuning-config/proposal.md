# Proposal

## Why

The balance numbers in `docs/activity-drops-numbers.md` and `docs/grinch-fight-numbers.md` are spread over Go constants and literals. Tuning one means editing code, rebuilding, and hunting for the right file. The team wants one reviewable YAML file where these numbers are checked and adjusted.

## What Changes

- Add a YAML tuning file that holds the numbers documented for the two activity drops, the gift rarity odds they use, and the Grinch fight.
- Both executables (`src/cmd/discord` and `src/cmd/web`) load the file when `MILKY_TUNING_CONFIG` names one. Without it they use the built-in values, which equal today's numbers, so nothing changes unless a file is supplied.
- Reject an invalid file at startup with a message that lists every problem, instead of running with a broken balance. Unknown keys are errors, and omitted keys keep their default.
- Ship `tuning.yaml` at the repository root with every value at its current default and a comment per value.
- Move the rules that use these numbers (rarity rolls, drop thresholds, Grinch HP phases, damage, timers, rewards) so they read from the tuning instead of constants. Player-facing text that quotes a number (for example "75 HP" or "2 damage points") follows the configured value.
- Update both docs to point at the YAML keys.

Out of scope: numbers not in those docs (daily claim amount and cooldown, Advent window, Christmas tree, Aura increments), structural numbers (eight gift buttons, the eight-tile boards, the five-button final row, the 100% boss total), and reloading the file without a restart.

## Capabilities

### New Capabilities

- `yaml-tuning-config` covers loading, defaults, validation, and effect of the tuning file in both interfaces.

### Modified Capabilities

## Impact

`core` gains pure tuning types, defaults, and validation, and its rules take the tuning as an input. `app` carries the tuning in `Service`, `ActivityDrops`, and `GrinchEncounter`. A new adapter package reads the YAML file. `cmd` composes the loaded tuning. The Discord and web adapters read durations and numbers from the tuning for their timers and messages. Adds the `go.yaml.in/yaml/v3` dependency.
