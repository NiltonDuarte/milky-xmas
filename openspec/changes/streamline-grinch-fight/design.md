# Design

## Decisions

- Keep the encounter's existing per-player registration validation. Discord component state is shared across viewers, so acknowledge repeat interactions silently and keep the shared control available.
- Reuse the encounter's existing message edit path. Discord uses a deferred message update for routine clicks and an ephemeral reply if storage fails. The browser shows the current board and omits routine action confirmations.
- Share one HP formatter between adapters and preserve the terminal result and leaderboard.
