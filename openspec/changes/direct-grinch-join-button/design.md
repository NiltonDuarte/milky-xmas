## Decisions

- **Silent acknowledgement first.** The click is acknowledged with `DeferredMessageUpdate`, so nothing changes on screen, and then the shared `Play` use case runs with the lobby move. The private confirmation is a follow-up.
- **One confirmation.** The player's click is the only request ID the core marks as accepted when it registers them, so the adapter sends the confirmation only when `Accepted[interaction ID]` is set. A repeat click gets the core's "already joined" reply, which the adapter does not send.
- **Refusals stay silent.** Closed registration or a finished fight shows nothing, as before. Store failures send a private "Could not update your registration" follow-up.
- **Same custom ID.** The button keeps `grinch:<fight>:registration:0`, so announcements already posted keep working after a restart. Old private panels keep their `join` button and use the same path.
