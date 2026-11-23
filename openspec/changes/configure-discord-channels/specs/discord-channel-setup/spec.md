# Discord channel setup

## Purpose

Let each Discord server choose where the bot and the Grinch can operate.

## ADDED Requirements

### Requirement: Moderator onboarding and setup

On first joining a server, the bot SHALL ask moderators to run `/setup channel` in a text channel where the bot can post. Members with Manage Server or Administrator SHALL be able to set command and boss permissions for each text channel and inspect the choices with `/setup view`. Setup SHALL be reachable before any channels are configured and settings SHALL persist across restarts.

#### Scenario: First join and setup

- **GIVEN** the bot joins a server and can post in a text channel
- **WHEN** a moderator follows the onboarding prompt and selects channels
- **THEN** the bot saves the choices and `/setup view` shows the configured roles

#### Scenario: Unauthorized member

- **GIVEN** a member lacks permission to manage the server
- **WHEN** they attempt to use setup
- **THEN** the bot rejects the request and the choices stay the same

### Requirement: Enforce Discord channel choices

The bot SHALL run gameplay slash commands only in channels enabled for commands. Manual Grinch summons SHALL also require a boss-enabled channel. Automatic spawns and encounter buttons SHALL operate only in boss-enabled channels. The two permissions SHALL be independently configurable per channel.

#### Scenario: Channels with different permissions

- **GIVEN** a general channel enables commands only and an event channel enables the boss only
- **WHEN** players use gameplay commands and the Grinch attempts to spawn
- **THEN** commands work in the general channel, automatic spawns can occur in the event channel, and manual summons require both permissions

#### Scenario: Disable a channel

- **GIVEN** a moderator disables one or both channel permissions
- **WHEN** players attempt the corresponding actions
- **THEN** the bot enforces the new permissions and continues to use the other configured channels

### Requirement: Announce Christmas setup

The bot SHALL post the requested Christmas introduction in a boss channel when it is first configured. Successful posts SHALL be recorded across restarts and repeated setup SHALL reuse that record.

#### Scenario: Configure a boss channel

- **GIVEN** the bot can post in the target channel
- **WHEN** a moderator enables the boss there for the first time
- **THEN** the bot posts the five-line Christmas introduction in that channel

#### Scenario: Announcement fails

- **GIVEN** the boss setting has been saved and Discord rejects the introduction
- **WHEN** the moderator repeats the setup after restoring permission
- **THEN** the bot retries the outstanding introduction
