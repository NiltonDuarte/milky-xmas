## ADDED Requirements

### Requirement: Hits grow with the players

The snowball, attack, and final phases SHALL each take about eight hits per player in the fight, and at least fifteen, with the default numbers. The fight clock after the snowball phase SHALL be three minutes. This replaces the six hits per phase and the one-minute clock of earlier changes.

#### Scenario: Three players

- **GIVEN** three players in the snowball phase
- **WHEN** they bring the Grinch to 75% HP
- **THEN** it takes 24 hits

### Requirement: Wait between actions

A player SHALL wait two seconds between accepted actions of a fight. Earlier clicks SHALL not count.

#### Scenario: Quick second click

- **GIVEN** a player hit one second ago
- **WHEN** they hit again
- **THEN** nothing changes

### Requirement: The Grinch hits back

An accepted attack SHALL have a 10% chance of costing the attacker the HP of a missed defense. When nobody is left standing, the Grinch wins.

#### Scenario: Counterattack

- **GIVEN** a neutral player with 20% HP left
- **WHEN** their attack is answered
- **THEN** they are out of HP and the panel shows the hit

### Requirement: Getting back up with a gift

A player out of HP SHALL be able to spend one of their own gifts to get back up with 25% HP for a normal gift, 50% for a rare one, and 100% for an epic one. It SHALL be refused, changing nothing, without that gift or while standing.

#### Scenario: Revival

- **GIVEN** Bob is out of HP and has a rare gift
- **WHEN** Bob presses the rare gift button
- **THEN** Bob has 50% HP and one rare gift less

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
