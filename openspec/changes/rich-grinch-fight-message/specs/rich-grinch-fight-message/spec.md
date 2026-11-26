# Rich Grinch fight message

## Purpose

Show the whole Grinch fight in one message that updates as the fight changes.

## ADDED Requirements

### Requirement: Registration shows who is ready

While registration is open, the fight message SHALL show the Grinch image, the time left, and the list of players who pressed Start, with the Start button.

#### Scenario: Players press Start

- **GIVEN** an open registration with one player who pressed Start
- **WHEN** a second player presses Start
- **THEN** the message lists both players as ready

### Requirement: Cards during the fight

While the fight is running, the message SHALL show the boss HP and a card for each of the eight players who dealt the most damage, with the player's current HP. A summary line SHALL cover the other players. The damage each player deals SHALL appear only in the final standings, not while the fight runs.

#### Scenario: Damage and HP change

- **GIVEN** a running fight with two players
- **WHEN** one player deals damage and the other loses HP in the defense phase
- **THEN** the boss HP and the second player's HP change on the message, and no player damage is written

#### Scenario: More than eight players

- **GIVEN** a running fight with ten players
- **WHEN** the message updates
- **THEN** it shows eight cards and a summary of the other two players
- **AND** the message stays within Discord's component limit

### Requirement: Final standings

When the fight ends, the message SHALL show the damage, gift, and Snowball leaderboards of the players who took part. Snowballs count what each player collected with the final ❄️ button.

#### Scenario: Final leaderboards

- **GIVEN** a fight that ends after players picked attack, gift, and Snowball buttons
- **WHEN** the final message is shown
- **THEN** it lists the damage, gift, and Snowball leaderboards

#### Scenario: Many players

- **GIVEN** a fight that ended with more than ten players
- **WHEN** the final message is shown
- **THEN** each leaderboard shows the top ten and the full standings are attached

### Requirement: Local chat standings

The local chat SHALL list the Snowball leaderboard in the final reply, after the gift leaderboard.

#### Scenario: Snowball leaderboard in the local chat

- **GIVEN** a fight in the local chat that ends after the player picked the ❄️ button
- **WHEN** the final reply is shown
- **THEN** it lists the Snowball leaderboard after the gift leaderboard
