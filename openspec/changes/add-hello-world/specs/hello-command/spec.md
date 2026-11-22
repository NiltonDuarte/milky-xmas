# Hello command

## Purpose

Provide a greeting that verifies the Milky Xmas command path in Discord and the local acceptance chat.

## ADDED Requirements

### Requirement: Shared greeting

Both interfaces SHALL reply with `Hello, world!` when a user submits `/hello`.

#### Scenario: Greeting in either interface

- **GIVEN** the local chat is running or the bot is online in the configured Discord server
- **WHEN** a user submits `/hello`
- **THEN** the interface displays `Hello, world!` as the bot's reply

### Requirement: Local test chat

The acceptance interface SHALL serve a composer and transcript on loopback and SHALL process commands without Discord credentials.

#### Scenario: Start and use the chat locally

- **GIVEN** no Discord credentials are available
- **WHEN** the developer starts the web executable, opens its loopback URL, and submits `/hello`
- **THEN** the transcript displays both the submitted command and the greeting

#### Scenario: Unknown command

- **GIVEN** the local chat is open
- **WHEN** the user submits text other than `/hello`
- **THEN** the transcript displays an error with the supported command and the composer accepts a subsequent attempt

### Requirement: Discord slash command registration

The Discord adapter SHALL register `/hello` in the configured server and respond to the command with a single visible interaction acknowledgement.

#### Scenario: Invoke the registered slash command

- **GIVEN** the bot is installed in the configured server and the user can use application commands in the channel
- **WHEN** the bot starts and the user selects and submits `/hello`
- **THEN** Discord displays one public bot reply to the interaction

#### Scenario: Missing Discord configuration

- **GIVEN** the Discord token or server ID is missing
- **WHEN** the developer starts the Discord executable
- **THEN** startup exits with an error naming the required configuration
