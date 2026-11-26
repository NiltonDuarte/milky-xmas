## ADDED Requirements

### Requirement: Stats panel

In Discord `/stats` SHALL show a private panel with a drop-down for server stats or leaderboards, a second drop-down with the categories and then the server stats, named Milky Way Galaxy, and the information below both. It SHALL open on Leaderboards with the server stats until something is picked. The side of the panel SHALL take a colour of its own for the server stats and for each leaderboard. Every command and choice SHALL be acknowledged before any storage work.

#### Scenario: Opening the panel

- **GIVEN** a player
- **WHEN** they run `/stats`
- **THEN** they see both drop-downs and the server stats below them

#### Scenario: Choosing a leaderboard

- **GIVEN** the panel is open
- **WHEN** the player picks Bad eggs and then Milky Way Galaxy in the second drop-down
- **THEN** the same private message shows the Bad eggs list and then the server stats, with both drop-downs

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

### Requirement: Leaderboard places

Each listed player SHALL show a medal for the first three places or the place number after them, the count, and under it a bar of eight squares proportional to the first place's count, with at least one filled square.

#### Scenario: Comparing players

- **GIVEN** the top four players have 105, 40, 22, and 1 bad deed points
- **WHEN** Bad eggs is shown
- **THEN** the first three have medals, the fourth has the number 4, and their bars have 8, 3, 2, and 1 filled squares
