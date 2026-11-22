# Design

## Context

The Go service already shares domain behavior between Discord and the local chat and uses SQLite for production state.

## Decisions

- Configure one Discord text channel by ID. With no configured channel, the encounter is disabled. The Discord adapter observes human message creation in the configured guild and channel and passes only the activity timestamp to the application. Use the guild messages intent and leave message content access disabled.
- The shared application evaluates spawn eligibility with explicit clock and random inputs supplied by the shell. Discord runs a minute tick. A separate deadline check resolves an escape even when there are no clicks or fresh messages. After restart, wait for a fresh activity signal before rolling a new spawn.
- The functional core owns shared HP, phase eligibility, and terminal outcomes. The snowball window starts at appearance and closes ten seconds later for all members. Resolve that deadline separately from the one-minute overall deadline, which starts at appearance and applies to every stage. Distinct clicks each deal one damage without a per-member cooldown while health is above the 75% floor. Apply the floor against the latest stored health in the atomic update. Reaching the floor closes the attack phase immediately and replaces it with one defense board, even when several attacks arrive concurrently. Use the interaction timestamp to judge eligibility and keep damage updates atomic. The application associates each button with the configured guild, channel, encounter ID, announcement message, and member. Persist attack updates and terminal transitions atomically in SQLite. Interaction retries and concurrent attacks must not reduce HP twice for the same accepted click.
- Register `/grinch summon` as guild-scoped with Manage Messages as its default permission. Enforce the effective channel permission and configured channel again at invocation. The application uses the same atomic reservation as automatic spawns and skips only the automatic activity, odds, and cooldown checks.
- Reserve an encounter before announcing it. Persist the returned message ID when confirmed. An ambiguous send remains non-playable and suppresses duplicate sends until it expires. For a confirmed fight, publish the latest saved state to the original message, serialize reconciliation to prevent older edits from overwriting newer ones, and retry failed edits through the scheduler.
- Use a private Discord acknowledgement for rejected attacks. The public announcement carries the snowball button, shared HP, phase deadline, and outcome. Disable the snowball button when the phase closes or the encounter ends. Generate the two rows of four emoji choices using injected randomness and persist the generated board with the phase transition. Route each button by its position on that board and apply snowflake damage only to the clicking member. At the phase transition, persist the count of distinct players who landed a snowball, including the last hit. Successful defense damage is 30 percentage points divided by that fixed count. Track cumulative successful defenses as an integer and calculate exact cumulative damage before formatting HP for display. Apply the 45% floor against the latest stored boss HP in the same atomic update. The local chat uses an isolated fixture with a deterministic clock and the same shared service. The health ranges are 100% to 75% for snowballs, 75% to 45% for defense, and 45% to 5% for phase three. Transition out of defense atomically when boss HP reaches 45%. Generate and persist a new randomized attack board at that transition. Apply the emoji multiplier before enforcing the 5% floor. Render the shout with the board on the fight announcement. At 5% HP, switch atomically from the grid to the single final attack button. Clamp final damage at zero and close the fight at victory or the original one-minute deadline, whichever comes first. The encounter leaves the existing gift balances independent.

## Risks and trade-offs

- Discord needs permission to view the selected channel and post announcements. A failed or ambiguous announcement should be logged and must not cause repeated public posts.
- A message can be deleted while a fight is running. A failed edit cannot reset HP or award a second victory. The recorded fight still resolves by its deadline.

## Open Questions

- Decide how many defense choices each member can make and what happens to a player who runs out of HP.
- Set phase-three base damage, the phase-three shout text, and the final attack emoji and damage per click.
