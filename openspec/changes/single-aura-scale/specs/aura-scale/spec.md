# Aura scale

## Purpose

Show where each player stands between bad deeds and good deeds on one scale.

## ADDED Requirements

### Requirement: One aura from -50 to +50

Each player SHALL have one aura between -50 and +50, starting at 0. Good actions SHALL raise it and bad actions SHALL lower it, without leaving the scale.

#### Scenario: Send gifts

- **GIVEN** a neutral player with two rare gifts
- **WHEN** they send both
- **THEN** their aura is +20 and the recipient's aura does not change

#### Scenario: Send Coal or a yellow snowball

- **GIVEN** a player with an aura of +5
- **WHEN** they send three Coal
- **THEN** their aura is -10

#### Scenario: Limit of the scale

- **GIVEN** a player with an aura of +45
- **WHEN** they send an epic gift
- **THEN** the gift is sent and the aura is +50

#### Scenario: Upgrade

- **GIVEN** a player with Bad and Good percentages from before
- **WHEN** the upgraded bot opens the database
- **THEN** their aura is 0 and their items are unchanged

### Requirement: Aura check shows the scale

`/aura check` SHALL show the scale from -50 to +50 with a marker at the player's position and the number under it.

#### Scenario: Bad side

- **GIVEN** a player with an aura of -15
- **WHEN** they run `/aura check`
- **THEN** the marker is on the bad side and the number reads -15

### Requirement: Drop bonus from +25

A player with an aura at or above the configured threshold, +25 by default, SHALL use the boosted rarity odds on a Random Active Drop.

#### Scenario: Below the threshold

- **GIVEN** a player with an aura of +24 or any negative aura
- **WHEN** they collect from a Random Active Drop
- **THEN** the normal odds apply

### Requirement: Send gift messages

`/send gift` SHALL name the quantity, item, and recipient without percentages.

#### Scenario: Gifts

- **GIVEN** a sender with two rare gifts
- **WHEN** they send both to Bob
- **THEN** they see `Sent 2 🎁 rare gifts to Bob. Your 😇 good deeds are going up.`

#### Scenario: Coal

- **GIVEN** a sender with two Coal
- **WHEN** they send both to Bob
- **THEN** they see `Sent 2 🌑 Coal to Bob. Your 😈 bad deeds are increasing. Beware!`

These apply to both interfaces. The aura image above the scale is Discord only.
