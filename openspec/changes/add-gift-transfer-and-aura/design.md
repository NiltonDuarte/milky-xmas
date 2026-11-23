# Design

## Context

The daily reward stores a server-scoped balance and a timestamp. The browser uses an in-memory store, and Discord uses SQLite. Both adapters call the same service.

## Goals

- Preserve the daily claim's quantity range and 24-hour cooldown through the data migration and subsequent transfers.
- Commit both inventory changes and the sender's aura in one operation.

## Decisions

- The functional core owns rarity probabilities, typed balances, cooldown rules, and aura arithmetic. The shared service validates recipient and gift inputs, loads records, and retries if a concurrent operation changed either snapshot. The cooldown follows the last claim timestamp even when an inventory is empty.
- One rarity roll per daily claim keeps the existing single award and gives players a way to acquire all three rarities. Display the rarity alongside the existing award message.
- A single player row stores per-rarity balances, aura, and claim time. SQLite compares complete snapshots before writing and wraps both player writes in one transaction. The fixture store provides the same atomic behavior under its mutex.
- Discord maps the selected user, rarity, and integer option to the shared service and reuses private deferred replies. The browser accepts `/send gift <recipient> <rarity> <quantity>` with simulated recipients and uses its existing fixture player for `/aura check`.

- The shared service returns the actual Good deeds increase from the committed transfer. Discord resolves the recipient's nickname or account display name, and the browser uses the recipient name typed into the command.

## Risks and trade-offs

- Concurrent claims and transfers can touch the same player. Compare the claim time and all balances and aura fields so a stale request cannot overwrite another player's updates.
- A success response can fail after a commit. The transfer remains recorded, and each new invocation represents its own requested transfer.

## Migration Plan

At startup, convert all old balances to normal gifts in a SQLite transaction and carry over their claim timestamps. A completed migration can be reopened without reimporting those balances. Keep a pre-upgrade backup if rollback to the old schema is needed.
