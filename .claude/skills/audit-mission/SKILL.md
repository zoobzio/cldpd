# Mission Audit

Evaluate whether a package's implementation aligns with its stated mission and identify drift.

## Principles

1. **Mission defines purpose** — Everything in the package MUST serve the mission
2. **Scope is bounded** — Non-goals are as important as goals
3. **Success is measurable** — The mission's success criteria MUST be met by the implementation
4. **Drift is caught early** — Feature creep and scope drift are identified before they compound

## Execution

1. Read `checklist.md` in this skill directory
2. Read MISSION.md for the package's stated purpose
3. Read PHILOSOPHY.md for zoobzio-wide principles
4. Work through each phase systematically
5. Compile findings into structured report

## Specifications

### MISSION.md Structure

A well-formed MISSION.md contains:

| Section | Purpose |
|---------|---------|
| Purpose | What the package does — one clear statement |
| What This Package Contains | Concrete list of what's included |
| What This Package Does NOT Contain | Explicit exclusions |
| Success Criteria | Measurable outcomes that define "working" |
| Non-Goals | Things this package deliberately avoids |

### Alignment Assessment

For each exported type, function, and capability:
- Does it serve the stated purpose?
- Is it covered by "What This Package Contains"?
- Does it contradict any non-goal?

### Drift Categories

**Mission-aligned** — Directly serves the stated purpose.

**Mission-adjacent** — Related to the purpose but not explicitly covered. May indicate the mission needs updating or the feature needs removing.

**Mission-contradictory** — Directly contradicts a non-goal or explicit exclusion. This is a defect.

**Mission-orphaned** — Neither serves nor contradicts the mission. Dead weight.

### Success Criteria Verification

Each success criterion in MISSION.md MUST be:
- Actually achievable with the current implementation
- Testable by following the described steps
- Not blocked by missing functionality

### Philosophy Alignment

The package MUST align with PHILOSOPHY.md:
- Dependency policy (minimal production deps, isolated providers)
- Type safety (generics over interface{})
- Boundaries (explicit data transformation)
- Composition (interfaces, processors, connectors)
- Errors (semantic, contextual)
- Context (all I/O accepts context.Context)

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall mission alignment and primary recommendation]

## Mission Statement
[Reproduce the package's stated purpose from MISSION.md]

## Alignment Matrix

| Category | Status | Primary Issue |
|----------|--------|---------------|
| Purpose Clarity | [✓/~/✗] | |
| Contains Accuracy | [✓/~/✗] | |
| Exclusions Accuracy | [✓/~/✗] | |
| Success Criteria | [✓/~/✗] | |
| Non-Goals Respected | [✓/~/✗] | |
| Philosophy Alignment | [✓/~/✗] | |

## Feature Alignment

| Feature/Capability | Classification | Notes |
|--------------------|---------------|-------|
| [feature] | [aligned/adjacent/contradictory/orphaned] | |

## Success Criteria Verification

| Criterion | Achievable | Blocked By |
|-----------|-----------|------------|
| [criterion] | [✓/✗] | [blocker if any] |

## Mission Drift
[Features or capabilities that have drifted from the stated mission]

## Mission Gaps
[Stated mission items that the implementation doesn't fulfill]

## MISSION.md Quality
[Assessment of the mission document itself — is it clear, complete, current?]

## Recommendations
[Prioritized list: fix the implementation, update the mission, or both]
```

Status legend: ✓ Compliant, ~ Partial, ✗ Missing/Non-compliant
