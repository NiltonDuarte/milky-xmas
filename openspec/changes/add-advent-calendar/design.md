# Design

## Context

The daily claim already provides server-scoped member identity, an injected clock, and durable SQLite storage.

## Decisions

- The core evaluates the December window from the injected UTC instant, and computes statuses for all 24 doors.
- The application provides view and open operations. Storage records one row per server, member, year, and day. The row itself is the reward, so a duplicate insert cannot award another gift. The daily table stays independent.
- SQLite and in-memory fixtures implement the same Advent storage contract. The Discord adapter registers native subcommands and uses private deferred replies. The chat accepts the equivalent typed commands.

## Risks and trade-offs

The UTC day may differ from a member's local date. Both interfaces identify the schedule as UTC. A browser fixture has transient progress, while Discord uses the existing database.
