## Decisions

- The damage of a hit is the phase's HP divided by the hits it needs, counted from the players in the fight at that moment, rounded up. Fallen players still count.
- With `grinch.hits` at zero the fixed damage numbers apply as before.
- The Grinch only answers attacks, not defenses, gifts, or Snowballs. A counterattack costs what a missed defense costs, so the aura lists keep their difference.
- A revival takes the gift in the same transaction that saves the fight. It does not move the aura or wait for the click cooldown.
- Gifts only bring players back; they do not heal a standing player.
- The last chance exists because a lone player would otherwise lose the moment they fell.
- A click during the wait is ignored. The panel says how long the wait is, rather than sending a message for every quick click.

## Risks / Trade-offs

- [A lone player needs fifteen hits per phase and is hit back more often] → Tunable; check in play.
- [More clicks mean more message edits] → The panel is already refreshed at most once a second.
