## Why

Opening a door only counted an "Advent gift". The community wants each door to give something to read, such as a recipe, to be revisited later, and to differ between players.

## What Changes

- Clicking a door sends its surprise in a private message. Opened doors stay clickable and turn green.
- The 24 buttons appear in an order of each player's own, and the surprises are dealt in another order of their own.
- The Advent gift count is no longer shown.
- `/advent open` is removed from Discord. The calendar is used through `/advent view` and its buttons.
- Surprises are text files built into the bot. The first is a chicken salpicão recipe, and the other 23 are placeholders.
- The temporary test window until November 29, 2026 stays.

## Impact

- Core: a stable per-seed shuffle. The calendar view carries the button order instead of a total.
- Storage is unchanged: openings are recorded as before.
- Replaces the door, button, and total rules of `add-advent-calendar` and `add-advent-calendar-buttons`.
