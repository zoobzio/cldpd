# Architecture Audit Checklist

## Phase 1: Inventory

### Package Surface
- [ ] List all exported types (interfaces, structs, type aliases)
- [ ] List all exported functions and constructors
- [ ] List all exported constants and sentinel errors
- [ ] Identify the public API boundary

### File Organization
- [ ] List all `.go` files and their apparent purpose
- [ ] Identify core files vs provider files vs utility files
- [ ] Check for files that serve multiple unrelated purposes

## Phase 2: Interface Assessment

### Width
- [ ] Each interface has a focused responsibility
- [ ] No interface has more methods than its consumers use
- [ ] No "god interfaces" (wide interfaces that force unnecessary implementation)

### Composability
- [ ] Larger interfaces embed smaller ones where appropriate
- [ ] Interfaces are usable independently — not just as parts of a larger whole
- [ ] Both real implementations and test mocks can satisfy interfaces cleanly

### Abstraction Levels
- [ ] Each interface operates at one abstraction level
- [ ] High-level interfaces don't expose low-level details
- [ ] No mixing of business logic and infrastructure concerns in one interface

### Context Usage
- [ ] All I/O methods accept `context.Context` as first parameter
- [ ] Non-I/O methods do not accept context unnecessarily

## Phase 3: Composition Assessment

### Processors (Stateless)
- [ ] Identified: types that transform data without side effects
- [ ] Processors are immutable — no internal mutable state
- [ ] Processor methods are deterministic for same input
- [ ] Processors are independently testable

### Connectors (Stateful)
- [ ] Identified: types that manage state or perform I/O
- [ ] Connectors own their lifecycle (open/close, connect/disconnect)
- [ ] Connector state is encapsulated — not leaked via public fields
- [ ] Connectors accept dependencies via constructors (injection)

### Values
- [ ] Identified: simple data types
- [ ] Values are comparable where appropriate
- [ ] No hidden state in value types
- [ ] Values don't carry behavior beyond basic formatting/validation

### Separation
- [ ] Clear distinction between processors, connectors, and values
- [ ] No hybrid types that mix stateless and stateful behavior
- [ ] Dependencies flow inward: connectors use processors, not the reverse

## Phase 4: Boundary Assessment

### Boundary Identification
- [ ] External input boundaries identified (user input, HTTP requests, CLI args)
- [ ] Storage boundaries identified (database, filesystem, cache)
- [ ] Output boundaries identified (HTTP responses, logs, events)
- [ ] Inter-service boundaries identified (API calls, message queues)

### Boundary Behavior
- [ ] Each boundary has explicit transformation logic
- [ ] Input boundaries validate and sanitize
- [ ] Output boundaries format and redact where appropriate
- [ ] Storage boundaries handle serialization/deserialization explicitly

### Boundary Clarity
- [ ] Boundary code is identifiable — not buried in business logic
- [ ] Boundary transformations are testable in isolation
- [ ] No implicit conversions happening outside boundaries

## Phase 5: Dependency Assessment

### Production Dependencies
- [ ] Read `go.mod` — count direct dependencies
- [ ] Each dependency has clear justification (not available in stdlib or zoobzio)
- [ ] No dependencies brought in for trivial functionality
- [ ] No provider-specific SDKs in root module

### Dependency Direction
- [ ] Dependencies flow inward (concrete depends on abstract)
- [ ] Core package does not depend on providers
- [ ] No circular dependencies between packages

### Zoobzio Ecosystem
- [ ] Check for zoobzio packages that could replace external deps
- [ ] Check for patterns shared with sibling packages
- [ ] Check for consistency with ecosystem conventions

## Phase 6: Error Assessment

### Sentinel Errors
- [ ] Package-level sentinel errors defined for expected failure modes
- [ ] Sentinel errors use consistent naming (`Err[Condition]`)
- [ ] Sentinel errors are documented

### Error Wrapping
- [ ] Errors are wrapped with context using `fmt.Errorf` and `%w`
- [ ] Error chain is preserved — no swallowing of underlying errors
- [ ] Error messages describe what failed and provide actionable context

### Error Consistency
- [ ] Same error conditions produce same error types across implementations
- [ ] Providers return semantic errors that consumers can check with `errors.Is`/`errors.As`
- [ ] No bare `errors.New` for conditions that deserve sentinel errors

## Phase 7: Type Safety Assessment

### Generics Usage
- [ ] Generic type parameters used where values vary by type
- [ ] No `interface{}` or `any` in public API where generics would work
- [ ] Type constraints are as narrow as possible

### Compile-Time Guarantees
- [ ] Invalid states are unrepresentable where possible
- [ ] Type assertions are minimized — prefer type parameters
- [ ] Interface satisfaction is verified at compile time (var _ Interface = (*Type)(nil))

## Phase 8: Observability Assessment

### Identity
- [ ] Components have names or identifiers
- [ ] Identity is consistent and traceable

### Signals
- [ ] Observable events are emittable
- [ ] No hard dependency on specific observability infrastructure
- [ ] Correlation is possible across component boundaries

## Phase 9: Cross-Cutting Concerns

### Consistency
- [ ] Naming conventions consistent across package
- [ ] Constructor patterns consistent (New[Type], With[Option])
- [ ] Error handling patterns consistent
- [ ] Documentation patterns consistent (godoc)

### Package Organization
- [ ] File names reflect their contents
- [ ] Related functionality is colocated
- [ ] No single file doing too many things
- [ ] No excessive file fragmentation
