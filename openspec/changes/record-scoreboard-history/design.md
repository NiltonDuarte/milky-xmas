## Decisions

- Write history in the same transaction as the change it describes, so they cannot disagree.
- Keep whole fight records instead of totals, so any fight scoreboard can be computed later.
- Waterings and Advent openings are already stored, so they get no new records. A watering at the top of the aura scale changes no inventory and is counted only by the tree.
- Cancelled fights have no players and are not kept.
- The local chat uses in-memory stores and keeps no history.

## Risks

- Changes made before this release have no history. Test data should be cleared before the event so it does not count.
