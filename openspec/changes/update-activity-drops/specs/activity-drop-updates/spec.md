# Activity drop updates

## Purpose

Refine the Random Active Drop and Gift Rain presentation, rewards, and ending.

## ADDED Requirements

### Requirement: Drop images

Discord SHALL attach the Grinch and Max image to a Random Active Drop and the gift rain image to a Gift Rain, with the existing announcement text and gift buttons.

#### Scenario: Random Active Drop image

- **GIVEN** a Random Active Drop appears
- **WHEN** the message is posted
- **THEN** it shows `Watch out! Something is falling down!`, the Grinch and Max image, and eight gift buttons

#### Scenario: Gift Rain image

- **GIVEN** a Gift Rain appears
- **WHEN** the message is posted
- **THEN** it shows the Gift Rain text, the gift rain image, and eight gift buttons

### Requirement: Better odds for good players

A player with Good deeds of 50% or more SHALL have a 50% chance of a normal gift, 35% of rare, and 15% of epic from a Random Active Drop. Other players SHALL keep 70%, 25%, and 5%.

#### Scenario: Player at 50% Good deeds

- **GIVEN** a player with Good deeds of 50%
- **WHEN** they collect from a Random Active Drop and the rarity roll is 85 or higher
- **THEN** they receive an epic gift

#### Scenario: Player below 50% Good deeds

- **GIVEN** a player with Good deeds of 49%
- **WHEN** they collect from a Random Active Drop and the rarity roll is 85
- **THEN** they receive a rare gift

### Requirement: Opened gifts are disabled

When a player collects from a Random Active Drop, Discord SHALL disable the button they clicked. A player who tries to open a button someone else already opened SHALL receive a private reply and no gift. Each player still collects once per drop.

#### Scenario: Button disabled

- **GIVEN** a Random Active Drop with eight enabled buttons
- **WHEN** a player collects with the fifth button
- **THEN** that button becomes disabled and the others stay enabled

#### Scenario: Slot already taken

- **GIVEN** the fifth button was opened
- **WHEN** another player clicks it at the same moment
- **THEN** they are told someone already opened that gift and receive nothing

### Requirement: Gift Rain length and Coal

A Gift Rain SHALL end after ten collections in total. Each collection SHALL have a 25% chance of giving one Coal and otherwise one gift of a daily rarity. The one-minute limit per player SHALL stay hidden from players.

#### Scenario: Tenth collection

- **GIVEN** nine collections from a Gift Rain
- **WHEN** a player makes the tenth
- **THEN** they receive their prize and the rain ends, and the next click says the drop is no longer available

#### Scenario: Coal prize

- **GIVEN** a Gift Rain collection rolls Coal
- **WHEN** the player collects
- **THEN** their Coal increases by one and they see a Coal message

### Requirement: Final message with collectors

When a drop ends, because its ten minutes passed or a Gift Rain reached ten collections, its message SHALL say `⌛ This drop has expired.`, remove its buttons, and list each collection as the player and the prize in order, or say that nobody collected a gift. The list SHALL NOT ping players. The local chat ends a drop with `/drop expire <id>`.

#### Scenario: Expired with collectors

- **GIVEN** two players collected from a drop
- **WHEN** the drop expires
- **THEN** the message lists both players with their prizes and has no buttons

#### Scenario: Expired without collectors

- **GIVEN** nobody collected
- **WHEN** the drop expires
- **THEN** the message says nobody collected a gift

#### Scenario: Ended once

- **GIVEN** a Gift Rain ended by its tenth collection
- **WHEN** its ten minutes later pass
- **THEN** the message is not edited again
