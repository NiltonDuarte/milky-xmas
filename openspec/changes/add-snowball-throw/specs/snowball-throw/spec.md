# Snowball throw

## Purpose

Let players throw a snowball to add one item to another player's inventory.

## ADDED Requirements

### Requirement: Throw a snowball

Both adapters SHALL support `/thrown snowball`. A successful throw SHALL add one Snowball to a different human recipient in the same server. Throws SHALL require no sender inventory balance. The sender SHALL see `You have thrown a ❄️ snowball at <recipient>. 🥶`.

#### Scenario: First received snowball

- **GIVEN** Bob has zero Snowballs
- **WHEN** Alice throws a snowball at Bob
- **THEN** Alice sees `You have thrown a ❄️ snowball at Bob. 🥶`
- **AND** Bob's `/socks` shows one Snowball
- **AND** the channel message mentions Bob and says `Alice have threw a ❄️ snowball at you, now you have 1 ❄️ snowball. 🥶`

#### Scenario: Add to an existing balance

- **GIVEN** Bob already has one Snowball
- **WHEN** Bob receives another throw
- **THEN** Bob's `/socks` shows two Snowballs and the notification reports that balance

#### Scenario: Reject an invalid recipient

- **GIVEN** the selected recipient is the sender, a bot account, or outside the sender's server
- **WHEN** the throw is processed
- **THEN** the sender sees a rejection and no snowball is credited

#### Scenario: Inventory update fails

- **GIVEN** the inventory update cannot be persisted
- **WHEN** a player throws a snowball at a valid recipient
- **THEN** the sender sees a failure and the recipient receives no notification

### Requirement: Discord delivery

Discord SHALL register `snowball` under `/thrown` with one required user option. Discord SHALL respond privately to the sender and post a message mentioning the recipient in the channel where the command was used after persisting the credit. The web adapter SHALL show the simulated notification in the local transcript.

#### Scenario: Channel post unavailable

- **GIVEN** the bot cannot post in the command channel
- **WHEN** a throw is persisted
- **THEN** the recipient keeps the snowball and the sender is told the notification could not be posted
