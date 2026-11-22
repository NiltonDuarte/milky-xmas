# Daily gift claim

## Purpose

Let Milky Xmas members collect gifts over time through a claim command with a predictable cooldown.

## ADDED Requirements

### Requirement: Rolling gift cooldown

Discord and the local test chat SHALL support `/daily claim`. Each successful claim SHALL add exactly one gift to the member's total for that server. Eligibility SHALL return exactly 24 hours after the last successful claim.

#### Scenario: First claim

- **GIVEN** a member has never claimed a gift in the server
- **WHEN** they submit `/daily claim`
- **THEN** they receive one gift and see a total of one

#### Scenario: Claim before the cooldown ends

- **GIVEN** a member claimed a gift less than 24 hours ago
- **WHEN** they submit `/daily claim`
- **THEN** they see when the next gift becomes available, and their total and next eligible time stay the same

#### Scenario: Claim exactly when the cooldown ends

- **GIVEN** a member claimed their first gift at 18:00 UTC yesterday
- **WHEN** they submit `/daily claim` at 18:00 UTC today
- **THEN** they receive one gift, see a total of two, and start a new 24-hour cooldown

#### Scenario: Independent member totals

- **GIVEN** one member has already claimed a gift and a second member has never claimed a gift in that server
- **WHEN** the second member submits `/daily claim`
- **THEN** the second member receives their first gift with a total of one

#### Scenario: Simultaneous attempts

- **GIVEN** a member is eligible to claim
- **WHEN** two requests for that member and server are processed concurrently
- **THEN** exactly one request awards a gift and the other reports the cooldown

### Requirement: Durable production claims

The Discord bot SHALL preserve successful claims across restarts and SHALL confirm a gift only after recording the new total and claim time together.

#### Scenario: Restart during the cooldown

- **GIVEN** a member claimed one gift and the bot restarts before the cooldown ends
- **WHEN** the same member claims again
- **THEN** the bot reports the original eligibility time
- **AND** when that time arrives, the next successful claim reports a total of two

#### Scenario: Claim cannot be saved

- **GIVEN** a member is eligible to claim and their claim cannot be saved
- **WHEN** they submit `/daily claim`
- **THEN** the bot reports that it could not record the claim
- **AND** a subsequent attempt after storage recovers can collect the gift

### Requirement: Discord native command

The bot SHALL register `claim` as a subcommand of `/daily` in the configured Discord server and SHALL show the result privately to the member who invoked it.

#### Scenario: Select the claim subcommand

- **GIVEN** the bot is online in the configured server and the member can use application commands
- **WHEN** the member selects `claim` under `/daily` and submits it
- **THEN** Discord displays the claim result only to that member
