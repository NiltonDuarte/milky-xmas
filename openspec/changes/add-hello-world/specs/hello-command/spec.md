# Command path

## Purpose

Verify the Milky Xmas command path in Discord and the local acceptance chat. The `/hello` greeting that first did this was removed; `/help me-milky` is the simplest command now.

## ADDED Requirements

### Requirement: Local test chat

The acceptance interface SHALL serve a composer and transcript on loopback and SHALL process commands without Discord credentials.

#### Scenario: Start and use the chat locally

- **GIVEN** no Discord credentials are available
- **WHEN** the developer starts the web executable, opens its loopback URL, and submits `/help me-milky`
- **THEN** the transcript displays both the submitted command and the list of commands

#### Scenario: Unknown command

- **GIVEN** the local chat is open
- **WHEN** the user submits a command that does not exist, such as `/hello`
- **THEN** the transcript displays an error with the supported commands and the composer accepts a subsequent attempt

### Requirement: Discord slash command registration

The Discord adapter SHALL register its slash commands in the configured server and SHALL remove commands it no longer offers.

#### Scenario: Removed command

- **GIVEN** the server still has `/hello` registered
- **WHEN** the bot starts
- **THEN** `/hello` is removed

#### Scenario: Missing Discord configuration

- **GIVEN** the Discord token or server ID is missing
- **WHEN** the developer starts the Discord executable
- **THEN** startup exits with an error naming the required configuration
