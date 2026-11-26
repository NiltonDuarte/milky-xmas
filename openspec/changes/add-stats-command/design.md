## Decisions

- Good eggs and Bad eggs add up the points of each deed, so they never cancel and have no limit: gifts 5, 10, or 30 each, a watering `tree.water_aura`, Coal and yellow snowballs 5 each. This differs from the aura, and from `/spirit`, which adds up the current aura.
- Opening the leaderboards reads nothing, so it updates the message directly instead of deferring first, which saves one request to Discord. Choices that take over 1.5 seconds are logged with their timing.
- `/spirit` keeps using the current aura, so it shows where the server stands on the last day of the event.
- Server stats leave out the good and bad totals so `/spirit` stays a surprise.
- Everyone may use `/stats`, and only they see it.
- Waterings come from the tree's counts; gifts, Coal, snowballs, and drops from the kept history; fights from the finished fights.
- Each leaderboard lists ten players and leaves out zeros.
- The local chat takes the choice as an argument instead of drop-downs.

## Risks

- Gifts, Coal, snowballs, drops, and fights from before the history was kept are not counted, while waterings are. Clearing the test data before the event removes the difference.
- Deed points use today's values, so changing them later also changes how past deeds count.
