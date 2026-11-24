# Advent calendar buttons

## Purpose

Let players open Advent doors from a picture with numbered buttons.

## ADDED Requirements

### Requirement: Calendar with numbered buttons

Discord SHALL answer `/advent view` privately with the Advent calendar image, the player's Advent gift total, and 24 buttons numbered 1 to 24. A door that can be opened today SHALL have an enabled button. An opened door SHALL show ❌ and a disabled button. A door that is locked or expired SHALL have a disabled button.

#### Scenario: View on December 4

- **GIVEN** it is December 4 UTC and the player opened door 2
- **WHEN** the player runs `/advent view`
- **THEN** the image appears above the text, buttons 1, 3, and 4 are enabled, button 2 shows ❌ disabled, and buttons 5 to 24 are disabled

### Requirement: Open a door by clicking

Clicking an enabled button SHALL open that door once for the player, award the Advent gift as `/advent open` does, show the result, and turn that button into a disabled ❌ in the same message. The player SHALL NOT be able to open that door again.

#### Scenario: Open door 3

- **GIVEN** door 3 is available
- **WHEN** the player clicks button 3
- **THEN** the message says the player opened door 3 with the new total, and button 3 shows ❌ and is disabled

#### Scenario: Stale button

- **GIVEN** a player clicks a button for a door that is already open or still locked
- **WHEN** the click is handled
- **THEN** no gift is awarded, the message gives the usual answer, and the buttons show the current state

#### Scenario: Calendar closed

- **GIVEN** it is after December 24 UTC
- **WHEN** the player runs `/advent view`
- **THEN** every button is disabled

### Requirement: Other ways to open keep working

`/advent open <day>` in Discord and the Advent commands in the local chat SHALL keep their behavior and the one-opening-per-day limit.

#### Scenario: Open by command

- **GIVEN** door 1 is available
- **WHEN** the player runs `/advent open 1`
- **THEN** the door opens once and its button shows ❌ the next time the player views the calendar
