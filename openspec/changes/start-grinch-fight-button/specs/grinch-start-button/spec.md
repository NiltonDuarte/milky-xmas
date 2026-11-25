# Grinch start button

## Purpose

Let the players decide when the Grinch fight begins.

## ADDED Requirements

### Requirement: Start begins a countdown

The Grinch announcement SHALL offer a Start button. The first press SHALL begin a countdown of the registration duration, and the fight SHALL start when it ends. Every player who presses Start SHALL be in the fight.

#### Scenario: First press

- **GIVEN** an announced fight that nobody has started
- **WHEN** a player presses Start
- **THEN** the player is in the fight and sees one private confirmation
- **AND** the announcement shows when the fight starts, one minute after the press

#### Scenario: Countdown from the press

- **GIVEN** a fight announced 30 seconds ago that nobody has started
- **WHEN** a player presses Start
- **THEN** the fight starts one minute after that press, not one minute after the announcement

#### Scenario: Three players

- **GIVEN** two players pressed Start
- **WHEN** a third player presses Start
- **THEN** the fight starts at once

### Requirement: Nobody starts the fight

If nobody presses Start within the start window of the announcement, the encounter SHALL be cancelled with no rewards.

#### Scenario: Nobody presses Start

- **GIVEN** an announced fight
- **WHEN** the start window passes without a press
- **THEN** the encounter is cancelled

### Requirement: Anyone can take part

There SHALL be no joining. The combat screen SHALL NOT offer a Join button, and any player SHALL be able to take part once the fight has started: their first accepted action SHALL put them in the fight. A refused action SHALL NOT.

#### Scenario: Combat screen

- **GIVEN** a running fight
- **WHEN** the message is shown
- **THEN** it has the buttons of the phase and no Join button

#### Scenario: A new player attacks

- **GIVEN** a running fight and a player who did not press Start
- **WHEN** the player presses an attack button
- **THEN** the attack counts and the player is in the fight, with a card and a place on the leaderboards
- **AND** the player receives the loss rewards if the Grinch wins

#### Scenario: A refused action

- **GIVEN** a running fight
- **WHEN** a player who is not in it presses a button of a board that has changed
- **THEN** the click is refused and the player is not in the fight

### Requirement: Older announcements

Buttons of announcements posted before this change SHALL do nothing.

#### Scenario: An older button

- **GIVEN** a button from an announcement with a registration, lobby, or join ID
- **WHEN** a player presses it
- **THEN** nothing is sent and nothing changes

### Requirement: Local chat

The local chat SHALL start the countdown with the first `/grinch join` and repeat the same rules.

#### Scenario: Local chat start

- **GIVEN** a fight in the local chat that nobody started
- **WHEN** the player sends `/grinch join`
- **THEN** the reply confirms the join and the fight starts when the countdown ends
