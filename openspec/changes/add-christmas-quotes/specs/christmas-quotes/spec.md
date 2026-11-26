## ADDED Requirements

### Requirement: Quotes follow the chat

The bot SHALL post a Christmas quote or poem after the configured number of player messages in a channel, and no sooner than the configured interval after that channel's last quote. A quote SHALL NOT repeat in a channel until all were shown. Both adapters apply this; the local chat uses `/message`.

#### Scenario: Enough activity

- **GIVEN** a quote is due every 3 messages
- **WHEN** players send 3 messages
- **THEN** the third one brings one quote with its author

#### Scenario: Too soon

- **GIVEN** the channel had a quote less than the minimum interval ago
- **WHEN** the message count is reached again
- **THEN** no quote is posted until a message arrives after the interval

#### Scenario: Restart inside the interval

- **GIVEN** a channel had a quote less than the minimum interval ago
- **WHEN** the bot restarts and players reach the message count
- **THEN** no quote is posted

#### Scenario: Quiet channel

- **GIVEN** nobody writes in a channel
- **WHEN** time passes
- **THEN** nothing is posted

### Requirement: Chosen channels only

In Discord quotes SHALL be posted only in channels where `/setup channel` turned `quotes` on, without pinging anyone. Messages from bots and webhooks SHALL NOT count. Leaving `quotes` out of `/setup channel` SHALL keep the current choice, and `/setup view` SHALL list the quote channels.

#### Scenario: Channel without quotes

- **GIVEN** a channel where quotes are off
- **WHEN** players chat there
- **THEN** no quote is posted

#### Scenario: Quotes-only channel

- **GIVEN** a channel with commands and boss off
- **WHEN** a moderator turns quotes on
- **THEN** quotes are posted there and commands stay disabled
