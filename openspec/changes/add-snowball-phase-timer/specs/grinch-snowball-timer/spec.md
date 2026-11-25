# Grinch snowball timer

## Purpose

Give the first phase of the Grinch fight a time limit.

## ADDED Requirements

### Requirement: Two minutes to reach 75 HP

The group SHALL have two minutes from the start of combat to bring the Grinch down to the snowball milestone. If the time runs out first, the Grinch SHALL win.

#### Scenario: Time runs out

- **GIVEN** a fight in the snowball phase with the Grinch above 75 HP
- **WHEN** two minutes pass since combat started
- **THEN** the fight ends as an escape and every registered player gets the loss rewards
- **AND** a snowball click after that is refused

#### Scenario: Milestone reached in time

- **GIVEN** a fight in the snowball phase
- **WHEN** the group brings the Grinch to 75 HP before the two minutes end
- **THEN** the defense phase opens and the one-minute combat timer starts at that click

#### Scenario: Longer timer

- **GIVEN** `grinch.snowballs.duration` is five minutes in the tuning file
- **WHEN** four minutes pass in the snowball phase
- **THEN** the fight is still running

#### Scenario: Message shows the limit

- **GIVEN** a fight in the snowball phase
- **WHEN** the message is shown
- **THEN** it shows when the Grinch escapes and says he wins if time runs out

#### Scenario: Local chat

- **GIVEN** a snowball phase in the local chat
- **WHEN** two minutes pass without reaching 75 HP
- **THEN** the fight view shows the encounter as escaped
