## ADDED Requirements

### Requirement: Stats panel

In Discord `/stats` SHALL show a private panel with a drop-down for server stats or leaderboards, and for leaderboards a second drop-down with the categories, both above the information. A choice that reads records SHALL be acknowledged before any storage work; opening the leaderboards SHALL be answered in one request.

#### Scenario: Choosing a leaderboard

- **GIVEN** the panel is open
- **WHEN** the player picks Leaderboards and then Bad eggs
- **THEN** the same private message shows the Bad eggs list with both drop-downs

#### Scenario: Records cannot be read

- **GIVEN** the database cannot be read
- **WHEN** the player picks a view
- **THEN** the panel says the statistics could not be read and keeps its drop-downs

### Requirement: Server stats

Server stats SHALL show the tree level, waterings, gifts and Coal sent, normal and yellow snowballs thrown, drops collected, and Grinch fights won and lost, without good or bad deed totals. Both adapters apply this.

#### Scenario: Receiving does not count

- **GIVEN** a player only received gifts
- **WHEN** the statistics are shown
- **THEN** that player has sent nothing

### Requirement: Leaderboards

Each leaderboard SHALL list up to ten players with a count above zero, highest first. Throwers SHALL list normal and yellow snowballs apart, and Gifters gifts and Coal apart. Bad eggs and Good eggs SHALL list each player with the total of their deeds, adding up the points without one side lowering the other and without a limit. Plant moms and dads SHALL count waterings. Both adapters apply this.

#### Scenario: Deeds do not cancel

- **GIVEN** a player sent two rare gifts, twelve Coal, and one yellow snowball
- **WHEN** the leaderboards are shown
- **THEN** they have 20 good deed points and 65 bad deed points

#### Scenario: Empty category

- **GIVEN** nobody threw a yellow snowball
- **WHEN** Throwers is shown
- **THEN** the yellow list says nobody yet
