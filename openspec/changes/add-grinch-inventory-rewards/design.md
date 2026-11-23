# Design

## Decisions

- Append one dedicated snowflake to the existing final board. The core accepts it under the same membership, phase, board revision, and request ID rules as other final actions.
- Derive loss rewards from the final roster on the transition to defeat or escape. Track Yellow Snowballs as a separate balance.
- Save a reward-bearing transition and all matching inventory changes in one SQLite transaction. Use one fixture store and lock for the same operation locally. The encounter revision guards replay and contention.
- Keep using the configured server and channel for encounters, and pass the server ID to inventory storage. Add a zero-default Yellow Snowball column during schema initialization.
