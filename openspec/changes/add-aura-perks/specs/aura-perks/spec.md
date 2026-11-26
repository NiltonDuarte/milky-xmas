## ADDED Requirements

### Requirement: Nice and Naughty lists

A player whose aura is at or above `aura.nice_at` (25) SHALL be on the Nice list, and at or below `aura.naughty_at` (-25) on the Naughty list. Anyone else gets no perk. Both adapters apply this.

#### Scenario: Limits

- **GIVEN** the default numbers
- **WHEN** players have auras of 25, 24, and -25
- **THEN** they are on the Nice list, no list, and the Naughty list

### Requirement: Daily claim and watering perks

The Nice list SHALL get one more gift of the claimed rarity from the daily claim and 1% more water from a watering. A Naughty list watering SHALL add 1% less water, and at least 1%. The Naughty list SHALL get one Yellow Snowball with the daily claim, and any player at aura zero or below one Coal. The reply SHALL name the bonus.

#### Scenario: Nice claim and watering

- **GIVEN** a player on the Nice list
- **WHEN** they claim a roll of one gift and water the tree
- **THEN** they collect two gifts and the tree gains 3%

#### Scenario: Naughty claim

- **GIVEN** a player on the Naughty list
- **WHEN** they claim
- **THEN** they also receive one Yellow Snowball and one Coal

#### Scenario: Neutral claim

- **GIVEN** a player at aura zero
- **WHEN** they claim
- **THEN** they also receive one Coal

#### Scenario: Naughty watering

- **GIVEN** a player on the Naughty list
- **WHEN** they water the tree with the default numbers
- **THEN** the tree gains 1% and the message says so

### Requirement: Drop perks

The boosted Random Active Drop odds SHALL apply to the Nice list. A Naughty list player SHALL have a 25% chance of getting a Gift Rain gift twice, counted as one collection. Coal is never doubled. A Naughty list player SHALL wait 30 seconds between Gift Rain collections instead of one minute.

#### Scenario: Doubled rain gift

- **GIVEN** a Naughty list player and a lucky roll
- **WHEN** they collect a gift from a Gift Rain
- **THEN** they get two of that rarity, are told so, and the summary shows ×2

#### Scenario: Shorter rain wait

- **GIVEN** a Naughty list player and a neutral player who collected 30 seconds ago
- **WHEN** both try again
- **THEN** only the Naughty list player collects

### Requirement: Grinch fight perks

A wrong defense SHALL cost 20% of maximum HP by default, replacing the 10% of earlier changes. For a Nice list player it SHALL cost 34% and every attack SHALL deal 20% more damage. For a Naughty list player it SHALL cost 10%, and they SHALL get no Coal when the Grinch wins or escapes, keeping the Yellow Snowballs. Perks are those the player had on entering the fight.

#### Scenario: Nice attack

- **GIVEN** a Nice list player in the snowball phase
- **WHEN** they hit with the default damage of 4.17
- **THEN** the Grinch loses 5 HP

#### Scenario: Naughty loss

- **GIVEN** a Naughty list player in a lost fight
- **WHEN** the rewards are given
- **THEN** they get Yellow Snowballs and no Coal

#### Scenario: Aura cannot be read

- **GIVEN** the player's aura cannot be read
- **WHEN** they press a fight button
- **THEN** the action is not counted and they are told it failed

### Requirement: One counted send per player a day

A gift, Coal, or Yellow Snowball SHALL move the sender's aura only if no send to that recipient moved it in the same direction during the last 24 hours (`aura.send_cooldown`). The items move either way, and the gift or Coal reply SHALL say when the aura did not change.

#### Scenario: Same gift twice

- **GIVEN** a player sent a rare gift to Bob today
- **WHEN** they send Bob another one
- **THEN** Bob gets it, the sender's aura stays, and the reply says why

#### Scenario: Other direction or player

- **GIVEN** a player sent a gift to Bob today
- **WHEN** they send Bob Coal, or send Carol a gift
- **THEN** the aura moves

### Requirement: Daily aura limit

The aura SHALL rise at most 15 points and fall at most 15 points per UTC day (`aura.daily_limit`). Items still move, and the gift or Coal reply SHALL say when the limit kept the aura where it was.

#### Scenario: Epic gift

- **GIVEN** a neutral player
- **WHEN** they send an epic gift
- **THEN** their aura is 15, and another good deed that day does not raise it

#### Scenario: Next day

- **GIVEN** the limit stopped a send yesterday
- **WHEN** they send to the same player after 00:00 UTC
- **THEN** the aura moves

### Requirement: Watering from zero up

A watering SHALL move the aura toward good only for a player at zero or above. The water still counts.

#### Scenario: Naughty player waters

- **GIVEN** a player below zero
- **WHEN** they water the tree
- **THEN** the tree gains water and their aura stays
