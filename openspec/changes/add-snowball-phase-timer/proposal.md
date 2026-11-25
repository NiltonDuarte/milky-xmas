## Why

The first phase of the fight had no time limit, so a group could throw snowballs for as long as it liked, or leave the fight open and idle. A limit gives it tension and a clear end.

## What Changes

- The group has two minutes, from the start of combat, to bring the Grinch down to the snowball milestone (75 HP).
- If time runs out first, the Grinch wins. The fight ends as an escape, with the same loss rewards.
- If the group reaches the milestone in time, the fight goes on as before, and the one-minute combat timer starts at that moment.
- The time is a tuning key: `grinch.snowballs.duration`, two minutes by default.

## Impact

- Core: the snowball phase has a deadline. The existing end-of-fight rules handle the rest.
- Discord message: the snowball phase shows when the Grinch escapes and says time can run out.
- A snowball fight stored before this change ends as soon as the bot starts if it began more than two minutes ago.
