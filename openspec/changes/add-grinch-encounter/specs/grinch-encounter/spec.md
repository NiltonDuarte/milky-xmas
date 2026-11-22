# Grinch encounter

## Purpose

Create spontaneous cooperative Grinch fights in an active Christmas event channel.

## ADDED Requirements

### Requirement: Activity-based appearances

With a Grinch channel configured, the bot SHALL check once per minute and roll a 10% spawn chance when a human message was posted in that channel within the past 10 minutes, there is no active fight, and at least 30 minutes have passed since the last fight ended. It SHALL announce appearances only in that channel.

#### Scenario: Eligible roll
- **GIVEN** the channel has a recent human message and no active fight or spawn cooldown
- **WHEN** the next check rolls a spawn
- **THEN** the bot posts one fight announcement in that channel with 100 HP and a snowball attack button

#### Scenario: Inactive channel or missed roll
- **GIVEN** the configured channel lacks a recent human message, or it is active and the check misses its roll
- **WHEN** the bot checks for a spawn
- **THEN** no fight is announced
- **AND** bot messages and messages in other channels do not qualify as activity

#### Scenario: Fight already active or cooling down
- **GIVEN** the channel has an active fight or the last fight ended less than 30 minutes ago
- **WHEN** the bot checks for a spawn
- **THEN** no new fight is announced

### Requirement: Snowball phase

The first phase SHALL start when the Grinch appears and use one 10-second attack window for all members. The announcement SHALL show an Attack button with a snowball icon. Each distinct accepted attack during the phase SHALL remove 1 HP from the shared total. Members SHALL be able to attack repeatedly within the window. Snowball damage SHALL keep shared health at or above 75% of maximum HP, regardless of how many attacks arrive together.

#### Scenario: Members throw snowballs together
- **GIVEN** the Grinch just appeared with 100 HP
- **WHEN** one member presses the button twice and another member presses it once within 10 seconds
- **THEN** the announcement shows 97 HP

#### Scenario: Late joiner shares the deadline
- **GIVEN** nine seconds have elapsed since the Grinch appeared
- **WHEN** a member attacks for the first time
- **THEN** they deal 1 damage and share the original deadline

#### Scenario: The snowball window ends
- **GIVEN** the Grinch is still above 75% HP after ten seconds
- **WHEN** the deadline arrives or a member tries to attack at or after it
- **THEN** the announcement shows the remaining HP and the phase as ended
- **AND** the snowball button is disabled and no further snowball damage is applied

#### Scenario: Simultaneous attacks at the health floor
- **GIVEN** the Grinch has 76 of 100 HP during the snowball phase
- **WHEN** multiple members attack simultaneously
- **THEN** exactly one more point of damage is applied, the Grinch reaches 75 HP, and the encounter enters the defense phase once

### Requirement: Defense phase

When the Grinch reaches 75% HP, the encounter SHALL replace the snowball button with two rows of four randomized emoji buttons. The choices SHALL use Christmas trees (🎄) and snowflakes (❄️) and include at least one safe choice. A Christmas tree SHALL leave the player's HP unchanged. A snowflake SHALL remove 10% of that player's maximum HP.

#### Scenario: Defense board appears at 75% HP
- **GIVEN** the Grinch has 76 of 100 HP with time remaining in the first phase
- **WHEN** a snowball hits
- **THEN** the attack button is replaced by eight defense choices in two rows of four
- **AND** stale snowball clicks cause no additional damage

#### Scenario: Choose a safe tile
- **GIVEN** a player has 100 HP and the defense board is active
- **WHEN** they choose a Christmas tree
- **THEN** their HP stays at 100

#### Scenario: Choose a snowflake
- **GIVEN** a player has 100 HP and the defense board is active
- **WHEN** they choose a snowflake
- **THEN** their HP drops to 90 and the HP of the boss and other players stays the same

### Requirement: Encounter outcome

At zero HP the group SHALL win. An undefeated Grinch SHALL escape ten minutes after appearing.

#### Scenario: The Grinch escapes
- **GIVEN** the fight still has HP ten minutes after appearing
- **WHEN** the bot checks the fight or a member attempts another attack
- **THEN** the announcement shows that the Grinch escaped and further attacks cannot lower HP

### Requirement: Recover an encounter safely

The bot SHALL preserve active fight state and its message identity across restarts and SHALL avoid duplicate announcements if the outcome of a send is uncertain.

#### Scenario: Restart during a fight
- **GIVEN** a fight was announced and some HP was removed
- **WHEN** the bot restarts before its deadline
- **THEN** the original announcement continues the same fight with the recorded HP

#### Scenario: Announcement could not be confirmed
- **GIVEN** the bot cannot determine whether Discord received a spawn announcement
- **WHEN** the spawn attempt ends
- **THEN** the bot does not immediately post a duplicate and recovers for a later eligible check

### Requirement: Local combat acceptance

The local chat SHALL let a developer exercise attack and outcome behavior against a fixture encounter without contacting Discord.

#### Scenario: Test combat locally
- **GIVEN** a fixture encounter is active in the local chat
- **WHEN** the developer attacks
- **THEN** the local interface displays the result from the shared combat rules
