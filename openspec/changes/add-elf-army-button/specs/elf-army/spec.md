# Elf army

## Purpose

Let players take the `elfs` role with one click on the setup message.

## ADDED Requirements

### Requirement: Join the elf army button

The Christmas introduction SHALL carry a button labeled "Join the elf army 🛷". A click SHALL give the player the `elfs` role.

#### Scenario: First click

- **GIVEN** a Grinch channel was just set up and the introduction shows the button
- **WHEN** a player without the `elfs` role clicks it
- **THEN** the player gets the `elfs` role
- **AND** only that player sees a reply saying they joined

#### Scenario: Already an elf

- **GIVEN** a player who already has the `elfs` role
- **WHEN** they click the button
- **THEN** nothing changes and the private reply says they are already in the elf army

#### Scenario: Role missing

- **GIVEN** the server has no `elfs` role
- **WHEN** a moderator sets up a Grinch channel, or a player clicks the button
- **THEN** the bot creates a mentionable `elfs` role and uses it

#### Scenario: Discord refuses the role

- **GIVEN** the bot lacks Manage Roles or its role is below `elfs`
- **WHEN** a player clicks the button
- **THEN** the private reply says the player could not be added and what a moderator should check

### Requirement: The elfs role gates nothing

The `elfs` role SHALL exist only so the community can tag its members. No command SHALL require it, and no moderator command SHALL depend on it.

#### Scenario: Player without the role

- **GIVEN** a player who never joined the elf army, in a channel where commands are enabled
- **WHEN** they use any command that is not for moderators
- **THEN** it works exactly as it does for an elf

#### Scenario: Moderator without the role

- **GIVEN** a moderator who is not an elf
- **WHEN** they use a moderator command (`/give`, `/grinch summon`, `/setup`)
- **THEN** it works, because those commands depend on Discord's Manage Messages and Manage Server permissions only

This applies to Discord only. The local chat has no roles.
