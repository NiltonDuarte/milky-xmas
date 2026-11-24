# Proposal

## Why

Moderators need a way to hand out gifts, and the daily claim, inventory, and aura replies should look like the event's intended Discord presentation.

## What Changes

- Add `/give gift <recipient> <amount> <rarity>` for moderators. It adds gifts to the recipient without touching any inventory or aura of the moderator.
- Reformat the `/daily claim`, `/look in-your-stocking`, and `/aura check` replies with a heading and code block.
- Rename `/socks` to `/look in-your-stocking`. Discord registers `/look` with the subcommand `in-your-stocking` because command names cannot contain spaces, and removes the old `/socks` command from the server.
- Show the stocking image above the inventory in Discord.
- Attach the Grinch image to the encounter announcement in Discord.
- Add Max's image to the final message when the Grinch survives.
- Add `/helpmemilky` to list the commands available to the player, including moderator commands only for members with the permission.
- Show the "bad good" image above the aura line in Discord.
- Change the snowball notification to `<recipient>❕  <sender> has thrown a <type> at you. 🥶` with the snowball image.

## Capabilities

### New Capabilities

- `mod-give-gift` covers moderator gift grants.
- `reply-formats` covers the new reply presentation.

### Modified Capabilities

## Impact

Both adapters call a new shared service method. Discord registers `/give` limited to members with Manage Messages. The aura image is embedded in the Discord adapter.
