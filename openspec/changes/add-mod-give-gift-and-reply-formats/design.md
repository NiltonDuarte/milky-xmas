# Design

## Context

Gift transfers and claims share one player record and compare-and-save store. Grinch summoning already limits a moderator tool to Manage Messages.

## Decisions

- Add a pure core rule that adds gifts of one rarity to a record, with typed errors for unknown rarity, non-positive amount, and a full inventory. Aura, cooldown, and other balances are unchanged.
- The shared service loads, decides, and compare-and-saves the grant, retrying on conflict. Discord checks Manage Messages or Administrator in the adapter, as the Grinch command does, and also registers the command with that default permission. The local chat acts as the fixture moderator.
- Reply text lives in `gifttext`, shared by both adapters. Headings and code fences are Discord markdown and appear literally in the local chat.
- The aura reply sends two embeds: the image first, then the aura line in a code block. The image is embedded in the binary and attached to the private reply. The local chat shows the aura line only.
