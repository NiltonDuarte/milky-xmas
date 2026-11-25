---
name: milky-develop
description: Implement Go bot behavior shared by the Discord and local browser adapters. Use when building features, fixing behavior, or preparing tests and user acceptance for an OpenSpec change.
---

# Develop Milky Xmas features

1. Read the root [AGENTS.md](../../../AGENTS.md). Tell the developer which OpenSpec change covers the work and what stage it is at. For behavior changes, read the scenarios before editing code and use them to decide which outcomes to verify. Follow the OpenSpec workflow described there.
2. Trace the interaction from the adapter through the shared application use case to the functional core and back. With the same input facts, both interfaces should reach the same decision and execute the same application steps. Select fixture or live dependencies at composition time. Keep transport parsing and response formatting in the corresponding adapter. When a step retries, verify it cannot repeat a state change or message where that would violate the scenario. For Discord handlers, acknowledge with a deferred response before any storage or network work and then update the user, as described in "Discord responses" in AGENTS.md.
3. Add table-driven tests for changed rules and edge cases. Pass fixed time and deterministic inputs where relevant. Use small fakes for shell tests that need to observe a write, message, or failure. Exercise the normal adapter entry point for changes to input or output mapping. Use `net/http/httptest` for the browser adapter and verify the fixture mode runs without credentials.
4. Run the checks in AGENTS.md. Give the developer repeatable acceptance steps with inputs and expected results. Use the local browser chat when available for shared behavior, and use a Discord test server for Discord-specific interaction details. Record observed results separately from steps awaiting execution.
5. Leave the change ready for the review described in [milky-review](../milky-review/SKILL.md), with the matching OpenSpec scenarios and verification evidence available.
