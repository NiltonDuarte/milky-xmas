# Design

## Context

The numbers live as constants in `core` (`activity_drops.go`, `gifts.go`, `grinch.go`), in `app/activity_drops.go`, and as literals in `core/grinch.go`, `app/grinch.go`, and the Discord adapter. The Grinch methods on `core.Grinch` do arithmetic in hundredths of HP, and the boss HP, phase milestones, and damage values are written inline.

## Decisions

- **Where the types live.** `core.Tuning` is plain data with `DefaultTuning()` and a pure `Validate()` that joins all problems with their key paths. It has three parts: `Gifts` (daily and boosted odds), `Drops`, and `Grinch` (`GrinchRules`). `core` never reads files or YAML.
- **How the rules receive it.** Pure functions and `Grinch` methods take the tuning as an explicit argument (`Move`, `Advance`, `Deadline`, `CanJoin`, `CanSpawn`, `Health`, `Leaderboard`, `LossRewards`, the rarity and drop helpers). I chose this over a hidden field on `Grinch` so a missed call site fails to compile instead of silently using defaults.
- **Where it is carried.** `app.Service` keeps the tuning (`NewService` uses the defaults, `WithTuning` replaces them), and `ActivityDrops` reads it from the service. `app.NewGrinchEncounter` takes the `GrinchRules` and exposes them with `Rules()` so the Discord and web adapters render timers and messages from the same values.
- **Where the file is read.** A new adapter package `src/internal/adapter/tuningfile` reads the YAML (`go.yaml.in/yaml/v3`, known fields only), overlays it on `DefaultTuning()`, converts units, and calls `Validate()`. `cmd` composes it: `MILKY_TUNING_CONFIG` unset means defaults, set means load or exit with the error. The file is read once at startup.
- **Units in the file.** Durations are Go duration strings (`10m`, `90s`). Percentages and HP are numbers: odds are whole numbers that add up to 100, boss HP and damage allow up to two decimals (`5.25`). The loader converts HP and damage to hundredths and rejects a third decimal. Snowball damage is clamped at the phase milestone, so it does not have to divide the distance evenly.
- **Key layout.** Keys mirror the docs: `gifts.daily_rarity`, `gifts.good_deeds_boost`, `drops.lifetime`, `drops.random_active.one_in`, `drops.gift_rain.{min_messages,max_messages,claims_to_end,claim_cooldown,coal_percent}`, `grinch.spawn.*`, `grinch.registration.*`, `grinch.fight_duration`, `grinch.snowballs`, `grinch.defense`, `grinch.attack`, `grinch.final_attack`, `grinch.loss_rewards`. The Gift Rain range is written as a minimum and a maximum (inclusive) instead of the internal `5 + 0..95`.
- **Defense math.** The defense phase removes `snowballs.boss_hp_after - defense.boss_hp_after` HP in total, split evenly across the players who registered, exactly as `30 ÷ players` today. Defense and final-attack scores follow the same milestones.
- **Messages.** Texts that quote a tuned number format it from the tuning (75 HP, one-minute timer, `n/3`, damage and gift point ranges, the Snowball reward). A small formatter prints whole numbers without decimals so default output is unchanged.
- **Validation ranges.** Odds 0–100 and summing to 100; chances 0–100; durations above zero except cooldowns, which may be zero; `one_in` at least 1; message range with `min_messages` at least 1 and `max_messages` at least the minimum; `0 < attack.boss_hp_after < defense.boss_hp_after < snowballs.boss_hp_after < 100`; damage values above zero; a defense miss costs 1–100% of a player's HP; gift points `1 <= min <= max`; rewards at least zero; early start at least 1.
- **Not tunable on purpose.** Eight gift buttons and eight board tiles (they are bit masks), the five-button final row, the 100% boss total, and numbers outside the two docs.

## Risks and trade-offs

- A balance mistake is caught at startup, not when players hit it, because the loader validates the whole file and an invalid file stops the bot. The cost is that a typo prevents a restart, which is why the message lists every problem.
- Fights already stored in SQLite keep their state but use the current tuning for the rest of the fight, so changing HP milestones while a fight is open can shift it. Restart between events rather than mid-fight.
- The defaults appear in `DefaultTuning()` and in `tuning.yaml`. A test loads the shipped file and compares it with the defaults to keep them from drifting.

## Boundaries

Decisions (rarity bands, thresholds, damage, rewards) stay in `core`. `app` orders effects and hands the tuning to the core. `tuningfile` and `cmd` own the filesystem and environment. Discord and web only format values the core and app return.
