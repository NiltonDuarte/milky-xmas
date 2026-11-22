# Tasks

## 1. Shared greeting

- [x] 1.1 Implement the pure greeting and the shared command service. Verify the greeting and unknown-command outcome with deterministic Go tests.

## 2. Interfaces

- [x] 2.1 Implement and document the loopback test chat. Verify page delivery, greeting, error recovery, and cross-origin write rejection through the HTTP handler and exercise the running web entry point.
- [x] 2.2 Implement and document the guild-scoped Discord bot. Verify registration and acknowledgement mapping with a fake transport, and check missing configuration through the executable.

## Verification

- In a headless Chrome session, the local chat showed the submitted command and `Hello, world!` for `/hello`. It showed the command hint for `/unknown` and accepted another `/hello` in the same transcript.
- The Discord transport tests verified a guild-scoped command registration and one public acknowledgement. Starting the binary with either required setting missing exited with the expected configuration error.
- Go formatting, vet, race tests, and strict OpenSpec validation passed.

## Workflow follow-up

- Run the documented live Discord acceptance in a configured test server before archiving.
