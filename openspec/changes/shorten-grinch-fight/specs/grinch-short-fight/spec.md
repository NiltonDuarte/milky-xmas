# Short Grinch fight

## Purpose

Let a group of any size finish the fight, and tell players when a drop is lost.

## ADDED Requirements

### Requirement: Six hits per phase

With the shipped numbers, each damage phase SHALL be cleared by at most six hits from the whole group, and the defense phase SHALL need one safe defense per player up to six.

#### Scenario: A single player

- **GIVEN** one player in the fight
- **WHEN** they land six snowballs
- **THEN** the Grinch is at 75% HP and the defense phase opens

#### Scenario: A large group

- **GIVEN** twenty players in the fight when defense begins
- **WHEN** six safe defenses are made
- **THEN** the Grinch is at 45% HP and the attack phase opens

### Requirement: A drop that runs out

A drop whose time runs out SHALL say that the Grinch stole all the presents, and SHALL list who collected.

#### Scenario: Time is up

- **GIVEN** an open drop with one collection
- **WHEN** its ten minutes pass
- **THEN** the message reads `Time's up! The Grinch stole all the presents!`, lists the collection, and has no buttons

These apply to both interfaces.
