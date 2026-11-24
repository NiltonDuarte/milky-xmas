# Design

## Context

Discord already observes gateway messages for the Grinch and stores player inventories through the shared application service. Channel settings mark where commands are enabled. See `proposal.md` and `specs/activity-gift-drops/spec.md` for the externally visible contract.

## Goals / Non-Goals

**Goals:**

- Keep activity thresholds, per-player eligibility, random rarity, and inventory persistence transport-independent.
- Render Discord drops as eight compact gift buttons and accept their component interactions.
- Count only human messages in channels that moderators enabled for commands.

**Non-Goals:**

- Persisting short-lived Discord drop controls across a bot restart.
- Adding moderator configuration for the drop frequency.
- Turning bot, webhook, or slash-command traffic into activity.

## Decisions

- A new application drop coordinator owns per-channel activity counters and in-memory live drops. The core exposes pure eligibility and inventory-award rules. This gives the browser adapter a reusable service surface while Discord owns gateway observation and message rendering.
- Random Active Drops use a fixed per-message chance. Gift Rain uses a newly drawn inclusive 5–100 message threshold per channel. This distinguishes the requested random active behavior from the explicit message-count trigger without a time-based background process.
- Random Active Drop tracks one successful claim per user. Gift Rain tracks each user's latest successful Gift Rain claim across active cards and channels, then applies the one-minute comparison in the shared service. Both use the existing compare-and-save inventory persistence, so duplicate component deliveries cannot award twice.
- Drops remain live for ten minutes in the running process. A restart also clears component eligibility, which is preferable to accepting stale Discord buttons with unknown state. The Discord adapter responds to stale controls with a generic private rejection.
- Discord presents each gift as an emoji-only button. Successful and rejected component acknowledgements are private. Gift Rain cooldown rejections omit the duration.

## Risks / Trade-offs

- [High message volume can create several cards] → The per-message probability is modest and the rain threshold is at least five messages.
- [Restart removes active card state] → Stale clicks are safely rejected and a later message starts fresh activity tracking.
- [In-memory activity cannot coordinate multiple bot processes] → The current runtime has one Discord process and inventory awards retain compare-and-save protection.

## Migration Plan

1. Deploy the updated binary with the existing SQLite inventory schema.
2. No existing player rows require migration because drops award the existing normal, rare, or epic fields.
3. Roll back by restoring the prior binary. Existing awarded gifts remain valid inventory.
