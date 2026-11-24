# Tasks

## 1. Rules and storage

- [x] 1.1 Add the core tree rules (decay, watering, cooldown, stages, top waterers) with table-driven tests.
- [x] 1.2 Add the application use case with compare-and-save, a memory store, and a SQLite table. Test concurrent waterings and persistence.

## 2. Adapters

- [x] 2.1 Add `/water` to Discord with the public watering message and image and private refusals. Test both.
- [x] 2.2 Add `/water` to the local chat, list it in `/helpmemilky`, and document acceptance steps.

## 3. Verification

- [x] 3.1 Run format, vet, race tests, and strict validation. Record acceptance evidence.
- [ ] 3.2 Add pictures for the 26-50%, 51-75%, 76-99%, and 100% stages when they are available.

## Acceptance evidence

- Tests cover decay by the hour, the 100% cap, the full tree, the cooldown and its boundary, other players and servers, 20 concurrent waterings, SQLite reload and stale revisions, the Discord public and private replies with the image, and the local chat.
- Live Discord acceptance is pending: see the README.
