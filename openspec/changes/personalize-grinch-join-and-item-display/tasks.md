# Tasks

## 1. Implementation

- [x] 1.1 Add Yellow Snowball aura and compact item messages with core and adapter tests.
- [x] 1.2 Add the private Discord Join panel and verify that two players can register independently.

## 2. Verification

- [x] 2.1 Run formatting, vet, race tests, and strict validation. Record acceptance.

## Acceptance evidence

- Simulated Discord callbacks opened separate private panels for two players, removed Join from each panel after enrollment, and silently acknowledged reopening after both players joined.
- Discord and local chat tests checked the yellow throw icon, inventory transfer, and five percentage point aura gain. A failed second throw did not increase the aura again.
- Both adapters verified the compact inventory and updated item icons. Formatting, vet, the full race suite, and strict validation passed.
- Live Discord verification of the two-player panel flow is pending deployment.
