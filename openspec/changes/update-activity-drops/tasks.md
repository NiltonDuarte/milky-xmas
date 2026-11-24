# Tasks

## 1. Rules and service

- [x] 1.1 Add core rules for the boosted rarity, the Coal chance, and Coal awards, with table-driven tests.
- [x] 1.2 Track opened slots, ordered collections, the ten-collection limit, and a single close in the shared service. Test boost, taken slots, Coal, and closing once.

## 2. Adapters

- [x] 2.1 Attach the images, disable opened buttons, and end drops by timer or tenth collection in Discord. Test each.
- [x] 2.2 Show Coal, the end of a rain, and `/drop expire <id>` in the local chat. Document acceptance steps.

## 3. Verification

- [x] 3.1 Run format, vet, race tests, and strict validation. Record acceptance evidence.

## Acceptance evidence

- Unit and adapter tests cover the rarity boost at 49% and 50% Good deeds, the Coal prize, the tenth-collection end, slot disabling, the timer expiry with the collector list, ending once, and both drop images.
- Live Discord acceptance is pending: see the README.
