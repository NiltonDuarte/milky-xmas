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

### Requirement: Moderator summon

The bot SHALL register `/grinch summon` in the configured Discord server. A member with effective Manage Messages permission in the encounter channel SHALL be able to summon the Grinch from that channel, independent of activity, spawn odds, and the automatic spawn cooldown. Manual and automatic appearances SHALL share the same single-encounter limit.

#### Scenario: Moderator summons the Grinch
- **GIVEN** a moderator can manage messages in the configured encounter channel and no fight is active
- **WHEN** the moderator uses `/grinch summon` in that channel
- **THEN** the bot announces a normal encounter with 100 HP and the snowball button

#### Scenario: Summon is unavailable
- **GIVEN** the caller lacks permission, uses another channel, or no encounter channel is configured
- **WHEN** the caller attempts `/grinch summon`
- **THEN** the request creates no encounter and any bot response is private

#### Scenario: An encounter already exists or is being announced
- **GIVEN** an active encounter exists or another summon or automatic appearance has reserved the channel
- **WHEN** a moderator uses `/grinch summon`
- **THEN** the bot replies privately that an encounter is already in progress and creates no duplicate announcement

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

When the Grinch reaches 75% HP, the encounter SHALL replace the snowball button with two rows of four randomized emoji buttons. The choices SHALL use Christmas trees (🎄) and snowflakes (❄️) and include at least one safe choice. A Christmas tree SHALL leave the player's HP unchanged and deal percentage-based damage to the boss based on the number of players in the encounter. Boss HP SHALL stay at or above 45%. A snowflake SHALL remove 10% of that player's maximum HP.

#### Scenario: Defense board appears at 75% HP
- **GIVEN** the Grinch has 76 of 100 HP with time remaining in the first phase
- **WHEN** a snowball hits
- **THEN** the attack button is replaced by eight defense choices in two rows of four
- **AND** stale snowball clicks cause no additional damage

#### Scenario: Choose a safe tile
- **GIVEN** a player has 100 HP and the defense board is active
- **WHEN** they choose a Christmas tree
- **THEN** their HP stays at 100 and the boss loses HP subject to the 45% floor

#### Scenario: Choose a snowflake
- **GIVEN** a player has 100 HP and the defense board is active
- **WHEN** they choose a snowflake
- **THEN** their HP drops to 90 and the HP of the boss and other players stays the same

### Requirement: Defense damage scaling

On entry to phase two, the encounter SHALL count the distinct players who landed a snowball in phase one, including the hit that triggered the transition. Each successful defense SHALL deal 30 percentage points of the boss's maximum HP divided by that fixed count, subject to the 45% floor. Damage SHALL accumulate without rounding each hit.

#### Scenario: Three players enter defense
- **GIVEN** three distinct players landed snowballs during phase one
- **WHEN** a player selects a Christmas tree during defense
- **THEN** the Grinch loses 10 percentage points of maximum HP unless the health floor caps the damage

#### Scenario: Six players enter defense
- **GIVEN** six distinct players landed snowballs during phase one
- **WHEN** a player selects a Christmas tree during defense
- **THEN** the Grinch loses 5 percentage points of maximum HP unless the health floor caps the damage

#### Scenario: Fractional damage reaches the threshold
- **GIVEN** seven distinct players landed snowballs during phase one
- **WHEN** the group records seven successful defenses
- **THEN** their combined damage brings the Grinch from 75% to exactly 45% HP

#### Scenario: Successful defenses reach the boss health floor
- **GIVEN** the boss is at 46 of 100 HP during the defense phase
- **WHEN** successful defenses together would otherwise deal more than 1 HP of damage
- **THEN** their combined damage is capped at 1 HP and the encounter enters phase three once at 45 HP

### Requirement: Phase three attacks

The encounter SHALL enter phase three at 45% HP. The announcement SHALL show the Grinch shouting and two rows of four randomized attack buttons using 🧊 and 🫯. An accepted 🫯 attack SHALL deal base damage. An accepted 🧊 attack SHALL deal 1.05 times base damage. Phase-three grid attacks SHALL keep boss HP at or above 5%.

#### Scenario: Attack board replaces defense
- **GIVEN** the boss is above 45% HP in the defense phase
- **WHEN** a successful defense brings him to 45%
- **THEN** the announcement shows the Grinch's shout and the randomized attack grid
- **AND** the defense buttons can no longer apply damage

#### Scenario: Ice attack deals bonus damage
- **GIVEN** the boss has enough HP above 5% to take full damage
- **WHEN** players make one eligible 🫯 attack and one eligible 🧊 attack
- **THEN** the two attacks deal base damage and 105% of base damage respectively

#### Scenario: Phase three reaches the health floor
- **GIVEN** phase three is active and the boss has 6 of 100 HP
- **WHEN** eligible phase-three actions would deal more than 1 HP of damage
- **THEN** their combined damage is capped at 1 HP and the boss enters the final attack at 5 HP

### Requirement: Final attack and encounter outcome

At 5% HP, the encounter SHALL replace the attack grid with one repeatable emoji attack button. Final attacks SHALL reduce boss HP to zero, which ends the fight in victory. The Grinch SHALL escape if undefeated one minute after appearing. The deadline SHALL apply across all phases and the final attack.

#### Scenario: Finish the Grinch
- **GIVEN** the Grinch has reached 5% HP with time remaining
- **WHEN** players use the single emoji attack until the Grinch reaches zero HP
- **THEN** the announcement shows a victory and disables the final attack button

#### Scenario: Overall time limit takes precedence
- **GIVEN** the Grinch still has HP one minute after appearing
- **WHEN** the bot checks the fight or receives an attack at or after the deadline
- **THEN** the announcement shows that the Grinch escaped and no further damage is applied

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
