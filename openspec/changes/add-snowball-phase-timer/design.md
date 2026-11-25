## Decisions

- **Same ending as an escape.** When the snowball timer ends, `Advance` ends the fight as an escape at the deadline, like the other timers. The final message, Max's image, and the loss rewards follow the existing rules, so no new ending is needed.
- **Two timers.** The snowball timer runs from the start of combat. Reaching the milestone starts the existing combat timer from that click, so the group does not carry unused time forward.
- **Clicks after the deadline** are refused with the usual "this part of the fight has ended" before the fight is closed by the next sync.
- **Tuning.** `grinch.snowballs.duration` must be longer than zero.
