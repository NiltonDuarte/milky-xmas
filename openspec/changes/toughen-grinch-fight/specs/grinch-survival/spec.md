## ADDED Requirements

### Requirement: Hits grow with the players

The snowball, attack, and final phases SHALL each take about eight hits per player in the fight, and at least fifteen, with the default numbers. The fight clock after the snowball phase SHALL be three minutes. This replaces the six hits per phase and the one-minute clock of earlier changes.

#### Scenario: Three players

- **GIVEN** three players in the snowball phase
- **WHEN** they bring the Grinch to 75% HP
- **THEN** it takes 24 hits

### Requirement: No wait by default

Actions SHALL count as fast as players press, unless `grinch.click_cooldown` sets a wait. The panel SHALL be refreshed right after an action.

#### Scenario: Quick second click

- **GIVEN** the default numbers and a player who hit a moment ago
- **WHEN** they hit again on the current buttons
- **THEN** the hit counts

### Requirement: Longer defense

The defense phase SHALL need four safe defenses for each player in the fight, at least eight and at most thirty. This replaces one per player up to six.

#### Scenario: Lone player

- **GIVEN** one player reaches the defense phase
- **WHEN** they defend
- **THEN** it takes eight safe defenses

### Requirement: The Grinch hits back

An accepted attack SHALL have a 10% chance of costing the attacker the HP of a missed defense. When nobody is left standing, the Grinch wins.

#### Scenario: Counterattack

- **GIVEN** a neutral player with 20% HP left
- **WHEN** their attack is answered
- **THEN** they are out of HP and the panel shows the hit

### Requirement: Getting back up with a gift

A player SHALL be able to spend one of their own gifts: when out of HP to get back up, and while standing to bring back the player who fell first, with 25% HP for a normal gift, 50% for a rare one, and 100% for an epic one. The panel SHALL name who a standing player's gift would bring back. It SHALL be refused, changing nothing, without that gift or with nobody down.

#### Scenario: Revival

- **GIVEN** Bob is out of HP and has a rare gift
- **WHEN** Bob presses the rare gift button
- **THEN** Bob has 50% HP and one rare gift less

#### Scenario: Helping a friend

- **GIVEN** Bob is out of HP and Alice is standing with a rare gift
- **WHEN** Alice presses the rare gift button
- **THEN** Bob has 50% HP and Alice has one rare gift less

#### Scenario: No gift

- **GIVEN** Bob has no epic gift
- **WHEN** he presses the epic gift button
- **THEN** only he is told he does not have one

### Requirement: Last chance

When nobody in the fight is standing, the Grinch SHALL win after 20 seconds unless somebody gets back up.

#### Scenario: Lone player

- **GIVEN** a lone player falls
- **WHEN** they use a gift within 20 seconds
- **THEN** the fight goes on

### Requirement: Moving Snowball button

The Snowball button of the final attack row SHALL be in a different place after each accepted pick. This replaces the fixed fifth button of earlier changes.

#### Scenario: After a pick

- **GIVEN** the final attack row
- **WHEN** a pick is accepted
- **THEN** the Snowball button is somewhere else in the row

