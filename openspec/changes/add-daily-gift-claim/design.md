# Design

## Context

The Discord bot and local chat already share an application service for `/hello`. Gift collection adds identity, time, and persistent state to the same FCIS architecture.

## Goals and scope

A gift is one unit in the member's collected total. The command is available while the bot is running. Seasonal scheduling, rarities, and trading are outside this change.

## Decisions

- The core computes eligibility from the last successful claim and the supplied current time. The application reads the clock once per attempt and uses instants in UTC. A rolling duration works across midnight and daylight saving changes without a scheduled reset.
- The shared application scopes claims by server ID and member ID. It loads the record, applies the core decision, and conditionally saves the new count and claim time against the observed record version. A competing update causes a reread and a fresh decision with the original request time. The response uses the committed total.
- Use a local SQLite database in production, with a unique record per server and member and an atomic conditional update. This protects claims on concurrent requests and across restarts. The local chat uses a fixed fixture identity, an in-memory store with the same atomic update contract, and an injected clock in tests.
- Discord maps the native `daily` command and `claim` option plus interaction identity to the shared use case. Defer an ephemeral response before storage work and edit it with the outcome. Discord can render the next eligible instant as a localized timestamp. The web adapter translates typed input and renders the same outcome in its transcript.

## Risks and trade-offs

- The production database must live on a persistent writable volume. The executable should report setup errors at startup and retain its existing guild-scoped command registration.
- Storage latency can affect the reply. A deferred Discord acknowledgement allows the work to finish within a bounded application deadline.
- The local fixture resets on restart. Production restart behavior is verified against a temporary SQLite file.
