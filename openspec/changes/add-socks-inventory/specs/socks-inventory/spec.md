# Socks inventory

## Purpose

Let players check the quantities of the gifts and event items in their own inventory.

## ADDED Requirements

### Requirement: View personal inventory

Both adapters SHALL support `/socks` to list the player's normal, rare, and epic gift quantities, Snowballs, and Coal in the current server. Items with no balance SHALL appear with quantity zero. The command SHALL only read inventory data.

#### Scenario: Show all item quantities

- **GIVEN** a player owns gifts and event items
- **WHEN** the player runs `/socks`
- **THEN** the reply lists the current quantity of all five items

#### Scenario: Show empty or upgraded inventory

- **GIVEN** a player has no inventory or their saved inventory predates Snowballs and Coal
- **WHEN** the player runs `/socks`
- **THEN** the reply shows zero for missing items and retains any existing gift balances

#### Scenario: Reflect a transfer

- **GIVEN** one player gave another player a gift
- **WHEN** both use `/socks`
- **THEN** each sees their updated quantity for that rarity

#### Scenario: Inventory is unavailable

- **GIVEN** the player's inventory cannot be read
- **WHEN** the player runs `/socks`
- **THEN** the reply reports that the inventory could not be loaded

### Requirement: Private Discord inventory

Discord SHALL register `/socks` in the configured server with no arguments and display the inventory only to the invoking player.

#### Scenario: View a private inventory

- **GIVEN** the bot is online and the player can use application commands
- **WHEN** the player selects `/socks`
- **THEN** Discord shows that player's current inventory in a private reply
