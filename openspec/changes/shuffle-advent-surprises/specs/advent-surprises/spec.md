# Advent surprises

## Purpose

Give each player their own calendar of surprises to open and revisit.

## ADDED Requirements

### Requirement: Each player has their own calendar

`/advent view` SHALL show the 24 numbered doors in an order of the player's own, and the surprises behind them SHALL be dealt in an order of the player's own. Both orders SHALL stay the same for that player through the year.

#### Scenario: Two players

- **GIVEN** two players in the same server
- **WHEN** each runs `/advent view` and opens door 2
- **THEN** their buttons are in different orders and their surprises are dealt differently

#### Scenario: Viewing again

- **GIVEN** a player who has viewed the calendar
- **WHEN** they run `/advent view` again
- **THEN** the buttons are in the same order

### Requirement: Clicking a door shows its surprise

A click on an available or opened door SHALL send the player its surprise privately. A door SHALL be available from its December date through December 24 UTC. An opened door SHALL be green and stay clickable, and a door that is not available SHALL be disabled.

#### Scenario: Open today's door

- **GIVEN** it is December 4 and door 3 is closed
- **WHEN** the player clicks 3
- **THEN** they receive its surprise privately and the button turns green

#### Scenario: Read an opened door again

- **GIVEN** door 1 was opened
- **WHEN** the player clicks it on a later day
- **THEN** they receive the same surprise

#### Scenario: A later day

- **GIVEN** it is December 4
- **WHEN** the player views the calendar
- **THEN** door 5 and later are disabled

#### Scenario: Storage failure

- **GIVEN** the opening cannot be saved
- **WHEN** the player clicks a door
- **THEN** they are told privately and the calendar stays as it was

These apply to Discord. The local chat lists the doors in the player's order and uses `/advent open <day>` in place of a click.
