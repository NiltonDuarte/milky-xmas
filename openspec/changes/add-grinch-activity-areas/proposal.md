## Why

During the fight the panel shows HP and cards but not what is happening. The Grinch has no voice, and nobody sees who just attacked or what a miss cost.

## What Changes

- Two small text areas appear in the fight message while it runs.
- **The Grinch** shows the latest lines he says, picked at random from a pool for each moment: the start, each new phase, a player he hits with a snowball in the defense phase (with the HP they lost), a gift taken back (with the gift points), a Snowball taken back, and the players' defeat.
- **The players** shows who attacked last and the damage points they did, and a player who hides behind the tree in the defense phase says a line from a pool.
- The attack buttons of each phase sit in that box, below the two text areas and above the players' cards. The cards sit in the same box. The panel with the image has no buttons during combat.
- The two areas sit in a box of their own, apart from the panel with the image and buttons, and keep a fixed height so the message does not change size as events come in. The Grinch's area shows his latest three events and the players' area the latest five. The local chat shows the same lines.

## Impact

- Core: the fight keeps two short lists of recent events and a random roll supplied by the application layer. The core does not pick text.
- Application layer: `Play` supplies the roll. Adapters turn events and rolls into lines in one shared text package.
- Discord adapter: two text areas in the fight panel. They are not part of registration or the final message.
