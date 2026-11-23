# Grinch inventory rewards

## Purpose

Award durable inventory items from the final Grinch board and from fights the players lose.

## ADDED Requirements

### Requirement: Final board snowballs

At 5% boss HP, both interfaces SHALL offer a ❄️ action alongside the existing final actions. Each accepted ❄️ pick SHALL credit one Snowball to the invoking player. Repeated requests, stale boards, and ineligible participants SHALL receive no credit.

#### Scenario: Earn multiple Snowballs

- **GIVEN** an eligible player in the final phase
- **WHEN** they pick ❄️ on two current rows
- **THEN** their inventory gains two Snowballs and the boss takes no damage from those picks

### Requirement: Rewards when the Grinch survives

When an active fight ends in defeat or escape, each registered participant SHALL receive one Coal and as many Yellow Snowballs as there were registered participants. Victory and cancelled lobbies SHALL award zero loss items. Ending a fight and awarding its items SHALL commit together.

#### Scenario: Four players lose

- **GIVEN** four registered players, including a player with zero HP
- **WHEN** the Grinch wins or escapes
- **THEN** each registered player gains one Coal and four Yellow Snowballs

#### Scenario: Reopen the encounter

- **GIVEN** a lost encounter has credited the participants
- **WHEN** the ending is processed again after restart or a competing request
- **THEN** each participant retains exactly the original credit

### Requirement: View reward balances

Both interfaces SHALL include Yellow Snowballs in `/socks` and show zero for inventories created before this item existed.

#### Scenario: View the rewards

- **GIVEN** a player earned one Coal and four Yellow Snowballs
- **WHEN** they use `/socks`
- **THEN** the list shows the credited quantities
