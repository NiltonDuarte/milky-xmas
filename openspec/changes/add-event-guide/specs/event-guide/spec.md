## ADDED Requirements

### Requirement: Event guide

`/guide` SHALL explain what a player can do in the event, one topic at a time: daily gifts and stocking, sending and opening gifts, gift drops, aura, snowballs, the Grinch fight, the tree, the Advent calendar, and statistics. It SHALL not explain moderator commands. Both adapters apply this.

#### Scenario: Opening the guide

- **GIVEN** a member in an enabled channel
- **WHEN** they run `/guide`
- **THEN** only they see the starting topic and a way to pick another

#### Scenario: Picking a topic

- **GIVEN** a member with the guide open in Discord
- **WHEN** they pick the Grinch fight in the menu
- **THEN** the same private reply shows how the fight works

#### Scenario: Unknown topic in the local chat

- **GIVEN** the local chat
- **WHEN** `/guide nothing` is sent
- **THEN** the reply lists the topics that exist

### Requirement: Guide numbers

The guide SHALL show the numbers the bot is using, and SHALL not describe a rule that tuning turned off.

#### Scenario: Tuned watering

- **GIVEN** tuning that makes a watering add 7%
- **WHEN** the tree topic is read
- **THEN** it says a watering adds 7%
