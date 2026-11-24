# Grinch announcement

## Purpose

Show the Grinch when an encounter appears.

## ADDED Requirements

### Requirement: Grinch image in the announcement

Discord SHALL attach the Grinch image to the announcement message when an encounter appears, whether it spawns on its own or a moderator summons it. Later updates to that message SHALL keep the image. The local chat has no announcement message and is unaffected.

#### Scenario: Spawned or summoned encounter

- **GIVEN** the Grinch appears in a configured channel, by activity or by `/grinch summon`
- **WHEN** the announcement is posted
- **THEN** the message shows the Grinch image with the existing registration text and button

#### Scenario: Image survives updates

- **GIVEN** an announcement with the image
- **WHEN** the fight state updates the message
- **THEN** the image stays on the message
