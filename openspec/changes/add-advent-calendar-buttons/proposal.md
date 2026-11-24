# Proposal

## Why

Players should open Advent doors by clicking the numbered calendar picture instead of typing the day.

## What Changes

- `/advent view` in Discord shows the Advent calendar image with 24 numbered buttons, one per day.
- Clicking an available day opens it once. The button becomes ❌ and is disabled.
- Doors that cannot be opened yet or any more are disabled.
- `/advent open <day>` and the local chat keep working as before.

## Capabilities

### New Capabilities

- `advent-calendar-buttons` covers the Discord button calendar.

### Modified Capabilities

## Impact

The Discord adapter renders and handles the buttons through the existing service. The Advent rules, storage, and rewards do not change.
