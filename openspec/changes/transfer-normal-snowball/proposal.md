## Why

A normal snowball was created from nothing on every throw, so players could pile Snowballs on each other for free. The yellow snowball already moves between inventories. The normal one should work the same way, so a snowball is something you have to own to throw.

## What Changes

- A normal throw takes one Snowball from the thrower and gives it to the recipient.
- A player with no Snowball cannot throw a normal one. The throw is refused and nothing changes.
- A recipient at the inventory limit refuses the throw and the thrower keeps the Snowball.
- Discord suggests the normal type only to a player who owns a Snowball, as it already does for yellow.
- A normal throw does not change Bad deeds. Only the yellow snowball does.

## Impact

- Core: a `TransferSnowball` rule next to `TransferYellowSnowball`. The old credit-only rule is removed.
- Both adapters: a new refusal text, and the Discord type suggestions.
- Players need Snowballs from other sources, such as Grinch fight rewards, before they can throw one.
- Replaces the "no sender inventory balance" rule of `add-snowball-throw`.
