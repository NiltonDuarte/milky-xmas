# Tasks

## 1. Implementation

- [x] 1.1 Implement shared item transfer rules and verify inventory and aura effects.
- [x] 1.2 Wire both adapters and verify item selection, Discord suggestions, and reply formatting.

## 2. Verification

- [x] 2.1 Run format, vet, race, and strict validation and record acceptance evidence.

## Acceptance evidence

- Discord simulations verified that the Yellow option is visible to a player with stock and omitted when the balance is zero. The channel notification names the Yellow Snowball after the transfer.
- The local HTTP adapter verified that one Yellow Snowball moves between inventories and an attempt with an empty balance is rejected.
- Discord simulations verified two Coal gives 10% Bad deeds with the sender's Good deeds preserved. The local HTTP adapter verified insufficient stock is rejected and a send at 95% Bad deeds adds only five percentage points.
- Live Discord verification awaits deployment.
- Formatting, vet, the full race suite, and strict OpenSpec validation passed. The race suite used a temporary Zig C compiler.
