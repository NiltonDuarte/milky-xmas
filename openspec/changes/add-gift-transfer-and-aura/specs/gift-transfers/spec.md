# Gift transfers

## Purpose

Let players share daily Christmas gifts and see how giving affects their personal aura.

## ADDED Requirements

### Requirement: Gift inventory by rarity

The transferable inventory SHALL track normal, rare, and epic gifts per player and server. Each successful daily claim SHALL award its usual quantity in one rarity drawn with probabilities of 70% normal, 25% rare, and 5% epic. The reply SHALL identify the rarity. Existing daily gifts SHALL become normal gifts with their original cooldown.

#### Scenario: Receive a gift of a known rarity

- **GIVEN** a player is eligible for the daily claim
- **WHEN** the claim selects a quantity and rarity
- **THEN** the player receives that quantity in that rarity and the reply identifies both

#### Scenario: Transfer all gifts during the cooldown

- **GIVEN** a player has claimed a daily gift less than 24 hours ago
- **WHEN** they give away their entire inventory and attempt another claim
- **THEN** the original cooldown still applies

#### Scenario: Receive gifts before claiming

- **GIVEN** a player has never used the daily claim
- **WHEN** they receive a gift and then use the daily claim
- **THEN** they collect their first daily reward in addition to the received gift

#### Scenario: Upgrade an existing database

- **GIVEN** a player has daily gift data from before rarities were added
- **WHEN** the upgraded bot opens the database
- **THEN** the player can send that balance as normal gifts and keeps the original cooldown

### Requirement: Send a gift

Both interfaces SHALL support `/send gift` with a different recipient in the same server, one of the three rarities, and a positive integer quantity. A successful send SHALL debit and credit the same rarity and quantity and increase the sender's Good deeds by 5, 10, or 30 percentage points per normal, rare, or epic gift, capped at 100%. The reply SHALL show the recipient and the actual Good deeds increase.

#### Scenario: Transfer multiple gifts

- **GIVEN** a player has at least two rare gifts and both players have neutral aura
- **WHEN** the player sends two rare gifts to the other player
- **THEN** two rare gifts move between their inventories and the sender sees `Sent 2 🎁rare gifts to Bob. Your 😇 good deeds is 20% more high.` for a recipient named Bob
- **AND** the sender has 20% Good deeds and the recipient remains neutral

#### Scenario: Reject an invalid transfer

- **GIVEN** the player invokes the command with an invalid recipient, an unsupported rarity, a nonpositive quantity, or insufficient gifts of the selected rarity
- **WHEN** the command is processed
- **THEN** the player sees why the transfer was rejected and both inventories and auras stay the same

#### Scenario: Competing transfers

- **GIVEN** a player has exactly one epic gift
- **WHEN** two requests concurrently try to send that gift
- **THEN** at most one succeeds and only that transfer affects the inventory and aura

#### Scenario: Transfer fails to persist

- **GIVEN** a transfer is valid
- **WHEN** storage fails before the transfer commits
- **THEN** the player sees a failure and neither balance nor aura changes

### Requirement: Personal aura

Both interfaces SHALL support `/aura check` and show `😈 Bad <bad>%  |  😇 Good <good>%`. Both percentages SHALL start at 0%. A successful gift SHALL affect only the sender's Good deeds. Discord transfers and aura SHALL persist across restarts.

#### Scenario: Check neutral aura

- **GIVEN** a player has no recorded deeds
- **WHEN** they check their aura
- **THEN** they see `😈 Bad 0%  |  😇 Good 0%`

#### Scenario: Aura reaches its limit

- **GIVEN** a player has 90% Good deeds and owns an epic gift
- **WHEN** they send that gift and check their aura
- **THEN** the transfer confirmation reports a 10 percentage point increase and the aura check shows 100% Good

#### Scenario: Check aura after a restart

- **GIVEN** a Discord player successfully sent a gift
- **WHEN** the bot restarts and the player checks their aura
- **THEN** they see the same percentages and the completed transfer is reflected in both inventories

### Requirement: Discord native commands

The bot SHALL register `gift` under `/send` with required Discord user, rarity choice, and integer quantity options, and `check` under `/aura`. The bot SHALL acknowledge both commands and show their results privately. Bot accounts cannot receive gifts.

#### Scenario: Use the Discord commands

- **GIVEN** the bot is online in the configured server and the player can use application commands
- **WHEN** they choose the recipient, rarity, and quantity in `/send gift` or run `/aura check`
- **THEN** Discord shows the result only to the invoking player

#### Scenario: Select a bot account

- **GIVEN** a player selects a bot account as the recipient
- **WHEN** they submit `/send gift`
- **THEN** the transfer is rejected without changing either player's data
