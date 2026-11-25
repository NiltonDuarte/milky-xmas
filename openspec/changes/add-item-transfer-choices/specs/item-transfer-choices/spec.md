# Item transfer choices

## Purpose

Let players transfer earned Coal and Yellow Snowballs using the existing social commands.

## ADDED Requirements

### Requirement: Choose a snowball type

`/throw snowball` SHALL accept normal or yellow. Discord SHALL require the type, like the player, and suggest yellow only when the sender owns one. The local chat SHALL default to normal when the type is left out. Yellow SHALL move one Yellow Snowball between inventories and name the chosen item in both replies. Normal SHALL move one regular Snowball between inventories, as `transfer-normal-snowball` defines.

#### Scenario: Throw an owned Yellow Snowball

- **GIVEN** the sender owns one Yellow Snowball
- **WHEN** they choose yellow and throw at another player
- **THEN** the sender loses one Yellow Snowball and the recipient gains one Yellow Snowball
- **AND** the confirmation and channel message identify it as yellow

#### Scenario: Yellow Snowball is unavailable

- **GIVEN** the sender owns no Yellow Snowballs
- **WHEN** they request suggestions or attempt to throw yellow
- **THEN** Discord omits the yellow suggestion and the throw is rejected without changing either inventory

### Requirement: Send Coal

`/send gift` SHALL offer Coal alongside the three gift rarities. A Coal send SHALL move the specified positive quantity between inventories and add five percentage points per Coal to the sender's Bad deeds up to 100%. The reply SHALL show the actual Bad deeds increase. Insufficient Coal SHALL leave both inventories and auras unchanged.

#### Scenario: Send two Coal

- **GIVEN** the sender owns two Coal and their Bad deeds is at 0%
- **WHEN** they send both to another player
- **THEN** the recipient gains two Coal and the sender's Bad deeds becomes 10%
- **AND** neither player's Good deeds changes

#### Scenario: Send Coal at the aura limit

- **GIVEN** a sender at 95% Bad deeds owns two Coal
- **WHEN** they send both Coal
- **THEN** the reply reports a five percentage point gain and the sender reaches 100% Bad deeds

#### Scenario: Send more Coal than owned

- **GIVEN** the sender has insufficient Coal
- **WHEN** they attempt to send it
- **THEN** the command reports the shortage and no balance or aura changes
