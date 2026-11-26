## ADDED Requirements

### Requirement: Picture per phase

The Discord fight message SHALL show the animated attack picture during the defense phase, the heart picture during the final attack, and the Grinch portrait during the other running phases.

#### Scenario: The Grinch attacks

- **GIVEN** the Grinch reaches 75% HP
- **WHEN** the defense phase starts
- **THEN** the fight message shows the animated attack picture

#### Scenario: Last 5%

- **GIVEN** the Grinch reaches 5% HP
- **WHEN** the final attack starts
- **THEN** the fight message shows the heart picture

#### Scenario: Same phase

- **GIVEN** the fight message already shows the phase's picture
- **WHEN** it is updated again
- **THEN** the picture is not uploaded again
