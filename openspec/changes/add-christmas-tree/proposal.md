# Proposal

## Why

A shared Christmas tree gives the community something to care for together until the end of the event.

## What Changes

- Add `/water` so players water one Christmas tree shared by the server.
- The tree starts withered at 10% water. It needs 100% water to be full and alive, loses 1% every hour, and gains 2% per watering.
- Each player can water once every hour.
- The event ends on January 5, 2027 at 01:00 UTC. After that the tree keeps its final level and can no longer be watered.
- In Discord, a watering is a public message with the withered tree image, the water level, and the top waterers.

## Capabilities

### New Capabilities

- `christmas-tree` covers the shared tree and `/water`.

### Modified Capabilities

## Impact

A new core rule, application use case, and SQLite table hold the tree per server. Both adapters call the shared use case. Only the withered tree image exists so far, so every water stage uses it until more images are added.
