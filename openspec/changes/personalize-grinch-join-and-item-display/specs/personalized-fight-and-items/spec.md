# Personalized fight and items

## Purpose

Give each player their own Grinch registration button and display item balances and effects consistently.

## ADDED Requirements

### Requirement: Personal Discord registration

The shared Grinch message SHALL offer a registration opener. Only a player eligible to join SHALL receive a private panel with Join. After enrollment, the panel SHALL remove Join and opening registration again SHALL offer no Join control. Other players SHALL remain able to register.

#### Scenario: Join privately

- **GIVEN** two eligible players open registration
- **WHEN** the first player clicks Join
- **THEN** Join disappears from their private panel with no new success message
- **AND** the second player can still join

### Requirement: Yellow throw aura

A committed Yellow Snowball throw SHALL increase the sender's Bad deeds by five percentage points, capped at 100%. A failed throw or a regular Snowball throw SHALL leave aura unchanged.

#### Scenario: Throw a Yellow Snowball

- **GIVEN** a player has one Yellow Snowball and 20% Bad deeds
- **WHEN** they throw it at another player
- **THEN** the item moves to the recipient and the sender has 25% Bad deeds

### Requirement: Compact inventory and item icons

Both adapters SHALL show `/socks` on four lines, with a title, gift quantities, both Snowball balances, and Coal. Item messages SHALL use 🌑 for Coal and 🟡 for Yellow Snowball.

#### Scenario: View an inventory with one Coal

- **GIVEN** a player owns one Coal and zero other items
- **WHEN** they use `/socks`
- **THEN** the reply shows columns with the gifts by rarity and both snowball kinds at 0, and Coal at 1, with no heading
