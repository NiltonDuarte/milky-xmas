# Reply formats

## Purpose

Present daily claim, inventory, and aura results in the event's chosen format.

## ADDED Requirements

### Requirement: Daily claim presentation

A collected daily claim SHALL show the heading `## 🎅🏻 Christmas won’t be Christmas without any presents.` and a code block with `You collected <amount> gifts 🎁.` and `Gift rarity · <rarity> ·`. The early-claim cooldown reply SHALL stay unchanged.

#### Scenario: Collected daily claim

- **GIVEN** a player can claim
- **WHEN** they run `/daily claim`
- **THEN** the reply uses the heading and code block with the amount and rarity

### Requirement: Inventory presentation

`/look in-your-stocking` SHALL show the heading `## 🧦 Your inventory` and a code block listing gifts by rarity, Snowballs, Yellow Snowballs, and Coal.

#### Scenario: Inventory listing

- **GIVEN** a player has any inventory
- **WHEN** they run `/look in-your-stocking`
- **THEN** the reply uses the heading and code block with every balance, including zeros

### Requirement: Aura presentation

`/aura check` SHALL show `😈 Bad <n>%  |  <n>% Good 😇` in a code block. Discord SHALL show the bad good image above that line in the private reply.

#### Scenario: Aura line

- **GIVEN** a player has any aura
- **WHEN** they run `/aura check`
- **THEN** the reply shows Bad on the left and Good on the right in a code block

#### Scenario: Image above the aura line

- **GIVEN** a player runs `/aura check` in Discord
- **WHEN** the reply is shown
- **THEN** the image appears above the aura line, and the local chat shows only the aura line

### Requirement: Stocking command name

The inventory command SHALL be `/look in-your-stocking` and `/socks` SHALL no longer be offered. Discord SHALL register `/look` with one subcommand `in-your-stocking` and no arguments, remove the old `/socks` command from the configured server, and reply privately.

#### Scenario: Inventory under the new name

- **GIVEN** a player has any inventory
- **WHEN** they run `/look in-your-stocking`
- **THEN** they see their inventory as before

#### Scenario: Old command removed

- **GIVEN** the bot starts in a server that still has `/socks` registered
- **WHEN** the bot registers its commands
- **THEN** `/socks` is removed from that server
