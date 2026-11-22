# Proposal

## Why

An Advent calendar gives members 24 Christmas doors to open, including days they missed.

## What Changes

- Add `/advent view` and `/advent open <day>` to Discord and the local test chat.
- Let members collect one Advent gift per door from its December date until the end of December 24.
- Persist Advent progress and keep its rewards separate from daily gifts.

## Capabilities

### New Capabilities

- `advent-calendar` governs seasonal doors and Advent rewards.

### Modified Capabilities

## Impact

Extends the shared Go service, SQLite and fixture storage, and both command adapters.
