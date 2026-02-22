# Implementation Audit Checklist

## Phase 1: Linter Compliance

### Run Linter
- [ ] Run `golangci-lint run ./...`
- [ ] Record total findings by linter
- [ ] Note any disabled linters or nolint directives

### Security Linters
- [ ] gosec: no security findings
- [ ] errorlint: error comparisons use `errors.Is`/`errors.As`
- [ ] noctx: no HTTP requests without context
- [ ] bodyclose: all response bodies closed
- [ ] sqlclosecheck: all SQL rows/statements closed

### Quality Linters
- [ ] govet (all analyzers): no issues
- [ ] ineffassign: no ineffectual assignments
- [ ] staticcheck: no static analysis findings
- [ ] unused: no unused code
- [ ] errcheck: all errors checked
- [ ] errchkjson: JSON marshal/unmarshal errors checked
- [ ] wastedassign: no wasted assignments

### Best Practice Linters
- [ ] gocritic: no diagnostic findings
- [ ] revive: no style findings
- [ ] unconvert: no unnecessary conversions
- [ ] dupl: no duplicated code blocks above threshold (150 tokens)
- [ ] goconst: repeated strings extracted to constants
- [ ] godot: all comments end with a period
- [ ] misspell: no misspellings
- [ ] prealloc: slices preallocated where size is known
- [ ] copyloopvar: no loop variable capture issues

### Nolint Directives
- [ ] List all `//nolint` directives
- [ ] Each directive specifies which linter(s) it suppresses
- [ ] Each directive has a justification comment
- [ ] No blanket `//nolint` without linter names

## Phase 2: Naming Conventions

### Package Names
- [ ] Lowercase, single word where possible
- [ ] No underscores or mixed caps
- [ ] Name reflects purpose (not util, common, misc)

### Type Names
- [ ] MixedCaps (exported) or mixedCaps (unexported)
- [ ] Interface names describe behavior (not `I[Name]` prefix)
- [ ] Struct names describe what they are

### Function/Method Names
- [ ] MixedCaps convention followed
- [ ] Constructors follow `New[Type]` pattern
- [ ] Functional options follow `With[Option]` pattern
- [ ] Getters do not use `Get` prefix (Go convention)

### Receiver Names
- [ ] Short (1-2 characters), consistent per type
- [ ] Not `this` or `self`
- [ ] Same receiver name used across all methods of a type

### Variable Names
- [ ] Error variables follow `Err[Condition]` pattern
- [ ] Short names for short scopes
- [ ] Descriptive names for wider scopes
- [ ] No single-letter names outside loops and very short functions

## Phase 3: Godoc Assessment

### Package Documentation
- [ ] Each package has a package comment
- [ ] Package comment is on a single file (conventionally `doc.go` or primary file)
- [ ] Package comment describes what the package does

### Exported Symbols
For each exported type, function, method, constant, variable:
- [ ] Has a godoc comment
- [ ] Comment starts with the symbol name
- [ ] Comment is a complete sentence ending with a period
- [ ] Comment describes what it does, not how

### Example Functions
- [ ] Key types/functions have `Example` functions where helpful
- [ ] Examples compile and run

## Phase 4: Error Handling

### Error Checking
- [ ] All error returns are checked (no `_` for errors)
- [ ] Type assertions use two-value form or are checked
- [ ] Channel operations handle closed channels

### Error Wrapping
- [ ] Errors are wrapped with `fmt.Errorf("context: %w", err)`
- [ ] Error messages describe the operation that failed
- [ ] Error chain is preserved — no `errors.New` that discards original

### Sentinel Errors
- [ ] Expected failure modes have sentinel errors
- [ ] Sentinel errors are package-level `var`s
- [ ] Sentinel errors are documented
- [ ] Implementations return consistent sentinel errors

### Error Types
- [ ] Custom error types implement `error` interface
- [ ] Custom error types implement `Unwrap()` where wrapping
- [ ] No exported error types that should be sentinel errors instead

## Phase 5: Context Usage

### I/O Functions
- [ ] All functions performing I/O accept `context.Context` as first parameter
- [ ] Context is threaded through call chains, not created mid-chain
- [ ] `context.TODO()` is not present in production code

### Context Storage
- [ ] Context is not stored in struct fields
- [ ] Context is passed as a parameter, not retrieved from a stored location

### Cancellation
- [ ] Long-running operations check `ctx.Done()`
- [ ] Context cancellation propagates to child operations

## Phase 6: Pattern Consistency

### Constructor Patterns
- [ ] All constructors follow the same pattern (functional options, config struct, or direct params)
- [ ] No mixed constructor styles within the package

### Error Patterns
- [ ] Same error handling approach throughout
- [ ] Error wrapping format is consistent
- [ ] Sentinel error usage is consistent

### Interface Patterns
- [ ] Interface satisfaction is consistent (all concrete types implement declared interfaces)
- [ ] Mock patterns are consistent

### Code Organization
- [ ] Similar types organized similarly across files
- [ ] Related functionality colocated
- [ ] No arbitrary file boundaries

## Phase 7: Dependency Usage

### Standard Library
- [ ] stdlib used where sufficient
- [ ] No external packages for trivial operations (string manipulation, sorting, etc.)

### Reflect Usage
- [ ] `reflect` package used only where necessary
- [ ] Each usage has clear justification
- [ ] No casual reflection for problems solvable with generics

### Unsafe Usage
- [ ] `unsafe` package is absent
- [ ] If present, heavily justified and documented

### Init Functions
- [ ] No `init()` functions
- [ ] If present, clearly necessary and documented

## Phase 8: Build Verification

### Compilation
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` clean
- [ ] No build warnings

### Module Tidiness
- [ ] `go mod tidy` produces no changes
- [ ] No unused dependencies in go.mod
- [ ] No missing dependencies
