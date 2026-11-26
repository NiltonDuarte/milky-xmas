# Christmas tree

## Purpose

Let the community keep one shared Christmas tree alive by watering it.

## ADDED Requirements

### Requirement: Shared tree water level

Each server SHALL have one Christmas tree. A new tree SHALL start at 10% water. The water level SHALL fall by 1 point for every full hour since it was last updated, down to 0%, and SHALL rise by 2 points per watering, up to 100%. The tree is full and alive only at 100%.

#### Scenario: Water decays

- **GIVEN** a tree at 12% that was last updated 3 hours ago
- **WHEN** a player looks at or waters the tree
- **THEN** the level before watering is 9%

#### Scenario: Watering raises the level

- **GIVEN** a tree at 98%
- **WHEN** a player waters it
- **THEN** the level becomes 100%

#### Scenario: Full tree

- **GIVEN** a tree at 100%
- **WHEN** a player waters it
- **THEN** nothing changes and the player is told the tree is full

### Requirement: Water command

Both adapters SHALL support `/water` with no arguments. A player SHALL be able to water once every hour. A refusal SHALL say how long remains, in hours and minutes.

#### Scenario: First watering

- **GIVEN** a new tree
- **WHEN** a player runs `/water`
- **THEN** the level becomes 12% and the player is counted among the waterers

#### Scenario: Watering too soon

- **GIVEN** a player watered the tree 4 minutes ago
- **WHEN** they run `/water` again
- **THEN** no water is added and they are told to try again in 56 minutes

#### Scenario: Different players

- **GIVEN** a player just watered the tree
- **WHEN** another player runs `/water`
- **THEN** the second watering counts

#### Scenario: Concurrent waterings

- **GIVEN** several players water at the same moment
- **WHEN** the waterings are saved
- **THEN** every watering is counted

#### Scenario: Different servers

- **GIVEN** two servers
- **WHEN** a player waters one tree
- **THEN** the other server's tree does not change

### Requirement: Event end

The event SHALL end on January 5, 2027 at 01:00 UTC. From then on the tree SHALL keep the level it had at that moment, and `/water` SHALL NOT add water. The reply SHALL tell the player the event is over and show the tree's final state privately.

#### Scenario: Watering after the end

- **GIVEN** the time is after January 5, 2027 01:00 UTC
- **WHEN** a player runs `/water`
- **THEN** no water is added and the player sees that the event is over with the tree's final level

#### Scenario: No decay after the end

- **GIVEN** a tree watered shortly before the end
- **WHEN** the tree is read a week after the end
- **THEN** its level is what it was at 01:00 UTC on January 5, 2027

### Requirement: Tree presentation

A Discord watering SHALL be shown to the whole channel without pinging anyone, as described by the `tree-watering-reward` capability. A refusal SHALL be private. The local chat SHALL show the same text without the image.

#### Scenario: Public watering

- **GIVEN** a player waters the tree in Discord
- **WHEN** the command completes
- **THEN** the channel sees the tree picture, `💧 <player> watered the tree! +2%`, and the level

#### Scenario: Private refusal

- **GIVEN** a watering is refused
- **WHEN** the command completes
- **THEN** only the player sees the reason

### Requirement: Tree survives restarts

The tree's level, last update, and waterings SHALL persist in the database.

#### Scenario: Restart

- **GIVEN** a watered tree and a restarted bot
- **WHEN** a player runs `/water`
- **THEN** the level and cooldowns continue from the stored values
