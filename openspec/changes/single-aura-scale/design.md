## Decisions

- **Same points as before.** The product owner kept 5, 10, 30, and -5, so two epic gifts take a neutral player to the top.
- **Reset, not convert.** The product owner chose to start everyone at 0 instead of Good minus Bad. The old columns are left in place, so the earlier values are not destroyed and an older build still starts.
- **Text bar.** Ten cells, five on each side of the centre, ten points each, inside a code block, with 👆🏻 on the next line under the marker. A number there was tried first and read as a stray value. The cells are box-drawing lines because the earlier squares were wider than a space, which put the pointer in the wrong column.
- **One box.** Every private reply with an image is a Components V2 container: image, text, then buttons. If Discord refuses it, the bot falls back to the two embeds so the player still gets the result. A failed Advent door now sends a private follow-up, because a boxed message has no text content to edit.
- **Drop threshold.** +25 is the middle of the good side, matching the earlier 50 of 100. It must be between 0 and 50.
