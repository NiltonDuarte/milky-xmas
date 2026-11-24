# Proposal

## Why

The Random Active Drop and Gift Rain should look like the event, reward kind players, give the rain a clear end, and report who collected what.

## What Changes

- Show the Grinch and Max image on the Random Active Drop and the gift rain image on the Gift Rain.
- Give players with Good deeds of 50% or more better odds of rare and epic gifts from a Random Active Drop.
- Disable the button a player opened in a Random Active Drop, so each of the eight gifts is taken once.
- End a Gift Rain at ten collections in total, with a 25% chance of Coal per collection.
- When a drop ends or expires, edit its message to say so and list who collected what.

## Capabilities

### New Capabilities

- `activity-drop-updates` covers the changes to both activity drops.

### Modified Capabilities

## Impact

The application service tracks opened slots and collections and closes drops once. The Discord adapter posts the images, disables buttons, and ends drops by timer or by the tenth collection. The local chat gains `/drop expire <id>`.
