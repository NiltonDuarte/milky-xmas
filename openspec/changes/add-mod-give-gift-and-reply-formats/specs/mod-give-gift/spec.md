# Moderator gift grants

## Purpose

Let moderators add gifts to a player's inventory.

## ADDED Requirements

### Requirement: Grant gifts to a player

Both adapters SHALL support `/give gift` with a recipient, an amount, and a rarity of normal, rare, or epic. The grant SHALL add that many gifts to the recipient in the current server and SHALL NOT change any inventory, aura, or cooldown of the moderator, or the recipient's aura or cooldown. Only moderators SHALL be able to use it.

#### Scenario: Moderator grants gifts

- **GIVEN** a moderator and a player in the same server
- **WHEN** the moderator runs `/give gift` for the player with amount 3 and rarity rare
- **THEN** the player's rare gifts increase by 3 and the moderator sees a confirmation

#### Scenario: Reject invalid grants

- **GIVEN** a moderator runs `/give gift`
- **WHEN** the recipient is a bot or the moderator, the amount is not a positive whole number, the rarity is unknown, or the recipient's inventory is full
- **THEN** no balance changes and the reply explains the problem

#### Scenario: Non-moderator is refused

- **GIVEN** a Discord member without Manage Messages or Administrator
- **WHEN** they invoke `/give gift`
- **THEN** no gifts are added and the reply says the permission is required

#### Scenario: Private Discord reply

- **GIVEN** a moderator uses `/give gift` in Discord
- **WHEN** the command completes
- **THEN** only the moderator sees the result
