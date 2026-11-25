# Grinch join button

## Purpose

Let a player join the Grinch fight with one click on the fight message.

## ADDED Requirements

### Requirement: Join from the fight message

The shared Grinch message SHALL offer a Join button while players can still register. Pressing it SHALL register the player and send one private confirmation. Pressing it again, or after registration closed, SHALL show the player nothing and SHALL NOT register them twice.

#### Scenario: First click

- **GIVEN** an open fight that a player has not joined
- **WHEN** the player presses Join
- **THEN** the player is registered
- **AND** only that player sees one private confirmation

#### Scenario: Click after joining

- **GIVEN** a player who already joined the fight
- **WHEN** the player presses Join again
- **THEN** nothing is shown to the player and the fight message does not change

#### Scenario: Other players still join

- **GIVEN** one player has joined
- **WHEN** a second eligible player presses Join
- **THEN** the second player is registered and sees their own confirmation

#### Scenario: Store failure

- **GIVEN** the store fails when the player presses Join
- **WHEN** the click is handled
- **THEN** the click was already acknowledged
- **AND** the player sees a private message to try again

#### Scenario: Local chat

- **GIVEN** an open fight in the local chat, where Discord's button does not exist
- **WHEN** a player sends `/grinch join` twice
- **THEN** the first reply confirms the join and the second says the player already joined
