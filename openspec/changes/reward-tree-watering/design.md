## Context

`/water` is public when it counts and private when it is refused, and Discord fixes the visibility of a reply when it is deferred.

## Decisions

- Defer privately. A watering that counts is posted to the channel as a normal bot message, then the private acknowledgement is deleted. Refusals edit the private reply. Trade-off: the public message has no "used /water" header, which the text makes up for by naming the player. Deferring publicly was rejected because every refusal would flash a public "thinking" message.
- If the channel post fails, the player sees the box privately. The watering still counts.
- The tree is saved first and the aura second. If the aura save fails, the watering stays and the failure is logged; the player gets no good-deed line. A shared transaction across the two stores is not worth it for one point.
- The aura moves like any other action and stops at +50, so a player already there sees no good-deed line.

## Risks

- One point per watering with a one-hour cooldown allows up to 24 points a day, so a player who waters often can reach +25, where drop odds improve, in about a day. `tree.water_aura: 0` turns the reward off and `tree.water_cooldown` slows it.
- The tree pictures were supplied by the developer and the repository is public; their usage rights are not recorded.
