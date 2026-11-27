## ADDED Requirements

### Requirement: Open a gift

`/open gift` SHALL take one gift of the chosen rarity and give one of its surprises, shown only to the player. Without that gift it SHALL be refused, changing nothing. Opening SHALL not move the aura. Both adapters apply this.

#### Scenario: Open a rare gift

- **GIVEN** a player with a rare gift
- **WHEN** they open it
- **THEN** they have one rare gift less and either three more Snowballs or an extra watering

#### Scenario: No gift

- **GIVEN** a player with no epic gift
- **WHEN** they try to open one
- **THEN** only they are told, and nothing changes

### Requirement: Shield

A kept shield SHALL stop the next wrong defense or counterattack against the player and then be gone.

#### Scenario: Blocked hit

- **GIVEN** a player with a shield misses a defense
- **WHEN** the fight is saved
- **THEN** they keep their HP and no longer have the shield

### Requirement: Extra watering

When the wait would refuse a watering, a kept extra watering SHALL let it count and be used up. It SHALL be kept if the tree refuses for another reason.

#### Scenario: Water twice

- **GIVEN** a player with an extra watering who just watered
- **WHEN** they water again
- **THEN** it counts, and a third try is refused
