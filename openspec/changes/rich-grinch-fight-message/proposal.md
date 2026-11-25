## Why

The Grinch fight message is plain text with an image embed. During the fight nobody sees who is fighting or how each player is doing, and the final ranking leaves out the Snowballs players collected.

## What Changes

- The fight message becomes a Discord Components V2 message: the Grinch image, the boss HP, and the action buttons in one panel.
- While players press Start, the panel lists who is ready.
- During the fight, the panel updates the boss HP and shows a small card for each of the eight players who dealt the most damage, with their current HP and damage. A summary line covers the rest.
- At the end, the panel shows the damage, gift, and Snowball leaderboards.
- The fight records the Snowballs each player collects, so the Snowball leaderboard exists in both interfaces.

## Impact

- Discord adapter: new message layout and the way the message is sent and edited. The buttons and their IDs do not change.
- Core: the fight keeps a per-player Snowball count and a Snowball leaderboard.
- Local chat: the final text adds the Snowball leaderboard. It has no cards or image.
- A fight already running when the bot is updated may keep its old layout until the next announcement. Check this in the Discord test server.
