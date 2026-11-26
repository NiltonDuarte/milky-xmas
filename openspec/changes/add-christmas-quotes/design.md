## Decisions

- A quote is due after `quotes.every_messages` player messages in a channel (default 3), and no sooner than `quotes.min_interval` after the channel's last quote (default 3h). Setting `every_messages` to 1 gives "every X minutes while people chat", so either rule can be tried without code changes.
- Nothing is posted in a quiet channel: only a message can bring a quote.
- Quotes are a separate channel choice, independent of commands and the boss. The setup option is optional, and leaving it out keeps the current choice.
- No quote repeats in a channel until all were shown.
- Only public-domain texts are included, because the repository is public.
- Message counts live in memory, like the Gift Rain counts. The time of each channel's last quote is saved, so a restart cannot bring a second quote inside the interval. If it cannot be read or saved, no quote is posted.

## Risks

- A failed post loses that quote; the next one comes after the same activity.
