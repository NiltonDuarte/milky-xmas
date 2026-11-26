## Decisions

- `core.AuraRules` decides the list and every perk; `app` only loads the aura and applies the answer.
- A claim, a watering, and a drop use the aura the player has when they act, before that action moves it.
- A fight remembers a player's perks from the moment they enter it, so a send during the fight changes nothing.
- The send limit covers both directions apart: lowering the aura can be farmed just like raising it. It lasts 24 hours from the last send that counted. Items still move.
- The limit is kept in `aura_sends`, written in the same transaction as the transfer, so it survives restarts and concurrent sends.
- A watering still counts when the aura cannot be read; it just gets no bonus.

- The daily limit follows the UTC date, like the Advent doors, and is kept on the player's row so a concurrent change cannot slip past it. A send stopped by the limit does not use up the send to that player.
- A Nice list player loses more HP on a wrong defense than anyone, as the price of more damage; the Naughty list loses less. The neutral cost doubles to 20% so that mistakes matter.

## Risks / Trade-offs

- [Two players can still trade once each way per day] → The daily limit caps what that is worth.
- [Nobody is on a list on the first day] → Entering takes two days of a 24 day event.
- [The damage bonus lets a Nice group clear a phase in five hits instead of six] → The number is tunable.
