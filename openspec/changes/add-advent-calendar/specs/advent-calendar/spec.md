# Advent calendar

## Purpose

Let members open December calendar doors and collect seasonal Advent gifts separately from their daily gifts.

## ADDED Requirements

### Requirement: Advent door availability

Both interfaces SHALL display a calendar with doors 1 through 24 and SHALL allow each member to open doors from the first instant of their day through the end of December 24, using UTC. Each UTC year has its own calendar.

#### Scenario: View available doors
- **GIVEN** it is December 4 UTC and a member opened door 1
- **WHEN** they use `/advent view`
- **THEN** they see door 1 as opened, doors 2–4 as available, and doors 5–24 as locked

#### Scenario: Open a missed door
- **GIVEN** it is December 24 UTC and door 1 is unopened
- **WHEN** the member uses `/advent open 1`
- **THEN** they receive one Advent gift and see their Advent total for the year

#### Scenario: Future, invalid, and closed doors
- **GIVEN** a member requests a door outside 1–24, a future door, or a door outside December 1–24 UTC
- **WHEN** they use `/advent open` with that door
- **THEN** the bot explains why it cannot open the door and no reward is recorded

### Requirement: Separate and durable rewards

Each opening SHALL award one Advent gift at most once per member, server, year, and door. Opened doors and Advent totals SHALL persist through Discord bot restarts. Advent activity SHALL leave the daily gift balance and cooldown unchanged.

#### Scenario: Duplicate or simultaneous opening
- **GIVEN** an eligible member sends two requests to open the same door
- **WHEN** both requests complete
- **THEN** exactly one awards an Advent gift and the other reports that the door is already open

#### Scenario: Restart and new year
- **GIVEN** a member opened door 1 and the Discord bot restarts
- **WHEN** they view the current year's calendar
- **THEN** door 1 remains open, and in the following December door 1 can be opened again

#### Scenario: Storage failure
- **GIVEN** an eligible door opening cannot be saved
- **WHEN** the member tries to open it
- **THEN** the bot reports the failure without confirming a reward

### Requirement: Native Advent commands

Discord SHALL register `view` and `open` beneath `/advent`, with a required integer `day` for `open`. It SHALL deliver results privately.

#### Scenario: Use the Advent commands on Discord
- **GIVEN** the bot is online in the configured server
- **WHEN** a member selects an Advent subcommand and submits it
- **THEN** only that member sees the calendar or opening result
