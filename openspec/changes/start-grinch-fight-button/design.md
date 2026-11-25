## Decisions

- **One button.** The lobby shows Start and the combat screen shows no join button. Start uses its own button ID, `start`, and the lobby move of the use case, so the Start press is the first one that also starts the countdown. Buttons with the older IDs, and fights stored in the older shape, are not handled.
- **Taking part.** `Grinch.Move` puts a player in the fight on their first accepted action. A refused action, such as a stale board or a click after the deadline, registers no one. The defense phase is sized by the players when it begins, so a player who arrives later takes part without changing that size.
- **Countdown in the core.** `Grinch.CountdownAt` holds the first press. `LobbyDeadline` is that time plus `registration.duration`. With no press yet it is the announcement time plus `registration.start_window`. A fight stored with players but no countdown, from before this change, keeps announcement time plus `registration.duration`.
- **Idle cancel.** The start window defaults to one minute, which keeps today's behaviour when nobody reacts. It is a tuning key so it can be made longer, since a countdown that begins on demand suits a longer wait.
- **Early start.** Reaching `early_start_players` starts the fight at once, as before.
- **Message.** Before the first press the announcement says to press Start and how long the countdown is. After it, the announcement shows when the fight starts.
