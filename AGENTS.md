# Milky Xmas

Build the Milky community Christmas event bot in Go. Discord is the production user interface. The local browser chat is a second adapter for development and user acceptance. With Functional Core, Imperative Shell (FCIS), both adapters call the same application service, which owns the shared shell and calls the same functional core.

## Guide developers through OpenSpec

- For a new feature, a change in observable behavior, or a meaningful architecture change, run `openspec list` and check the relevant specs in `openspec/specs/`. Identify the matching change and explain how to continue it. For new work, offer `$openspec-propose "<idea>"` and explain what the proposal, design, spec scenarios, and tasks will capture. Documentation-only and behavior-preserving maintenance can proceed directly.
- When the developer asks only to plan, use the OpenSpec proposal or update skill. For an end-to-end implementation request, create or update the planning artifacts through the OpenSpec CLI as needed, then apply the change once it is ready. To implement an existing plan, use `$openspec-apply-change <change-id>`. Flag decisions that need developer input and explain the current stage when handing work back.
- Turn each changed behavior into observable Given/When/Then scenarios. Include failure paths that matter to users and identify which scenarios apply to both interfaces and which need Discord-specific checks. Update the artifacts if implementation changes an agreed behavior. Mark tasks complete when their work and relevant checks are complete.
- Validate the change with `openspec validate <change-id> --strict`. Report the OpenSpec change ID, test and acceptance evidence, and any remaining work. After review and applicable acceptance, guide the developer to `$openspec-archive-change <change-id>` and use it when asked to archive.

Use [milky-develop](.agents/skills/milky-develop/SKILL.md) for implementation and [milky-review](.agents/skills/milky-review/SKILL.md) for review. OpenSpec's generated skills handle its artifact lifecycle.

## Architecture

Add packages as features need them.

| Location | Responsibility |
| --- | --- |
| `src/internal/core` | Domain types and pure rules for event behavior. Take state, time, randomness, and other facts as explicit inputs where needed. Return decisions, next state, or typed errors. |
| `src/internal/app` | Shared use cases and effect orchestration. Fetch the inputs, call the core, then persist or deliver the result. Define small interfaces here when the use case needs an effectful dependency. |
| `src/internal/adapter/discord` | Translate Discord interactions to use-case inputs and render results as Discord responses. Own SDK types, acknowledgement, and transport details. |
| `src/internal/adapter/web` | Development-only HTTP adapter. Translate browser chat input into the same use-case requests and render results in a local transcript. |
| `src/cmd/` | Compose dependencies and manage configuration, startup, and shutdown for each executable. |

Share transport-independent validation, authorization, state loading, persistence, and effect orchestration wherever the use case is the same. Limit adapter-specific code to transport parsing, identity translation, response rendering, and protocol handling. Keep filesystem, network, storage, logging, and ambient time or randomness at the boundaries.

Select fixtures or live dependencies when wiring the application service. When live Discord messaging is needed, inject a Discord client as an outbound adapter and limit it to a configured test channel. Keep bot credentials on the server. See the [local acceptance design](README.md#local-acceptance-testing).

## Branches

- Create a new branch only when the developer asks for one or when the session starts in a worktree (which already has its own branch). Otherwise stay on the branch you are on.
- When working directly on `main`, stay on `main`: do not create or switch to another branch on your own.
- Keep `main` up to date with development. When working directly on `main`, commit and push often, at every point that is a meaningful user acceptance milestone (for example, a scenario that now works end to end, or a change the developer can try in Discord or the local chat). Run the checks from "Tests and review" first, and keep each commit small and focused. If you are unsure whether a point is a milestone, or whether the work is ready to push, ask the developer.

## Tests and review

- Run `go fmt ./...`, `go vet ./...`, and `go test -race -count=1 ./...` before handoff. GitHub Actions checks formatting, vet, tests, and OpenSpec validation.
- For product changes, cover the affected OpenSpec scenarios with deterministic core tests and exercise the shared use case through its shell where effect order or failure handling matters. Test parsing and response mapping at the adapters. Keep the default suite independent of Discord, credentials, and external services.
- Use the local browser chat for repeatable acceptance with known fixtures once it exists. Check registration, permissions, acknowledgements, and message presentation in a Discord test server when the feature touches them. Record steps, expected and observed results, and any checks still pending.
- Before merge, get a reviewer pass on the current diff for correctness, code smells, and FCIS boundaries. The author should resolve blocking findings and rerun affected checks. Record the review and relevant acceptance evidence in the pull request.
- Always merge pull requests with squash (`method: squash`), never a merge commit or rebase. For stacked pull requests, merge from the bottom up. After each squash, rebase the next layer onto `main` (`git rebase --onto main <old-base-branch>`), push, and retarget it to `main` before merging it.
