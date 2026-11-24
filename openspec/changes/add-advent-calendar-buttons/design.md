# Design

## Context

`ViewAdvent` and `OpenAdvent` already decide door status by UTC date and record each opening once per player and year.

## Decisions

- The private `/advent view` reply carries the image and 24 buttons in rows of five, which fits Discord's limit of 25 buttons per message. Button state comes from the calendar the service returns: available doors are enabled, opened doors show ❌ and are disabled, and locked or expired doors are disabled.
- A click acknowledges the interaction, opens the door through `OpenAdvent`, reloads the calendar, and edits the same private message. The image is uploaded again and listed by position in the attachments so the edit replaces it. A stale button, such as one for a door that is already open or still locked, gets the service's normal answer and the buttons are redrawn from the current state.
- Buttons act for the member who clicked. They live in a private message, so only the player who ran the command sees them.
- The same position-based attachment list replaces attachments in the Grinch final message edit.
