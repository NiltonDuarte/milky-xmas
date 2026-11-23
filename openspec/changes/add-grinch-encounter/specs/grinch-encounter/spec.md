# Grinch encounter

## Purpose

Create spontaneous cooperative Grinch fights in an active Christmas event channel.

## ADDED Requirements

### Requirement: Activity-based appearances

With a Grinch channel configured, the bot SHALL check once per minute and roll a 10% spawn chance when a human message was posted in that channel within the past 10 minutes, there is no active fight, and at least 30 minutes have passed since the last fight ended. It SHALL announce appearances only in that channel.

#### Scenario: Eligible roll
- **GIVEN** the channel has a recent human message and no active fight or spawn cooldown
- **WHEN** the next check rolls a spawn
- **THEN** the bot posts one fight announcement in that channel with a 🎅 join button

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
- **THEN** the bot announces the encounter with the 🎅 join button

#### Scenario: Summon is unavailable
- **GIVEN** the caller lacks permission, uses another channel, or no encounter channel is configured
- **WHEN** the caller attempts `/grinch summon`
- **THEN** the request creates no encounter and any bot response is private

#### Scenario: An encounter already exists or is being announced
- **GIVEN** an active encounter exists or another summon or automatic appearance has reserved the channel
- **WHEN** a moderator uses `/grinch summon`
- **THEN** the bot replies privately that an encounter is already in progress and creates no duplicate announcement

### Requirement: Encounter registration

The 🎅 button SHALL remain available until the encounter ends. Each member can join once. The same message SHALL start untimed combat with 100 HP as soon as three members join, or after one minute if one or two joined. With no members at the deadline, the encounter SHALL be cancelled. Only members who have joined can fight.

#### Scenario: A member joins before combat
- **GIVEN** registration is open
- **WHEN** a member presses 🎅 twice before the deadline
- **THEN** the first click privately confirms that they joined, the second privately confirms they already joined, and they are counted once
- **AND** they can attack when the snowball button appears

#### Scenario: Third member starts the fight early
- **GIVEN** two members joined within the first minute
- **WHEN** a third distinct member presses 🎅 before the deadline
- **THEN** combat starts immediately, with no defense countdown until the Grinch reaches 75 HP

#### Scenario: A member joins during combat
- **GIVEN** combat is underway and a member has not joined
- **WHEN** the member presses 🎅
- **THEN** they receive private confirmation and can participate in the current phase
- **AND** the boss's health and deadline remain unchanged by joining

#### Scenario: Nobody joins
- **GIVEN** registration opened and nobody joined
- **WHEN** one minute elapses
- **THEN** the announcement shows a cancellation without starting combat

### Requirement: Snowball phase

The first phase SHALL start when the Grinch appears and continue without a time limit until he reaches 75% HP. The announcement SHALL show an Attack button with a snowball icon. Each distinct accepted attack during the phase SHALL remove 1 HP from the shared total. Members SHALL be able to attack repeatedly while this phase is active. Snowball damage SHALL keep shared health at or above 75% of maximum HP, regardless of how many attacks arrive together.

#### Scenario: Members throw snowballs together
- **GIVEN** the Grinch just appeared with 100 HP
- **WHEN** one member presses the button twice and another member presses it once during phase one
- **THEN** the announcement shows 97 HP

#### Scenario: A single player can finish phase one without a timer
- **GIVEN** one registered member is fighting and the Grinch is above 75% HP after more than a minute
- **WHEN** the member keeps throwing snowballs
- **THEN** the Grinch stays in phase one until he reaches exactly 75% HP after a total of 25 snowballs
- **AND** the defense phase starts a shared one-minute countdown

#### Scenario: Simultaneous attacks at the health floor
- **GIVEN** the Grinch has 76 of 100 HP during the snowball phase
- **WHEN** multiple members attack simultaneously
- **THEN** exactly one more point of damage is applied, the Grinch reaches 75 HP, and the encounter enters the defense phase once

### Requirement: Defense phase

When the Grinch reaches 75% HP, the encounter SHALL replace the snowball button with two rows of four randomized emoji buttons. The choices SHALL use Christmas trees (🎄) and snowflakes (❄️) and include at least one safe choice. A Christmas tree SHALL leave the player's HP unchanged and deal percentage-based damage to the boss based on the number of players in the encounter. Boss HP SHALL stay at or above 45%. A snowflake SHALL remove 10% of that player's maximum HP.

#### Scenario: Defense board appears at 75% HP
- **GIVEN** the Grinch has 76 of 100 HP during the first phase
- **WHEN** a snowball hits
- **THEN** the attack button is replaced by eight defense choices in two rows of four
- **AND** stale snowball clicks cause no additional damage

#### Scenario: Reroll after a defense choice
- **GIVEN** the defense board is active
- **WHEN** a player makes an accepted tree or snowflake choice and defense remains active
- **THEN** the board changes to a new randomized arrangement with both emoji types

#### Scenario: A click arrives from the previous board
- **GIVEN** an accepted choice changed the board
- **WHEN** a second click arrives for the old arrangement
- **THEN** it changes neither boss HP nor player HP

#### Scenario: Player runs out of HP
- **GIVEN** a player has 10 HP left during defense
- **WHEN** they choose a snowflake
- **THEN** their HP reaches zero and further actions by that player have no effect for that fight
- **AND** if another registered member has HP left, the fight continues

#### Scenario: Choose a safe tile
- **GIVEN** a player has 100 HP and the defense board is active
- **WHEN** they choose a Christmas tree
- **THEN** their HP stays at 100 and the boss loses HP subject to the 45% floor

#### Scenario: Choose a snowflake
- **GIVEN** a player has 100 HP and the defense board is active
- **WHEN** they choose a snowflake
- **THEN** their HP drops to 90 and the HP of the boss and other players stays the same

### Requirement: Defense damage scaling

On entry to phase two, the encounter SHALL use the number of distinct players who joined with 🎅 before the defense phase, including those who never attacked. Each successful defense SHALL deal 30 percentage points of the boss's maximum HP divided by that fixed count, subject to the 45% floor. Damage SHALL accumulate without rounding each hit.

#### Scenario: Three players enter defense
- **GIVEN** three distinct players joined with 🎅 and only one landed snowballs
- **WHEN** a player selects a Christmas tree during defense
- **THEN** the Grinch loses 10 percentage points of maximum HP unless the health floor caps the damage

#### Scenario: Six players enter defense
- **GIVEN** six distinct players joined with 🎅
- **WHEN** a player selects a Christmas tree during defense
- **THEN** the Grinch loses 5 percentage points of maximum HP unless the health floor caps the damage

#### Scenario: Fractional damage reaches the threshold
- **GIVEN** seven distinct players joined with 🎅
- **WHEN** the group records seven successful defenses
- **THEN** their combined damage brings the Grinch from 75% to exactly 45% HP

#### Scenario: Successful defenses reach the boss health floor
- **GIVEN** the boss is at 46 of 100 HP during the defense phase
- **WHEN** successful defenses together would otherwise deal more than 1 HP of damage
- **THEN** their combined damage is capped at 1 HP and the encounter enters phase three once at 45 HP

### Requirement: Phase three attacks

The encounter SHALL enter phase three at 45% HP. The announcement SHALL show the Grinch shouting and two rows of four randomized attack buttons using 🧊 and ☃️. An accepted ☃️ attack SHALL deal base damage. An accepted 🧊 attack SHALL deal 1.05 times base damage. Phase-three grid attacks SHALL keep boss HP at or above 5%.

#### Scenario: Attack board replaces defense
- **GIVEN** the boss is above 45% HP in the defense phase
- **WHEN** a successful defense brings him to 45%
- **THEN** the announcement shows the Grinch's shout and the randomized attack grid
- **AND** the defense buttons can no longer apply damage

#### Scenario: Ice attack deals bonus damage
- **GIVEN** the boss has enough HP above 5% to take full damage
- **WHEN** players make one eligible ☃️ attack and one eligible 🧊 attack
- **THEN** the attacks deal 5 HP and 5.25 HP respectively

#### Scenario: Phase three reaches the health floor
- **GIVEN** phase three is active and the boss has 6 of 100 HP
- **WHEN** eligible phase-three actions would deal more than 1 HP of damage
- **THEN** their combined damage is capped at 1 HP and the boss enters the final attack at 5 HP

### Requirement: Final attack and encounter outcome

At 5% HP, the encounter SHALL replace the attack grid with a random row of four buttons containing ⚔️ and 🎁. Each accepted choice SHALL reroll the row. 🎁 awards the player 1–5 random gift points. ⚔️ removes 1 HP and awards 2 damage points. Zero boss HP ends the fight in victory.

#### Scenario: Finish the Grinch
- **GIVEN** the Grinch has reached 5% HP with time remaining
- **WHEN** players make five accepted ⚔️ choices, dealing 1 HP each
- **THEN** the announcement shows a victory and removes the final row

#### Scenario: Choose a gift in the final row
- **GIVEN** the current row contains a gift button
- **WHEN** a registered player presses that button
- **THEN** the player receives 1–5 gift points and the boss HP stays the same
- **AND** the row is rerolled and a stale click on the previous row awards nothing

#### Scenario: Choose an attack in the final row
- **GIVEN** the Grinch has 5 HP
- **WHEN** a registered player presses ⚔️ on the current row
- **THEN** the Grinch has 4 HP and the player receives 2 damage points

### Requirement: Defeat and escape

The fight SHALL end in defeat when all registered players reach zero HP. The Grinch SHALL escape if undefeated one minute after reaching 75% HP. The deadline SHALL persist through all remaining phases.

#### Scenario: Solo player is defeated
- **GIVEN** a single registered player is fighting the Grinch and has 10 HP left
- **WHEN** the player chooses a snowflake
- **THEN** the fight immediately ends in defeat and the buttons are removed

#### Scenario: A single player can win
- **GIVEN** only one member registered
- **WHEN** they complete all combat phases and reduce the Grinch to zero before the deadline
- **THEN** the announcement shows a victory

#### Scenario: Overall time limit takes precedence
- **GIVEN** the Grinch still has HP one minute after reaching 75% HP
- **WHEN** the bot checks the fight or receives an attack at or after the deadline
- **THEN** the announcement shows that the Grinch escaped and no further damage is applied

### Requirement: Damage leaderboard

After a fight ends in victory, defeat, or escape, both interfaces SHALL show separate damage-point and gift-point rankings for registered players. Before the final row, damage points equal actual boss HP removed after phase floors. In the final row, ⚔️ awards 2 damage points for 1 HP removed. Only accepted actions count. The result SHALL include players who dealt zero damage and SHALL retain recorded damage after a restart.

#### Scenario: Rank contributions when combat ends
- **GIVEN** Alice dealt 20 HP, Bruno dealt 10 HP, and Carla dealt no damage
- **WHEN** the fight ends
- **THEN** the leaderboard lists Alice, Bruno, and Carla with 20, 10, and 0 HP respectively

#### Scenario: An attack reaches a phase floor
- **GIVEN** an attack would deal 5 HP but there is only 1 HP available above the current phase floor
- **WHEN** the attack is accepted
- **THEN** the attacker receives credit for 1 HP once

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
