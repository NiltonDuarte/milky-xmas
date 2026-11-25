## Decisions

- **Role by name.** The bot looks for a role named `elfs` (ignoring case) and creates it when none exists, so a moderator can create or recolor it first. It stores no role ID.
- **Private, deferred reply.** The click is acknowledged with a private deferred reply before any role call, then edited into the result or the error, following the Discord responses guideline. The role calls go to Discord and can exceed three seconds.
- **No state.** Whether the player is already an elf comes from the roles on the interaction, so repeated clicks make no extra Discord call.
- **Where the button sits.** The introduction is the message that ends setup. The setup poster adds the button row to that message only, and to no other setup text.
- **Failure messages** tell the player what a moderator should check: the Manage Roles permission, or the bot's role being above `elfs`.
