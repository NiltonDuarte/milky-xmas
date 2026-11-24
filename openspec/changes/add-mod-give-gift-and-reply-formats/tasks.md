# Tasks

## 1. Moderator grants

- [x] 1.1 Add the core grant rule and the shared service method with compare-and-save. Test success, errors, conflicts, and unchanged aura.
- [x] 1.2 Register `/give gift` in Discord with the permission check and in the local chat. Test routing, rejections, and the private reply.

## 2. Reply formats

- [x] 2.1 Update the daily claim, inventory, and aura text and their tests.
- [x] 2.2 Attach the aura image in Discord and test the embeds. Update the README.

## 3. Verification

- [x] 3.1 Run format, vet, race tests, and strict validation. Record acceptance steps.

## Acceptance evidence

- Unit tests cover the core grant rule, the shared service (success, repeat, invalid recipient, rarity, quantity, unchanged moderator and aura), Discord `/give gift` (permission refusal, Administrator, rejections, private reply, registered permission), the local chat, the new reply text, and the aura reply's two embeds with one attached image.
- Live Discord acceptance is pending: run `/give gift`, `/daily claim`, `/socks`, and `/aura check` in the test server and check the formatting and image order.
