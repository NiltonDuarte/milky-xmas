# Grinch activity areas

## Purpose

Show the Grinch's voice and the players' actions while the fight runs.

## ADDED Requirements

### Requirement: The Grinch's area

While the fight runs, the message SHALL show the latest lines of the Grinch. Each line SHALL be picked at random from a pool for its moment. A snowball that hits a player SHALL show the player and the HP they lost, and a gift or Snowball a player takes back SHALL show the player and the amount.

#### Scenario: A player is hit

- **GIVEN** the defense phase
- **WHEN** a player chooses a tile the Grinch hit with a snowball
- **THEN** the Grinch's area shows a line for the hit, the player, and the 10% HP lost

#### Scenario: Gifts taken back

- **GIVEN** the final attack
- **WHEN** a player picks a gift tile that gives 3 gift points
- **THEN** the Grinch's area shows a line and that the player took back 3 gift points

#### Scenario: A new phase

- **GIVEN** a fight that reaches the defense phase
- **WHEN** the message updates
- **THEN** the Grinch's area shows a line for that phase

#### Scenario: Latest events

- **GIVEN** more events have happened than the areas show
- **WHEN** the message updates
- **THEN** the Grinch's area shows his latest three events and the players' area the latest five, newest first

### Requirement: The players' area

While the fight runs, the message SHALL show who attacked last and the damage they did. A player who hides behind the tree in the defense phase SHALL say a line picked at random from a pool.

#### Scenario: A player hides behind the tree

- **GIVEN** the defense phase
- **WHEN** a player chooses a safe tile
- **THEN** the players' area shows the player, a line, and the damage they did

#### Scenario: A snowball hit

- **GIVEN** the snowball phase
- **WHEN** a player attacks
- **THEN** the players' area shows the player and 1.00 damage

### Requirement: Fixed size

The two areas SHALL keep the same height whether they have no events or more than they can show, and SHALL sit in a box of their own apart from the panel with the image and the buttons.

#### Scenario: Events come in

- **GIVEN** a running fight with no events, and the same fight with many
- **WHEN** the message is shown in each case
- **THEN** both areas have the same number of lines
- **AND** the areas are in a separate box from the panel with the image and buttons

### Requirement: Buttons below the text areas

During combat, the buttons of the current phase SHALL be in the box of the two text areas, below them and above the players' cards, all in the same box. The panel with the image SHALL have no buttons during combat.

#### Scenario: Layout of the combat screen

- **GIVEN** a running fight in any phase
- **WHEN** the message is shown
- **THEN** the panel with the image comes first
- **AND** one box follows with the Grinch's text, the players' text, the buttons, and then the players' cards

### Requirement: Areas belong to the running fight

The areas SHALL show only while the fight runs, with a placeholder before the first event, and SHALL be shown in the local chat as text.

#### Scenario: Before any event

- **GIVEN** a fight that has just started
- **WHEN** the message is shown
- **THEN** the Grinch's area says he is watching and the players' area says nobody has attacked yet

#### Scenario: Lobby and final message

- **GIVEN** the registration or the final message
- **WHEN** the message is shown
- **THEN** it has no Grinch or players area

#### Scenario: Local chat

- **GIVEN** a running fight in the local chat
- **WHEN** the player views the fight
- **THEN** it lists the Grinch's lines and the players' attacks as text
