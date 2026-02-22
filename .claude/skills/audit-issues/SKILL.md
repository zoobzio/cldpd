# Issue Audit

Evaluate existing GitHub issues against quality standards and provide actionable recommendations.

## Principles

1. **Issues are actionable** — Every issue MUST have enough information to begin work
2. **Scope is explicit** — What's in and what's out MUST be clear
3. **Acceptance is measurable** — Criteria MUST be verifiable, not subjective
4. **No duplicates** — Each issue MUST represent a distinct unit of work

## Execution

1. Read `checklist.md` in this skill directory
2. Fetch open issues via `gh issue list`
3. Work through each phase systematically
4. Compile findings into structured report

## Specifications

### Required Issue Sections

Every issue MUST have:

| Section | Purpose |
|---------|---------|
| Objective | One clear statement of what needs to be accomplished |
| Context | Why this matters, background for architecture |
| Acceptance Criteria | Specific, verifiable checklist items |
| Scope | What's in scope and what's explicitly out |

### Objective Quality

An objective MUST:
- Be a single statement
- Describe an outcome, not a task
- Be specific enough to evaluate completion
- Not contain implementation details

### Acceptance Criteria Quality

Each criterion MUST:
- Be independently verifiable
- Use checkbox format (`- [ ]`)
- Describe observable behavior or state
- Not be vague ("works correctly", "is fast", "looks good")

### Scope Quality

Scope MUST:
- Define "in scope" items explicitly
- Define "out of scope" items explicitly
- Not contradict acceptance criteria
- Not leave ambiguous boundaries

### Labels

Issues MUST have:
- A type label (`feature`, `bug`, `docs`, `infra`)
- A phase label when in active work

### Duplicates and Overlap

Issues MUST NOT:
- Duplicate another open issue
- Partially overlap without cross-reference
- Depend on another issue without explicit linkage

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall issue backlog health and primary recommendation]

## Coverage Matrix

| Category | Status | Primary Issue |
|----------|--------|---------------|
| Issue Structure | [✓/~/✗] | |
| Objective Quality | [✓/~/✗] | |
| Acceptance Criteria | [✓/~/✗] | |
| Scope Definition | [✓/~/✗] | |
| Labeling | [✓/~/✗] | |
| Duplicates/Overlap | [✓/~/✗] | |

## Issue-by-Issue Assessment

| # | Title | Structure | Objective | Criteria | Scope | Labels |
|---|-------|-----------|-----------|----------|-------|--------|
| [#] | [title] | [✓/~/✗] | [✓/~/✗] | [✓/~/✗] | [✓/~/✗] | [✓/~/✗] |

## Common Problems
[Patterns of issues seen across multiple issues]

## Duplicate/Overlap Groups
[Issues that duplicate or overlap each other]

## Quick Wins
[Low-effort fixes — missing labels, minor clarifications]
```

Status legend: ✓ Compliant, ~ Partial, ✗ Missing/Non-compliant
