## Why

The Grinch announcement counted a minute from the moment it appeared, whether or not anyone was ready. Players want to decide when the fight begins: press Start, and the fight begins a minute later, or at once when enough players press it.

## What Changes

- The Join button on the announcement becomes **Start**.
- The first press begins a one-minute countdown. Three players pressing Start start the fight at once, as three joins do today. Everyone who presses Start is in the fight.
- If nobody presses Start within a minute of the announcement, the encounter is cancelled, as when nobody joined.
- There is no joining. The combat screen has no Join button, and anyone can take part after the fight has started: a player is in the fight from their first accepted click on an attack button.
- Buttons of announcements posted before this change stop working, and the code that answered them is removed.
- The wait for the first press is a new tuning key, `grinch.registration.start_window`, one minute by default. `grinch.registration.duration` is now the countdown after the first press.

## Impact

- Core: the fight records when the countdown began, and the registration deadline follows it.
- Discord adapter: the button label and the announcement text. The button ID is unchanged.
- Tuning file and `docs/grinch-fight-numbers.md`: the new key. Fights stored before this change keep their old deadline.
- The local chat keeps `/grinch join`, which now starts the countdown.
