# Tasks

## 1. Implementation

- [x] 1.1 Implement and test recipient validation and atomic Snowball credit.
- [x] 1.2 Wire the command and notification in both adapters. Verify Discord registration, notification failure, and local chat behavior.

- [x] 1.3 Post the recipient notification in the Discord interaction channel and test routing, mention permissions, and failed posts.

## 2. Verification

- [x] 2.1 Run format, vet, race tests, and strict OpenSpec validation. Record acceptance evidence.

## Acceptance evidence

- The local executable processed two throws over HTTP. The simulated notification reported one and then two Snowballs. The sender's inventory still showed zero.
- Simulated Discord interactions verified the recipient selector, private sender acknowledgement, and commit before notification.
- Formatting, vet, the full race suite, and strict OpenSpec validation passed. The race suite used a temporary Zig C compiler.
- Live Discord acceptance awaits deployment to the test server. The README documents the two-account check.
- The revised Discord simulation verified that the notification goes to the command channel, only the selected player can be mentioned, and a missing send permission leaves the credit intact with a private warning.
