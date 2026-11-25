## Decisions

- **Image.** The final edit uploads only Max's picture and the gallery shows only it. A win by the group keeps the Grinch's picture alone.
- **No permission to attach.** The bot cannot upload Max's picture, so the message has no picture when the Grinch won. The announcement's Grinch picture stays attached but is not shown.
- **Damage, not points.** `GrinchDamage` keeps the HP a player removed (`Hundredths`) apart from score points (`Points`). The damage leaderboard, the cards, and the players' area use the damage. A final hit removes at most the HP the Grinch has left. Points are accumulated only for final hits, at `final_attack.hit_score` each, and are not shown anywhere yet.
- **Scoring players only.** `core.WithPoints` drops players at zero from a leaderboard, in both interfaces. The full standings file uses the same filter. The "more than ten" check counts only players with points.
