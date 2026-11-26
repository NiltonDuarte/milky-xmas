## ADDED Requirements

### Requirement: Finished fights are kept

The Discord bot SHALL keep every won, lost, or escaped Grinch fight, one record per fight.

#### Scenario: A later fight in the same channel

- **GIVEN** a finished fight
- **WHEN** another fight finishes in the same channel
- **THEN** both fights are stored

#### Scenario: Cancelled fight

- **GIVEN** nobody pressed Start
- **WHEN** the fight is cancelled
- **THEN** no fight record is added

### Requirement: Inventory changes are kept

The Discord bot SHALL keep, for each saved daily claim, drop claim, gift or Coal send, Snowball throw, moderator gift, Grinch reward, and watering that changes aura, the player, the reason, the change to each item and to aura, and the time. A transfer SHALL also keep the other player.

#### Scenario: Sending gifts

- **GIVEN** a player sends two rare gifts
- **WHEN** the transfer is saved
- **THEN** the history holds -2 rare for the sender and +2 rare for the receiver, each naming the other

#### Scenario: Change that is not saved

- **GIVEN** a change that is refused or loses to a concurrent one
- **WHEN** it is not saved
- **THEN** no history is added
