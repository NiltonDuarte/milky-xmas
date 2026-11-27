## Decisions

- A separate command, not a topic of `/guide`, so Discord can hide it from players.
- Registered for Manage Messages. The handler also accepts Manage Server, which `/setup` needs, in case a server grants the command to such a role.
- Private, with a drop-down, like `/guide`. It holds no tokens, IDs, or other secrets.

## Risks / Trade-offs

- [A rule changes and the text does not] → Numbers follow the tuning; a new command or rule needs its topic updated.
