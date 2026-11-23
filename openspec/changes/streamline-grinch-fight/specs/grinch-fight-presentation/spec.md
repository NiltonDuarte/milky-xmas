# Grinch fight presentation

## Purpose

Keep the Grinch fight focused on the shared board and its current health and controls.

## ADDED Requirements

### Requirement: One enrollment per player

A registered player SHALL be unable to join the same encounter again. On Discord, repeat Join clicks SHALL have no effect and be acknowledged silently. The shared Join control SHALL remain available to players who have yet to join.

#### Scenario: A registered player clicks again

- **GIVEN** one player has joined the current fight
- **WHEN** they click Join again while a different player has yet to join
- **THEN** their registration count stays the same, they receive no message, and the other player can still join

### Requirement: Silent fight controls

Discord SHALL acknowledge successful and stale board interactions without sending a message to the player. A reshuffle SHALL edit the existing fight message. A storage error MAY produce a private error message. Both adapters SHALL present the updated state after an accepted action.

#### Scenario: Pick an action and refresh the board

- **GIVEN** a player clicks a current defense or final board button
- **WHEN** the action is accepted and the board reshuffles
- **THEN** the existing board updates and the action sends no new confirmation message

### Requirement: Display Grinch health

The fight display SHALL show the boss health with two decimals using `👽 Grinch · <value>/100% HP`.

#### Scenario: The boss reaches zero health

- **GIVEN** the players defeat the boss
- **WHEN** the final state is displayed
- **THEN** the health line reads `👽 Grinch · 0.00/100% HP`
