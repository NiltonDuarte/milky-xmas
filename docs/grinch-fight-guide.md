# The Grinch fight: full guide

For organizers and moderators. It describes the shipped numbers, which can be changed in `tuning.yaml`.

## How the Grinch shows up

- Every minute the bot may summon him on its own: 10% chance, only if someone chatted in the boss channel in the last 10 minutes, and at least 30 minutes after the previous fight.
- A moderator with **Manage Messages** can use `/grinch summon` at any time.
- Only one fight runs at a time.

## Getting started

- The announcement has a **🎅 Start** button. If nobody presses it within 1 minute, the fight is cancelled.
- The first press starts a 1-minute countdown. If 3 players press Start, the fight begins at once.
- There is no sign-up: anyone can click an attack button during the fight and is in from their first accepted click.

## The four phases

Six hits are enough to clear each damage phase, no matter how many players there are. The last hit of a phase only counts what is left.

| Phase | Boss HP | What players do | Per click |
| --- | --- | --- | --- |
| 1. Snowballs | 100% → 75% | **☃️ Attack** | 4.17 HP |
| 2. Defense | 75% → 45% | Hide behind a **🎄** | 30 HP shared by the defenses needed |
| 3. Attack | 45% → 5% | **🧊** or **☃️** tiles | 7 HP (🧊) or 6.67 HP (☃️) |
| 4. Final attack | 5% → 0% | **⚔️**, **🎁** or **❄️** | 0.84 HP (⚔️ only) |

**1. Snowballs.** Throw snowballs until the Grinch reaches 75% HP. There is a time limit: if it runs out, the Grinch wins.

**2. Defense.** The Grinch throws snowballs back. Eight tiles show 🎄 or ❄️ and reshuffle after every click.
- 🎄 is a safe defense. The group needs one per player who was in the fight when the phase began, up to 6.
- ❄️ means you were hit and lose 10% of your own HP. At 0 HP you are out. If every player is out, the Grinch wins at once.

**3. Attack.** Eight tiles, 🧊 hits a little harder than ☃️. The board stays the same for the whole phase.

**4. Final attack.** A row of five buttons that reshuffles after every click.
- ⚔️ damages the Grinch and earns 2 points (kept for later use).
- 🎁 earns 1 to 5 gift points, with no damage.
- ❄️ gives you 1 Snowball in your inventory, with no damage.

## Timers

- Phase 1 has its own limit: 2 minutes.
- When the Grinch reaches 75% HP, a second clock starts for phases 2 to 4: 1 minute.

## How it ends

| Result | When | Rewards |
| --- | --- | --- |
| Victory | The Grinch reaches 0% HP | Keep the Snowballs picked with ❄️ |
| Escape | A timer runs out | Each player gets 1 Coal and Yellow Snowballs equal to the group size |
| Defeat | Every player reaches 0 HP | Same as an escape |
| Cancelled | Nobody pressed Start | Nothing |

## The fight message

One message updates during the fight: the Grinch image, his HP, a box with what he says and what players did, the buttons, and a card with HP and damage for the 8 strongest players. At the end it shows the damage, gift, and Snowball leaderboards (top 10, only players who scored).
