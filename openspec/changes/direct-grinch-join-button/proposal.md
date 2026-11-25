> Superseded by `start-grinch-fight-button`: the Join button became Start, and there is no joining any more.

## Why

Joining the Grinch fight took two clicks: "Open registration" on the fight message, then Join on a private panel. The extra step slows players down and gives the Discord interaction more places to time out.

## What Changes

- The fight message shows a Join button that registers the player directly.
- A player who joins gets one private confirmation. Clicking Join again, or after registration is closed, shows nothing.
- The private registration panel is removed. Panels from before this change still work when clicked.
- Every click is acknowledged before the store is used, following the Discord response rules in AGENTS.md.

## Impact

- Replaces the "Personal Discord registration" requirement of `personalize-grinch-join-and-item-display`.
- Discord adapter only. The core rules for joining are unchanged, and the local chat has its own join command.
