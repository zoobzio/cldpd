# Testing Audit

Evaluate existing test infrastructure against standards and provide actionable recommendations.

## Principles

1. **1:1 mapping** — Every source file MUST have a corresponding test file
2. **Colocated tests** — Unit tests live next to source, infrastructure lives in `testing/`
3. **Helpers are tested** — Test utilities MUST themselves be tested
4. **Coverage includes integration** — CI MUST capture coverage from integration tests

## Execution

1. Read `checklist.md` in this skill directory
2. Work through each phase systematically
3. Compile findings into structured report

## Specifications

### 1:1 Test Mapping

- Every `.go` file MUST have a corresponding `_test.go` file
- No `_test.go` file should exist without a corresponding `.go` file
- Exception: Files containing only delegation, re-exports, or trivial wiring

### testing/ Directory

MUST contain:
- `testing/README.md` — Testing strategy overview
- `testing/helpers.go` — Domain-specific test helpers
- `testing/helpers_test.go` — Tests for helpers
- `testing/benchmarks/` — Performance tests
- `testing/benchmarks/README.md`
- `testing/integration/` — End-to-end tests
- `testing/integration/README.md`

### Helper Conventions

Helpers MUST:
- Have build tag: `//go:build testing`
- Call `t.Helper()` as first statement
- Accept `*testing.T` as first parameter
- Be domain-specific, not generic utilities

### Coverage Targets

| Metric | Target | Threshold |
|--------|--------|-----------|
| Project | 70% | 1% drop allowed |
| Patch | 80% | 0% drop allowed |

### CI Coverage

CI MUST:
- Run unit tests with coverage
- Run integration tests with coverage
- Merge coverage profiles before upload
- Upload combined report to Codecov

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall test infrastructure health and primary recommendation]

## Coverage Matrix

| Category | Status | Primary Issue |
|----------|--------|---------------|
| 1:1 Test Mapping | [✓/~/✗] | |
| testing/ Directory | [✓/~/✗] | |
| Test Helpers | [✓/~/✗] | |
| Benchmarks | [✓/~/✗] | |
| Integration Tests | [✓/~/✗] | |
| CI Coverage Config | [✓/~/✗] | |

## Missing Tests
[Source files without corresponding test files]

## Orphan Tests
[Test files without corresponding source files]

## Helper Issues
[Problems with test helper conventions]

## CI Coverage Gaps
[Integration tests not contributing to coverage]

## Quick Wins
[Low-effort fixes]
```

Status legend: ✓ Compliant, ~ Partial, ✗ Missing/Non-compliant
