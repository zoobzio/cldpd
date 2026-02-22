# Workspace Audit

Evaluate existing workspace structure for packages with providers, ensuring proper dependency isolation.

## Principles

1. **Core stays lightweight** — Root module MUST have minimal dependencies
2. **Providers are independent** — Each provider MUST be its own module
3. **Consumers pay for what they use** — Importing one provider MUST NOT bring others' deps
4. **Local development is seamless** — `go.work` MUST unify development experience

## Execution

1. Read `checklist.md` in this skill directory
2. Work through each phase systematically
3. Compile findings into structured report

## Specifications

### Required Files

- `go.work` — Workspace definition
- `go.mod` — Root module
- `*/go.mod` — Provider modules
- `testing/` — Test utilities (NOT a separate module)

### Root Module Requirements

- Module path: `github.com/zoobzio/[package]`
- Go version: 1.24+
- Toolchain directive present
- NO provider-specific SDKs in root go.mod

Red flags in root go.mod:
- Database drivers (pgx, mysql, mongo)
- Message queue SDKs (kafka, nats, amqp)
- Cloud SDKs (aws-sdk, google-cloud)
- testcontainers

### go.work Requirements

```go
go 1.24

toolchain go1.25.x

use (
    .
    ./[provider-a]
    ./[provider-b]
)
```

- Root module MUST be included (`.`)
- All provider modules MUST be included
- `go.work.sum` MUST be in .gitignore

### Provider Module Requirements

Each provider MUST have:
- Module path: `github.com/zoobzio/[package]/[provider]`
- Go version matching root
- Replace directive: `replace github.com/zoobzio/[package] => ../`
- Provider-specific SDK as dependency

Each provider MUST NOT:
- Import other providers
- Have root module dependencies leak in

### Release Workflow Requirements

- `tag-submodules` job MUST run before `release` job
- Submodule go.mod MUST be updated to reference released root version before tagging
- Tags MUST be created for each provider: `[provider]/v*.*.*`
- `GORELEASER_CURRENT_TAG` MUST be set in release job

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall workspace health and primary recommendation]

## Coverage Matrix

| Category | Status | Primary Issue |
|----------|--------|---------------|
| Root Module | [✓/~/✗] | |
| Go Workspace | [✓/~/✗] | |
| Provider Modules | [✓/~/✗] | |
| Dependency Isolation | [✓/~/✗] | |
| Testing Package | [✓/~/✗] | |
| CI/Release Config | [✓/~/✗] | |

## Dependency Analysis
[For each provider, what dependencies it brings]

## Isolation Issues
[Cases where dependencies leak between modules]

## Missing Replace Directives
[Providers not properly referencing root module]

## Quick Wins
[Low-effort fixes]
```

Status legend: ✓ Compliant, ~ Partial, ✗ Missing/Non-compliant
