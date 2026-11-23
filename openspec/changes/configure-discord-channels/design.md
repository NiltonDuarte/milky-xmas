# Design

## Context

The Discord process currently accepts all commands in a single configured guild and runs an optional encounter in one environment-configured channel.

## Decisions

- Persist channel roles and published introductions per guild. The application owns moderator authorization, persistence and announcement sequencing. The Discord adapter resolves channels, maps permissions, and posts messages.
- Register globally by default and retain the test guild setting for scoped registration. On guild availability, restore enabled boss channels and attempt outstanding onboarding.
- Manage one encounter worker per enabled boss channel. Keep the existing guild-and-channel storage key so a reselected channel can load its saved encounter. A removed channel stops its worker and refuses further button actions.
- The configuration commands are available in any channel that Discord permits. Boss channels can be selected independently of command channels. Manual summons require both.
