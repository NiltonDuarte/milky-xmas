# Design

## Context

The Go service already shares domain behavior between Discord and the local chat and uses SQLite for production state.

## Decisions

- Configure one Discord text channel by ID. With no configured channel, the encounter is disabled. The Discord adapter observes human message creation in the configured guild and channel and passes only the activity timestamp to the application. Use the guild messages intent and leave message content access disabled.
- The shared application evaluates spawn eligibility with explicit clock and random inputs supplied by the shell. Discord runs a minute tick. A separate deadline check resolves an escape even when there are no clicks or fresh messages. After restart, wait for a fresh activity signal before rolling a new spawn.
- The functional core owns shared HP, phase eligibility, and terminal outcomes. The snowball window starts at appearance and closes ten seconds later for all members. Resolve that deadline separately from the overall encounter deadline. Distinct clicks each deal one damage without a per-member cooldown while health is above the 75% floor. Apply the floor against the latest stored health in the atomic update. Reaching the floor closes the attack phase immediately and replaces it with one defense board, even when several attacks arrive concurrently. Use the interaction timestamp to judge eligibility and keep damage updates atomic. The application associates each button with the configured guild, channel, encounter ID, announcement message, and member. Persist attack updates and terminal transitions atomically in SQLite. Interaction retries and concurrent attacks must not reduce HP twice for the same accepted click.
- Reserve an encounter before announcing it. Persist the returned message ID when confirmed. An ambiguous send remains non-playable and suppresses duplicate sends until it expires. For a confirmed fight, publish the latest saved state to the original message, serialize reconciliation to prevent older edits from overwriting newer ones, and retry failed edits through the scheduler.
- Use a private Discord acknowledgement for rejected attacks. The public announcement carries the snowball button, shared HP, phase deadline, and outcome. Disable the snowball button when the phase closes or the encounter ends. Generate the two rows of four emoji choices using injected randomness and persist the generated board with the phase transition. Route each button by its position on that board and apply its effect only to the clicking member. The local chat uses an isolated fixture with a deterministic clock and the same shared service. The encounter has three planned phases. The encounter leaves the existing gift balances independent.

## Risks and trade-offs

- Discord needs permission to view the selected channel and post announcements. A failed or ambiguous announcement should be logged and must not cause repeated public posts.
- A message can be deleted while a fight is running. A failed edit cannot reset HP or award a second victory. The recorded fight still resolves by its deadline.

## Open Questions

- Decide how phase two ends, how many defense choices each member can make, and what happens to a player who runs out of HP.
- Define phase three and its entry and victory conditions.
