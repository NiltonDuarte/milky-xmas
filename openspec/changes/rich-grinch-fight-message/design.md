## Decisions

- **A message, not a modal.** A Discord modal is a form that opens for one person, cannot show images, and does not update. The panel is the fight message, edited as the fight changes, as it is today.
- **Components V2.** Text, the image, separators, and buttons all sit in containers. A V2 message cannot also carry plain `content` or embeds, so the fight message is sent with the V2 flag and edits clear those fields. Existing button IDs are kept.
- **Same builder for text and buttons.** `grinchMessage` keeps producing the phase text and the button rows. A new layout function wraps them in containers and adds the joined list, the cards, and the leaderboards, so the phase rules stay in one place.
- **Component budget.** A V2 message holds at most 40 components. The largest board (attack grid and Join) takes 16 with the main container, and eight cards take two each, a container and its text, plus one summary text. That is 33.
- **Cards.** The eight players with the most damage get a card with their HP and damage, marked by a coloured square for their HP. (The activity box of `add-grinch-activity-areas` later takes the cards into itself.) A summary line gives the number of other players, how many are still alive, and their combined damage. Ties break by player ID, as in the leaderboards.
- **Snowballs.** `Grinch.Snowballs` counts the Snowballs each player collected from the final ❄️ button. Old stored fights decode with no counts.
- **Standings.** Each leaderboard shows the top ten. When more players took part, the full standings are attached as a file, as before, and the message says so.
- **Mentions never ping.** The message allows no mentions, so listing players does not notify them.
