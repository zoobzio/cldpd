# Architecture Audit

Evaluate existing package architecture against zoobzio design principles and provide actionable recommendations.

## Principles

1. **Composition over inheritance** — Interfaces MUST be small, nestable, and composable
2. **Boundaries are explicit** — Data transformations MUST happen at well-defined boundaries
3. **Dependencies are deliberate** — Production dependencies MUST be justified
4. **Type safety is non-negotiable** — Generics over `interface{}`, compile-time over runtime

## Execution

1. Read `checklist.md` in this skill directory
2. Work through each phase systematically
3. Compile findings into structured report

## Specifications

### Interface Design

Interfaces MUST:
- Express one abstraction level per interface
- Be composable — larger interfaces embed smaller ones
- Be implementable by both real and test types
- Accept `context.Context` for I/O operations

Interfaces MUST NOT:
- Be wider than their consumers need
- Contain methods from multiple abstraction levels
- Require concrete type knowledge to implement

### Composition Model

- Immutable processors: stateless, transform data, easily testable
- Mutable connectors: manage state, handle I/O, own lifecycle
- Values: simple, comparable, no hidden state

### Boundary Design

Each boundary MUST:
- Be identifiable (receiving input, loading, storing, sending output)
- Transform data explicitly
- Validate at ingress
- Redact at egress where appropriate

### Dependency Policy

Production dependencies:
- stdlib first
- zoobzio packages second
- external packages as last resort with clear justification

Provider-specific dependencies MUST live in submodules, never in root.

### Error Design

- Errors carry context: what failed, where, why
- Semantic errors (`ErrNotFound`, `ErrDuplicate`) are consistent across implementations
- Error wrapping preserves the chain
- Sentinel errors are defined at the package level

### Observability

- Components have identity
- Signals are emittable without requiring external infrastructure
- Correlation is possible across component boundaries

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall architectural health and primary recommendation]

## Coverage Matrix

| Category | Status | Primary Issue |
|----------|--------|---------------|
| Interface Design | [✓/~/✗] | |
| Composition Model | [✓/~/✗] | |
| Boundary Design | [✓/~/✗] | |
| Dependency Policy | [✓/~/✗] | |
| Error Design | [✓/~/✗] | |
| Observability | [✓/~/✗] | |
| Type Safety | [✓/~/✗] | |

## Interface Analysis
[Assessment of interface design — width, composability, abstraction levels]

## Composition Analysis
[Assessment of processor/connector/value separation]

## Boundary Analysis
[Assessment of boundary explicitness and data flow]

## Dependency Analysis
[Assessment of dependency justification and isolation]

## Structural Issues
[Problems with overall package organization]

## Quick Wins
[Low-effort fixes]
```

Status legend: ✓ Compliant, ~ Partial, ✗ Missing/Non-compliant
