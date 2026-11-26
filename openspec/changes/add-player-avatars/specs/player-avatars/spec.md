## ADDED Requirements

### Requirement: Avatar before a mention

In Discord, wherever the bot mentions a player in the leaderboards, the Grinch fight, the gift drop summaries, a watering, or a snowball notice, it SHALL show that player's round avatar before the mention once it has one. A missing avatar or a full message SHALL leave the mention as it was.

#### Scenario: Known player

- **GIVEN** the bot has a player's avatar
- **WHEN** they appear on a leaderboard
- **THEN** their picture is before their mention

#### Scenario: Avatar not ready

- **GIVEN** the bot has not made the player's avatar yet
- **WHEN** they are mentioned
- **THEN** the mention is shown without it and the avatar is made for later
