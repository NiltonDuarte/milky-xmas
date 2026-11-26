## ADDED Requirements

### Requirement: Watering is a good deed

Each watering that counts SHALL move the player's aura 1 point toward good deeds by default, without passing +50. A refused watering SHALL NOT change the aura. Both adapters apply this.

#### Scenario: Watering that counts

- **GIVEN** a player with aura -20
- **WHEN** they water the tree
- **THEN** their aura is -19 and the reply says `😇 Good deeds +1`

#### Scenario: Refused watering

- **GIVEN** a player who watered within the last hour
- **WHEN** they run `/water` again
- **THEN** their aura does not change

#### Scenario: Top of the scale

- **GIVEN** a player with aura +50
- **WHEN** they water the tree
- **THEN** the watering counts, the aura stays +50, and there is no good-deed line

#### Scenario: Aura cannot be saved

- **GIVEN** the aura cannot be saved
- **WHEN** a player waters the tree
- **THEN** the watering still counts, the reply has no good-deed line, and the failure is logged

### Requirement: Water bar

The tree SHALL show its level as ten squares, one blue square for each full 10% and white squares for the rest, followed by the percentage.

#### Scenario: Partly watered

- **GIVEN** a tree at 34%
- **WHEN** it is shown
- **THEN** the bar has three blue squares, seven white squares, and `34%`

### Requirement: Tree box and picture

A Discord tree reply with a picture SHALL be one box holding the picture, the watering line, the tree's level and state, without a list of waterers. Waterings are still counted for each player. The picture SHALL be the withered tree from 0% to 50% and the decorated tree at 100%. From 51% to 99% the box SHALL have no picture.

#### Scenario: Passing the halfway mark

- **GIVEN** a tree at 49%
- **WHEN** a player waters it
- **THEN** the channel sees one box with the text and no picture

#### Scenario: Filling the tree

- **GIVEN** a tree at 98%
- **WHEN** a player waters it
- **THEN** the channel sees one box with the decorated tree

#### Scenario: Channel post fails

- **GIVEN** Discord refuses the channel message
- **WHEN** a player waters the tree
- **THEN** the player sees the box privately

#### Scenario: Watering a full tree

- **GIVEN** a tree at 100%
- **WHEN** a player runs `/water`
- **THEN** no water or good deed is added and only that player sees the decorated tree with `🎄 The tree is fully watered. Try again later!`

### Requirement: Water answers in time

In Discord `/water` SHALL be acknowledged privately before any storage work.

#### Scenario: Refusal after acknowledgement

- **GIVEN** a watering is refused
- **WHEN** the command runs
- **THEN** the interaction is acknowledged first and then edited into the private reason
