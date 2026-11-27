## ADDED Requirements

### Requirement: Moderator guide

`/mod-guide` SHALL explain, one topic at a time, every command with who may use it, how the automatic features behave, what is saved, and what a restart loses. Both adapters apply this; the local chat player is treated as a moderator.

#### Scenario: Moderator opens the guide

- **GIVEN** a member with Manage Messages, Manage Server, or Administrator
- **WHEN** they run `/mod-guide`
- **THEN** only they see the overview and a way to pick another topic

#### Scenario: Player

- **GIVEN** a member without those permissions
- **WHEN** the command reaches the bot
- **THEN** they are privately told they lack permission

### Requirement: Moderator guide numbers

The moderator guide SHALL show the numbers the bot is using.

#### Scenario: Tuned spawn chance

- **GIVEN** tuning that makes the Grinch's spawn chance 33%
- **WHEN** the topic about how a fight runs is read
- **THEN** it says 33%
