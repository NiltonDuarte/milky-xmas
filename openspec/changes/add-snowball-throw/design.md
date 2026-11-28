# Design

## Context

The existing player record stores Snowball balances and supports compare-and-save writes. Discord already resolves users for gift transfers and defers private command replies.

## Decisions

- Share validation of a different human recipient in the same server with gift transfers. The core adds one Snowball to the recipient inventory and guards the counter against overflow.
- The application retries concurrent inventory updates using the existing store and returns the balance after a successful commit.
- The Discord notification has one embedded picture per type: `snowballnormal.png` for a normal snowball and `snowballyellow.png` for a yellow one.
- Discord resolves the display names and posts to the interaction channel after commit, with a mention limited to the selected recipient. A failed channel post is reported in the private confirmation. The local chat displays both messages with fixture names.
