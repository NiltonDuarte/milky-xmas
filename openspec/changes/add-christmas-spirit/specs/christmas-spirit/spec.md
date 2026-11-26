## ADDED Requirements

### Requirement: Server Christmas spirit

`/spirit` SHALL show the server's good deeds as the sum of all positive aura, the bad deeds as the sum of all negative aura, and which side wins. Both adapters apply this.

#### Scenario: Good deeds win

- **GIVEN** players with aura 30 and -10
- **WHEN** a moderator runs `/spirit`
- **THEN** it shows 30 good deeds, 10 bad deeds, and that the good deeds win

#### Scenario: Bad deeds win

- **GIVEN** the server's bad deeds are more than its good deeds
- **WHEN** a moderator runs `/spirit`
- **THEN** it says the bad deeds win

### Requirement: Moderator reveal

In Discord only players with Manage Messages SHALL use `/spirit`. The result SHALL be posted to the channel as one box with the picture, without pinging anyone. A refusal or failure SHALL be private. The interaction SHALL be acknowledged before any storage work.

#### Scenario: Player without permission

- **GIVEN** a player without Manage Messages
- **WHEN** they run `/spirit`
- **THEN** only they see that moderators can reveal it, and nothing is read

#### Scenario: Auras cannot be read

- **GIVEN** the database cannot be read
- **WHEN** a moderator runs `/spirit`
- **THEN** only the moderator sees that it failed
