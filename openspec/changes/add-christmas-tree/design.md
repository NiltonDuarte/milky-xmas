# Design

## Context

Per-server state with compare-and-save already exists for the Grinch encounter. Time and randomness enter the use cases as inputs.

## Decisions

- `core.Tree` holds the water level, the time the level was last current, each player's last watering, and each player's watering count. Decay is computed on read from whole hours since the level was last current, so nothing runs in the background and a restart loses nothing. A watering records the new level and time.
- Decay is 1 point per hour, a watering adds 2 points up to 100, the cooldown is one hour per player, and a new tree starts at 10%. A watering of a full tree is refused without a cooldown.
- The tree is stored as a JSON payload with a revision per server. `app.Tree` loads, decides in the core, and saves with the revision, retrying on conflict, so concurrent waterings are all kept.
- Water stages are 0-25, 26-50, 51-75, 76-99, and 100%. Only the 0-25% picture exists, so Discord uses it for every stage through one function that can map stages to pictures later.
- A successful Discord watering is public: the picture embed first, then the text, with no pings. Refusals are private. The local chat has no clock and no images and shows the same text.
- The event ends on January 5, 2027 at 01:00 UTC. From then on decay is computed up to that moment and stops, `/water` is refused, and the reply shows the tree's final state privately. An untouched tree still reports the starting level.
