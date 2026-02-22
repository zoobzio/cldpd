# Workspace Audit Checklist

## Phase 1: Inventory

### Files Present
- [ ] `go.work` exists
- [ ] `go.mod` exists (root)
- [ ] List all `*/go.mod` files (provider modules)
- [ ] `testing/` directory exists
- [ ] `testing/` is NOT a separate module (no `testing/go.mod`)

### Provider Inventory
- [ ] List all provider directories
- [ ] Each provider has `go.mod`
- [ ] Each provider has implementation file
- [ ] Each provider has test file

## Phase 2: Root Module Assessment

### go.mod Analysis
- [ ] Module path follows `github.com/zoobzio/[package]`
- [ ] Go version is 1.24+
- [ ] Toolchain directive present

### Dependency Weight
- [ ] Count direct dependencies in root go.mod
- [ ] Identify any provider-specific SDKs that shouldn't be here
- [ ] Root should have minimal deps (interfaces + utilities only)

**Red flags in root go.mod:**
- Database drivers (pgx, mysql, mongo)
- Message queue SDKs (kafka, nats, amqp)
- Cloud SDKs (aws-sdk, google-cloud)
- testcontainers

## Phase 3: Go Workspace Assessment

### go.work Analysis
- [ ] Root module included (`.`)
- [ ] All provider modules included
- [ ] No missing providers
- [ ] No extraneous entries

### Workspace Sync
- [ ] `go work sync` succeeds without errors
- [ ] `go.work.sum` in .gitignore

## Phase 4: Provider Module Assessment

For each provider:

### Module Structure
- [ ] Has `go.mod` with correct module path: `github.com/zoobzio/[package]/[provider]`
- [ ] Go version matches root (1.24+)
- [ ] Has replace directive: `replace github.com/zoobzio/[package] => ../`

### Dependencies
- [ ] Has provider-specific SDK as dependency
- [ ] Does NOT import other providers
- [ ] testcontainers (if used) is only in this module, not root

### Implementation
- [ ] Implements core interface from root package
- [ ] Has corresponding `_test.go` file

## Phase 5: Dependency Isolation Analysis

### Cross-Provider Contamination
For each provider, check:
- [ ] Does importing this provider bring other provider SDKs?
- [ ] Run: `go list -m all` from a clean consumer importing one provider

### Consumer Simulation
```bash
# Create temp module
mkdir /tmp/test-consumer && cd /tmp/test-consumer
go mod init test
go get github.com/zoobzio/[package]/[provider]
go list -m all | wc -l  # Should be minimal
```

- [ ] Single provider import has reasonable dep count
- [ ] No surprise heavyweight dependencies

## Phase 6: Testing Package Assessment

### Structure
- [ ] `testing/` exists in root
- [ ] `testing/go.mod` does NOT exist (not a separate module)
- [ ] `testing/helpers.go` exists
- [ ] `testing/helpers_test.go` exists

### Dependencies
- [ ] helpers.go does NOT import provider-specific deps
- [ ] helpers.go provides mocks/utilities for core interfaces
- [ ] No testcontainers in testing package (those stay in providers)

### Build Tag
- [ ] helpers.go has `//go:build testing` tag

## Phase 7: Replace Directives

### Consistency Check
For each provider go.mod:
- [ ] Has `replace github.com/zoobzio/[package] => ../`
- [ ] Replace path is correct (relative to provider dir)
- [ ] No version pinned in replace (just path)

### Missing Replaces
- [ ] No provider missing replace directive
- [ ] No provider with incorrect replace path

## Phase 8: CI Configuration

### Workspace Transparency
- [ ] CI uses `go test ./...` (workspace handles multi-module)
- [ ] No manual workspace initialization needed
- [ ] Tests run across all modules automatically

### Multi-Module Validation
- [ ] CI validates root go.mod is tidy
- [ ] CI validates all provider go.mod files are tidy
- [ ] CI verifies no uncommitted changes to any go.mod/go.sum

### Release Workflow Structure
- [ ] release.yml exists
- [ ] Triggers on `v*.*.*` tags
- [ ] Has `validate` job (tests, lints, go mod tidy check)
- [ ] Has `tag-submodules` job
- [ ] Has `release` job (GoReleaser)
- [ ] Has `verify` job (confirms module fetchable)

### Submodule Tagging
- [ ] `tag-submodules` job runs before `release` job
- [ ] Tags created for each provider: `[provider]/v*.*.*`
- [ ] All providers listed explicitly in tagging script
- [ ] **CRITICAL**: Submodule go.mod updated before tagging
  - [ ] `go mod edit -require` updates root dependency to released version
  - [ ] `go mod tidy` runs after edit
  - [ ] Changes committed before tag is created
  - [ ] Commits pushed to main before tags pushed
- [ ] Tags pushed with `git push origin --tags`

**Why this matters:** Replace directives are ignored when a module is imported as a dependency. If the submodule's go.mod references a non-existent root version, external consumers cannot use the package.

### GoReleaser Configuration
- [ ] GORELEASER_CURRENT_TAG environment variable set
- [ ] Set to github.ref_name (GitHub Actions context)
- [ ] Prevents GoReleaser from picking up submodule tags

### Release Verification
- [ ] Verify job waits for module proxy (sleep 30s or similar)
- [ ] Creates temp module and attempts `go get`
- [ ] Release summary in GITHUB_STEP_SUMMARY

## Phase 9: Makefile

### Workspace Commands
- [ ] `make test` works across workspace
- [ ] `make tidy` tidies all modules
- [ ] `make sync` syncs workspace (optional)

## Phase 10: Cross-Cutting Concerns

### Consistency
- [ ] All providers use same Go version
- [ ] All providers use same toolchain
- [ ] All providers follow same structure

### Documentation
- [ ] README explains multi-module structure
- [ ] Import paths documented for consumers
- [ ] Provider-specific docs exist

### Verification
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` succeeds
- [ ] No import cycles detected
