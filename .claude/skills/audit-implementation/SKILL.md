# Implementation Audit

Evaluate existing source code against zoobzio conventions, Go idioms, and internal consistency.

## Principles

1. **Code follows the spec** — Implementation MUST match architectural intent
2. **Patterns are consistent** — Same problem, same solution throughout the codebase
3. **Linter compliance is baseline** — golangci-lint MUST pass clean
4. **Godocs are complete** — Every exported symbol MUST be documented

## Execution

1. Read `checklist.md` in this skill directory
2. Work through each phase systematically
3. Compile findings into structured report

## Specifications

### Linter Compliance

The project uses golangci-lint v2 with these linter groups:

**Security:** gosec, errorlint, noctx, bodyclose, sqlclosecheck
**Quality:** govet (all analyzers), ineffassign, staticcheck, unused, errcheck, errchkjson, wastedassign
**Best practices:** gocritic, revive, unconvert, dupl (threshold 150), goconst (min-len 3, min-occurrences 3), godot, misspell, prealloc, copyloopvar

Additional settings:
- `errcheck.check-type-assertions: true`
- Test files exempt from: funlen, goconst, dupl

### Go Conventions

Code MUST follow:
- Effective Go naming (MixedCaps, not underscores)
- Receiver naming (short, consistent, not `this` or `self`)
- Constructor naming (`New[Type]`, `With[Option]` for functional options)
- Error variable naming (`Err[Condition]`)
- Package naming (lowercase, single word where possible, no underscores)

### Godoc Conventions

Every exported symbol MUST have a godoc comment:
- Package comment on one file per package
- Type comments describe what the type represents
- Function comments start with the function name
- Method comments start with the method name
- Comments are complete sentences ending with a period

### Context Usage

- All I/O functions accept `context.Context` as first parameter
- Context is not stored in structs
- Context cancellation is respected

### Error Handling

- All errors are checked (`errcheck` enforces this)
- Errors are wrapped with context: `fmt.Errorf("operation: %w", err)`
- Type assertions are checked (`errcheck.check-type-assertions`)
- No bare `_` for error returns in non-test code

### Dependency Usage

- stdlib preferred over external packages
- `reflect` usage is justified (not casual)
- `unsafe` usage is absent or heavily justified
- No `init()` functions without clear necessity

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall implementation health and primary recommendation]

## Coverage Matrix

| Category | Status | Primary Issue |
|----------|--------|---------------|
| Linter Compliance | [✓/~/✗] | |
| Naming Conventions | [✓/~/✗] | |
| Godoc Coverage | [✓/~/✗] | |
| Error Handling | [✓/~/✗] | |
| Context Usage | [✓/~/✗] | |
| Pattern Consistency | [✓/~/✗] | |
| Dependency Usage | [✓/~/✗] | |

## Linter Results
[Output of golangci-lint run, summarized by category]

## Naming Issues
[Violations of Go naming conventions]

## Godoc Gaps
[Exported symbols missing documentation]

## Error Handling Issues
[Unchecked errors, missing context, bare returns]

## Pattern Inconsistencies
[Same problem solved differently in different places]

## Quick Wins
[Low-effort fixes]
```

Status legend: ✓ Compliant, ~ Partial, ✗ Missing/Non-compliant
