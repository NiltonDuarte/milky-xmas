# Tasks

## 1. Calendar buttons

- [x] 1.1 Render the image and 24 buttons from the service's calendar, in rows of five. Test the row layout and the button states.
- [x] 1.2 Handle a click: open the door, redraw the private message, and answer stale or invalid buttons safely. Test an open, a repeat, a locked door, and invalid IDs.
- [x] 1.3 List uploaded files by position when an edit replaces attachments, including the Grinch final message.

## 2. Verification

- [x] 2.1 Run format, vet, race tests, and strict validation. Document the Discord acceptance steps.

## Acceptance evidence

- Adapter tests cover the layout (5, 5, 5, 5, 4), enabled, ❌, and disabled states on December 4, a click that opens door 3 and redraws the message, a stale click, a locked click, and invalid button IDs.
- Live Discord acceptance is pending: see the README.
