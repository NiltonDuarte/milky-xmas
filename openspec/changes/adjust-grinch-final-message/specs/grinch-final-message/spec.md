# Grinch final message

## Purpose

Make the end of the Grinch fight clear: who won, and who scored.

## ADDED Requirements

### Requirement: Only Max when the Grinch wins

When a fight ends with the Grinch alive, because he escaped or defeated all registered players, the final message SHALL show Max's image and SHALL NOT show the Grinch's image. A victory by the group SHALL show only the Grinch's image.

#### Scenario: The Grinch escapes or wins

- **GIVEN** a fight ends in escape or in the defeat of all players
- **WHEN** the final message is shown
- **THEN** only Max's image is shown

#### Scenario: The group wins

- **GIVEN** the group defeats the Grinch
- **WHEN** the final message is shown
- **THEN** only the Grinch's image is shown

### Requirement: Damage, not points

The damage leaderboard, the player cards, and the players' text area SHALL show the damage a player dealt, the HP they took from the Grinch. Only the final attack SHALL award score points, and the points SHALL be kept for the player without being shown.

#### Scenario: A final hit

- **GIVEN** the final attack, where a hit removes 1 HP and awards 2 points
- **WHEN** a player lands three hits
- **THEN** the damage leaderboard shows 3.00 damage for those hits
- **AND** the player has 6 points kept

#### Scenario: Other phases

- **GIVEN** a player attacks in the snowball, defense, and attack phases
- **WHEN** the fight ends
- **THEN** they have no points from those phases

#### Scenario: Damage adds up to the Grinch's HP

- **GIVEN** the group takes the Grinch from 100 HP to 0
- **WHEN** the final message is shown
- **THEN** the damage on the leaderboard adds up to 100

### Requirement: Leaderboards list players with points

The damage, gift, and Snowball leaderboards SHALL list only players with points above zero, and SHALL be left out when nobody scored.

#### Scenario: Players without points

- **GIVEN** three players joined and only one earned gift points
- **WHEN** the final message is shown
- **THEN** the gift leaderboard lists that player alone

#### Scenario: Nobody scored on a board

- **GIVEN** no player collected a Snowball
- **WHEN** the final message is shown
- **THEN** there is no Snowball leaderboard

#### Scenario: More than ten scoring players

- **GIVEN** more than ten players have points on a leaderboard
- **WHEN** the final message is shown
- **THEN** it lists the top ten and attaches the full standings of the players with points

#### Scenario: Local chat

- **GIVEN** a finished fight in the local chat where one player scored damage and gifts
- **WHEN** the player views the fight
- **THEN** the leaderboards list only the players with points

### Requirement: Counts without units

The gift and Snowball leaderboards SHALL show each player's number alone. The damage leaderboard SHALL keep `dmg`.

#### Scenario: Gifts and Snowballs

- **GIVEN** a finished fight where a player has 4 gift points and another collected 2 Snowballs
- **WHEN** the final message is shown
- **THEN** the lines end with `— 4` and `— 2`
