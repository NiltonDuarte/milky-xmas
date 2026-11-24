# YAML tuning config

## Purpose

Let the team check and adjust the documented balance numbers in one YAML file.

## ADDED Requirements

### Requirement: Built-in defaults match the documented numbers

Without a tuning file, the bot SHALL behave exactly as documented in `docs/activity-drops-numbers.md` and `docs/grinch-fight-numbers.md`. The shipped `tuning.yaml` SHALL contain every tunable value at that default.

#### Scenario: No file configured

- **GIVEN** `MILKY_TUNING_CONFIG` is not set
- **WHEN** the bot starts and a player collects a Gift Rain gift
- **THEN** the result follows the documented 25% Coal, 52.5% normal, 18.75% rare, 3.75% epic odds

#### Scenario: Shipped file equals the defaults

- **GIVEN** the repository's `tuning.yaml`
- **WHEN** it is loaded
- **THEN** the resulting tuning equals the built-in defaults

### Requirement: Load the tuning file at startup

Both executables SHALL read the file named by `MILKY_TUNING_CONFIG` once at startup. Keys left out of the file SHALL keep their default value.

#### Scenario: Partial file

- **GIVEN** a file that only sets `drops.gift_rain.claims_to_end: 3`
- **WHEN** the bot starts
- **THEN** a Gift Rain ends on its third collection and every other number keeps its default

#### Scenario: File cannot be read

- **GIVEN** `MILKY_TUNING_CONFIG` names a missing file
- **WHEN** the bot starts
- **THEN** it refuses to start and the message names the file

### Requirement: Reject invalid tuning

The loader SHALL refuse a file with an unknown key, a value of the wrong type or unit (including a decimal where a whole number is required), or a value outside its allowed range, and SHALL report every problem it finds. A range problem names the key path. An unknown key or a wrong type names the key and its line.

#### Scenario: Odds that do not add up

- **GIVEN** `gifts.daily_rarity` sets normal 70, rare 25, epic 10
- **WHEN** the file is loaded
- **THEN** loading fails and the message names `gifts.daily_rarity` and that the odds must add up to 100

#### Scenario: Misspelled key

- **GIVEN** a file with `drops.gift_rain.claims_to_edn: 3`
- **WHEN** the file is loaded
- **THEN** loading fails and the message names the unknown key

#### Scenario: Boss HP milestones out of order

- **GIVEN** `grinch.defense.boss_hp_after` is 80 while `grinch.snowballs.boss_hp_after` is 75
- **WHEN** the file is loaded
- **THEN** loading fails and the message says the milestones must fall from 100 toward 0

#### Scenario: Several problems at once

- **GIVEN** a file with two invalid values
- **WHEN** the file is loaded
- **THEN** one error lists both key paths

### Requirement: Tuning drives the activity drops

The activity drop rules SHALL read these numbers from the tuning: drop lifetime, the Random Active Drop chance, the Good deeds threshold and boosted rarity odds, the daily rarity odds, and for a Gift Rain the message range, the collections that end it, the per-player cooldown, and the Coal chance.

#### Scenario: Shorter rain

- **GIVEN** `claims_to_end` is 3
- **WHEN** three players each collect from a Gift Rain
- **THEN** the third collection ends the drop and the final list shows three collections

#### Scenario: Different Random Active Drop chance

- **GIVEN** `drops.random_active.one_in` is 1
- **WHEN** an eligible player message is not a Gift Rain trigger
- **THEN** a Random Active Drop appears

#### Scenario: Different drop lifetime

- **GIVEN** `drops.lifetime` is `2m`
- **WHEN** two minutes pass without the drop ending
- **THEN** the drop expires and further clicks are refused

### Requirement: Tuning drives the Grinch fight

The Grinch rules SHALL read these numbers from the tuning: the spawn chance, activity window, check interval, and cooldown after a fight; the registration duration and early start count; the fight duration; per phase the boss HP milestone and damage; the HP a defense miss costs; the final-attack damage, score, gift points, and Snowball reward; and the loss rewards. Messages that quote these numbers SHALL show the configured values.

#### Scenario: Different snowball milestone

- **GIVEN** `grinch.snowballs.boss_hp_after` is 90 and `grinch.snowballs.damage` is 1
- **WHEN** players land ten accepted snowball hits
- **THEN** the boss is at 90 HP, the defense phase starts, and the countdown begins

#### Scenario: Defense scales with the milestones

- **GIVEN** the defense phase goes from 75 to 45 HP with four players
- **WHEN** the players land four safe defenses
- **THEN** each took 7.5 HP and the attack phase starts at 45 HP

#### Scenario: Early start count

- **GIVEN** `grinch.registration.early_start_players` is 2
- **WHEN** a second player joins during registration
- **THEN** combat starts immediately and the announcement showed `n/2 to start early`

#### Scenario: Loss rewards

- **GIVEN** `loss_rewards.coal` is 2 and `yellow_snowballs_per_group_member` is 2 and three players registered
- **WHEN** the Grinch escapes
- **THEN** each registered player receives 2 Coal and 6 Yellow Snowballs

#### Scenario: Message quotes the configured number

- **GIVEN** `final_attack.hit_score` is 3
- **WHEN** the final attack opens
- **THEN** the message offers `3 damage points` and an accepted ⚔️ awards 3 points

### Requirement: Both interfaces use the same tuning

The Discord bot and the local chat SHALL load the tuning the same way and apply it through the same shared use cases.

#### Scenario: Local chat check

- **GIVEN** the local chat started with a file where `claims_to_end` is 3
- **WHEN** `/message` triggers a Gift Rain and three collections are made
- **THEN** the drop ends after the third

#### Scenario: Discord check

- **GIVEN** the Discord bot started with the same file
- **WHEN** a Gift Rain appears in a configured channel
- **THEN** it ends after the third collection (Discord test server check pending)
