# Snowball notification

## Purpose

Present the snowball a player receives with an image and a short message.

## ADDED Requirements

### Requirement: Received snowball message

The message telling a recipient about a thrown snowball SHALL read `<recipient>❕  <sender> has thrown a <snowball type> at you. 🥶`. In Discord the recipient is a mention and the message SHALL include the snowball image. The local chat shows the same text without the image. The message SHALL NOT report the recipient's balance.

#### Scenario: Recipient is notified with the image

- **GIVEN** Alice throws a normal snowball at Bob in a Discord channel
- **WHEN** the throw is recorded
- **THEN** the channel message mentions Bob, says `Alice has thrown a ❄️ snowball at you. 🥶`, and shows the snowball image

#### Scenario: Yellow snowball

- **GIVEN** Alice throws a Yellow Snowball at Bob
- **WHEN** the throw is recorded
- **THEN** the message names the yellow snowball type
