# Help command

## Purpose

Let players discover the commands they can use.

## ADDED Requirements

### Requirement: List available commands

Both adapters SHALL support `/help me-milky` to list the commands available to the player. Everyone SHALL see the player commands. Discord SHALL also list the moderator commands `/give gift` and `/grinch summon` to members with Manage Messages or Administrator, and `/setup` to members with Manage Server or Administrator. The Discord reply SHALL be private. The local chat player is treated as a moderator.

#### Scenario: Regular member

- **GIVEN** a member without moderator permissions
- **WHEN** they run `/help me-milky`
- **THEN** they see the player commands and no moderator commands

#### Scenario: Moderator

- **GIVEN** a member with Manage Messages
- **WHEN** they run `/help me-milky`
- **THEN** they also see `/give gift` and `/grinch summon` but not `/setup`

#### Scenario: Server manager or administrator

- **GIVEN** a member with Manage Server, or Administrator
- **WHEN** they run `/help me-milky`
- **THEN** a manager sees `/setup`, and an administrator sees every group

#### Scenario: Old names

- **GIVEN** the bot has started
- **WHEN** a member looks for `/help me-milky` or `/hello`
- **THEN** neither command is offered

#### Scenario: Private Discord reply

- **GIVEN** a member runs `/help me-milky` in an enabled channel
- **WHEN** the command completes
- **THEN** only that member sees the list
