# Mission: samoa

Template repository for zoobzio Go packages.

## Purpose

Provide a complete, ready-to-use foundation for new Go libraries. Clone it, rename it, start building.

## What This Package Contains

- Go module configuration (1.24+)
- CI/CD workflows (test, lint, coverage, release, security)
- Tooling configuration (golangci-lint v2, codecov, goreleaser)
- Testing infrastructure (helpers, benchmarks, integration)
- Repository files (LICENSE, CONTRIBUTING, SECURITY)
- Skills for package maintenance
- Agent definitions for collaborative work

## What This Package Does NOT Contain

- Application code (this is infrastructure only)
- Package-specific logic
- README content beyond template instructions

## Success Criteria

A developer can:
1. Create a new repo from this template
2. Update MISSION.md and go.mod with their package name
3. Run `make check` successfully
4. Push and have CI pass
5. Start building their package immediately

## Non-Goals

- Opinionated application architecture
- Framework patterns
- Runtime dependencies
