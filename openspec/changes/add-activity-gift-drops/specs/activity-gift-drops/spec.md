# Activity gift drops

## Purpose

Reward conversation in enabled event channels through shared, collectible gift drops while keeping collection fair for all participating players.

## ADDED Requirements

### Requirement: Random Active Drop

The system SHALL use qualifying human activity in a channel enabled for commands to randomly announce a Random Active Drop. Its message SHALL be `Watch out! Something is falling down!` and show eight gift controls in two rows of four. A player SHALL receive at most one gift from each such drop while it is live for ten minutes.

#### Scenario: Player collects a Random Active Drop

- **GIVEN** a Random Active Drop is visible in an enabled channel
- **WHEN** a player selects one of its gift controls
- **THEN** the player receives one randomly selected gift rarity
- **AND** a later selection by that player from the same drop is rejected

#### Scenario: Ignore ineligible channel activity

- **GIVEN** a channel is not enabled for bot commands
- **WHEN** a human sends a message in that channel
- **THEN** that message does not create or advance an activity gift drop

### Requirement: Activity-counted Gift Rain

The system SHALL select a fresh activity threshold from five through one hundred qualifying chat messages for each channel and announce a Gift Rain after the threshold is reached. Its message SHALL be `🎁 It's raining gifts! 🎁 Hallelujah` and show eight gift controls in two rows of four.

#### Scenario: Gift Rain appears after active chat

- **GIVEN** an enabled channel has a selected threshold of five messages
- **WHEN** its fifth qualifying human chat message is observed
- **THEN** the system announces a Gift Rain in that channel
- **AND** it selects a new threshold for later activity

#### Scenario: Expired controls do not award gifts

- **GIVEN** an activity gift drop was announced more than ten minutes ago
- **WHEN** a player selects one of its gift controls
- **THEN** no gift is awarded

#### Scenario: Each player has a one-minute Gift Rain limit

- **GIVEN** a Gift Rain is visible and a player successfully collects a gift
- **WHEN** the same player selects another Gift Rain control before one minute has elapsed
- **THEN** no additional gift is awarded
- **AND** the response does not reveal the remaining cooldown time

#### Scenario: Gift Rain does not advance without human chat

- **GIVEN** an enabled channel receives no human messages
- **WHEN** time passes
- **THEN** no Gift Rain is announced only because of elapsed time

### Requirement: Shared claim outcome

The system SHALL validate a drop claim and persist the awarded inventory gift through the shared application behavior so that the same eligibility and gift-rarity rules are reusable outside Discord.

#### Scenario: Concurrent selection only awards once

- **GIVEN** a player is eligible to collect one Random Active Drop
- **WHEN** duplicate selections for that player and drop arrive concurrently
- **THEN** exactly one gift is added to their inventory
