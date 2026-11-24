# Design

## Context

Drops live in memory in `ActivityDrops`. Inventory awards are saved through the shared compare-and-save store.

## Decisions

- Pure core rules pick the rarity: Good deeds of 50% or more use 50% normal, 35% rare, 15% epic instead of 70, 25, 5. A Gift Rain collection is Coal on the first 25 of 100 rolls, otherwise it uses the daily rarity rolls. Coal never changes aura or cooldowns.
- The shared service tracks opened slots and the ordered list of collections. `Close` removes a drop and returns its list once, so the timer and the tenth collection cannot both announce the end. Collections after the end are refused as unavailable.
- Discord disables a button by editing the message components from the opened-slot bits returned by the claim. A slot taken by someone else at the same moment is refused with a private reply.
- Discord ends a drop with a timer set when it is posted, and the Gift Rain also ends on its tenth collection. The final edit removes the buttons, says the drop expired, and lists mentions without pinging. The local chat has no clock, so `/drop expire <id>` ends a drop on demand.
- Drops stay in memory, so a restart leaves an open drop's message unchanged.
