---
name: milky-review
description: Review Milky Xmas Go changes for correctness, code smells, FCIS architecture, and OpenSpec scenario coverage. Use for pull request review or a pre-merge review of implementation.
---

# Review Milky Xmas changes

Read the root [AGENTS.md](../../../AGENTS.md), the diff and relevant entry points, and the applicable OpenSpec design and scenarios. Review the current revision and trace a normal request and a relevant failure through the changed code.

- **Correctness and scenarios.** Check success and failure behavior, permissions, error handling, and the evidence for browser and Discord acceptance. Where retries, partial failures, or concurrent interactions are involved, check for duplicate effects and inconsistent state. If the browser chat can post to Discord, verify that live mode is explicit, credentials stay server-side, and the destination is limited to the configured test channel.
- **FCIS architecture.** Look for shell dependencies, ambient time, randomness, hidden configuration, mutation of shared inputs, or I/O inside the core. Check that both adapters call the same application service, that authorization, persistence, and effects follow a shared path, and that transport types stay in their adapters. Look for duplicate rules or orchestration in interaction handlers.
- **Code smells.** Look for duplicated decisions, long or deeply nested handlers, misleading names, dead branches, broad interfaces, speculative abstractions, and tests coupled to implementation details. Explain the concrete maintenance or correctness cost before marking a smell as a finding.
- **Evidence.** Run relevant local checks when possible and inspect CI and acceptance results when available. Check that each changed scenario has a test or a reproducible acceptance step appropriate to its boundary.

Lead with findings ordered by impact. For each, name the affected file and line, explain the behavior or maintenance cost, and suggest a focused fix. Distinguish merge blockers from follow-ups. When there are no blocking findings, say so and list verification gaps. Keep a review request read-only unless the developer asks for fixes.
