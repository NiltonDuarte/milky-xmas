## Decisions

- Discord has no inline images, so each avatar becomes an application emoji, cut into a circle. Its name carries the user ID and a hash of the picture, so a restart finds it again and a new picture replaces it.
- Emojis are made in the background, one at a time, when a player speaks, uses the bot, or is first mentioned. Nothing waits for them.
- Emojis are added only while the message stays within Discord's text limit.

## Risks / Trade-offs

- [A player's first appearance can be without a picture] → It appears from the next message or update.
- [The bot keeps a copy of members' avatars as its own emojis] → Same service they already live on; replaced when the avatar changes.
- [An application holds 2000 emojis] → Far above the server's active players; a failure is only logged.
