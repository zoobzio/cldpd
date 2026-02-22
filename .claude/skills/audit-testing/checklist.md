# Testing Audit Checklist

## Phase 1: Inventory

### Source Files
- [ ] List all `.go` files in package root (excluding `_test.go`)
- [ ] List all `_test.go` files in package root
- [ ] Identify missing test files (source without test)
- [ ] Identify orphan test files (test without source)

### testing/ Directory
- [ ] Does `testing/` directory exist?
- [ ] Does `testing/README.md` exist?
- [ ] Does `testing/helpers.go` exist?
- [ ] Does `testing/helpers_test.go` exist?
- [ ] Does `testing/benchmarks/` exist?
- [ ] Does `testing/benchmarks/README.md` exist?
- [ ] Does `testing/integration/` exist?
- [ ] Does `testing/integration/README.md` exist?

## Phase 2: 1:1 Mapping Assessment

### Completeness
For each source file, check for corresponding test file:

- [ ] All `.go` files have `_test.go` counterparts
- [ ] No orphan `_test.go` files
- [ ] Exceptions documented (delegation-only files)

### Exemption Criteria
Files may skip tests only if they contain:
- Pure delegation to other packages
- Re-exports only
- Trivial wiring with no logic

- [ ] Any exemptions are justified and documented

## Phase 3: Test Quality Assessment

### Test Conventions
- [ ] Tests use `testing` package (not third-party frameworks)
- [ ] Table-driven tests used for multiple cases
- [ ] Both success and failure paths tested
- [ ] `t.Run()` used for subtests
- [ ] `t.Parallel()` used where safe
- [ ] Race detector enabled (`-race` flag)

### Test Isolation
- [ ] Tests don't depend on execution order
- [ ] Tests clean up after themselves
- [ ] No shared mutable state between tests

## Phase 4: Helper Assessment

### helpers.go
- [ ] Build tag present: `//go:build testing`
- [ ] All helpers call `t.Helper()` as first statement
- [ ] All helpers accept `*testing.T` as first parameter
- [ ] Helpers are domain-specific (not generic utilities)
- [ ] Naming follows conventions:
  - `Assert[Condition]` for assertions
  - `New[Thing]` or `With[Thing]` for setup

### helpers_test.go
- [ ] Every exported helper has tests
- [ ] Both passing and failing cases covered
- [ ] Error messages verified to be helpful

## Phase 5: Benchmark Assessment

### Structure
- [ ] Benchmarks exist in `testing/benchmarks/`
- [ ] README documents how to run and interpret
- [ ] Baseline results documented

### Conventions
- [ ] Function naming: `Benchmark[Operation]`
- [ ] Memory benchmarks included (`-benchmem`)
- [ ] Benchmarks are deterministic

## Phase 6: Integration Test Assessment

### Structure
- [ ] Integration tests exist in `testing/integration/`
- [ ] README documents prerequisites and setup
- [ ] Tests are self-contained (setup/teardown)

### Conventions
- [ ] Tests skip gracefully if dependencies unavailable
- [ ] Build tags used if needed: `//go:build integration`
- [ ] External dependencies documented

## Phase 7: CI Coverage Configuration

### Coverage Workflow
- [ ] CI runs unit tests with coverage
- [ ] CI runs integration tests with coverage
- [ ] Coverage profiles are merged before upload
- [ ] Combined coverage uploaded to Codecov

### Check coverage.yml for:
```yaml
# Should include something like:
- name: Run tests with coverage
  run: |
    go test -coverprofile=coverage-unit.out ./...
    go test -coverprofile=coverage-integration.out ./testing/integration/...
    # Merge profiles
    echo "mode: atomic" > coverage.out
    tail -n +2 coverage-unit.out >> coverage.out
    tail -n +2 coverage-integration.out >> coverage.out
```

- [ ] Unit coverage captured
- [ ] Integration coverage captured
- [ ] Profiles merged correctly
- [ ] Codecov flags configured for unit/integration separation

### Makefile
- [ ] `make coverage` generates combined report
- [ ] Coverage includes integration tests
- [ ] HTML report generated

## Phase 8: Coverage Thresholds

### .codecov.yml
- [ ] Project target: 70%
- [ ] Project threshold: 1%
- [ ] Patch target: 80%
- [ ] Patch threshold: 0%

### Actual Coverage
- [ ] Current coverage meets 70% floor
- [ ] Recent patches meet 80% target

## Phase 9: Cross-Cutting Concerns

### Documentation
- [ ] testing/README.md explains strategy
- [ ] Benchmark README documents baselines
- [ ] Integration README lists prerequisites

### Consistency
- [ ] Test commands consistent between Makefile and CI
- [ ] Build tags consistent (`testing`, `integration`)
- [ ] Coverage flags consistent (`-covermode=atomic`)

### Runnable
- [ ] `make test` passes
- [ ] `make test-integration` passes or skips gracefully
- [ ] `make test-bench` runs successfully
- [ ] `make coverage` generates report
