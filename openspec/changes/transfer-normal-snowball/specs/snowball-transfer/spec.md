# Snowball transfer

## Purpose

Make a normal snowball something the thrower owns and gives away.

## ADDED Requirements

### Requirement: A normal throw moves a Snowball

A normal `/throw snowball` SHALL take one Snowball from the thrower and give it to the recipient. A thrower with no Snowball SHALL be refused.

#### Scenario: Throw a Snowball

- **GIVEN** Alice has two Snowballs and Bob has none
- **WHEN** Alice throws a normal snowball at Bob
- **THEN** Alice has one Snowball and Bob has one
- **AND** Alice's Bad deeds does not change

#### Scenario: No Snowball to throw

- **GIVEN** Alice has no Snowballs
- **WHEN** she throws a normal snowball at Bob
- **THEN** she sees "You don't have a Snowball to throw." and neither inventory changes

#### Scenario: Recipient inventory is full

- **GIVEN** Bob's Snowball inventory is at its limit
- **WHEN** Alice throws a normal snowball at him
- **THEN** the throw is refused and Alice keeps her Snowball

#### Scenario: Type suggestions

- **GIVEN** a player owns Snowballs but no Yellow Snowballs
- **WHEN** they open the `type` choices in Discord
- **THEN** only the normal type is suggested, and yellow appears only when they own one

This applies to both interfaces. The local chat defaults to the normal type, and the Discord suggestions apply to Discord only.
