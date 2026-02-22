# Issue Audit Checklist

## Phase 1: Inventory

### Fetch Issues
- [ ] Run `gh issue list --state open --limit 500`
- [ ] Record total count of open issues
- [ ] Note any issues without labels

### Categorize
- [ ] Group issues by type label (feature, bug, docs, infra)
- [ ] Identify issues missing type labels
- [ ] Identify issues in active phases (phase labels present)

## Phase 2: Structure Assessment

For each open issue:

### Required Sections Present
- [ ] Has Objective section (or equivalent clear statement of intent)
- [ ] Has Context section (or equivalent background)
- [ ] Has Acceptance Criteria section
- [ ] Has Scope section (in scope / out of scope)

### Section Quality
- [ ] Objective is a single clear statement
- [ ] Context provides enough background for someone unfamiliar
- [ ] No required section is empty or placeholder text

## Phase 3: Objective Assessment

For each issue:

### Clarity
- [ ] Objective describes an outcome, not a task
- [ ] Objective is specific — not "improve X" or "fix Y" without detail
- [ ] Objective does not contain implementation instructions
- [ ] Objective is achievable within a single issue's scope

### Uniqueness
- [ ] Objective is distinct from all other open issues
- [ ] If related to another issue, relationship is documented

## Phase 4: Acceptance Criteria Assessment

For each issue:

### Format
- [ ] Criteria use checkbox format (`- [ ]`)
- [ ] Each criterion is a single verifiable statement
- [ ] Criteria are ordered logically

### Quality
- [ ] Each criterion describes observable behavior or measurable state
- [ ] No vague criteria ("works correctly", "is performant", "looks right")
- [ ] No criteria that repeat the objective without adding specificity
- [ ] Criteria are collectively sufficient — meeting all criteria means the issue is done

### Completeness
- [ ] Criteria cover the full scope of the objective
- [ ] Edge cases mentioned in context are covered by criteria
- [ ] No implicit requirements that aren't captured in criteria

## Phase 5: Scope Assessment

For each issue:

### In Scope
- [ ] In-scope items are listed explicitly
- [ ] In-scope items align with the objective
- [ ] No in-scope items that expand beyond the objective

### Out of Scope
- [ ] Out-of-scope items are listed explicitly
- [ ] Adjacent work that could be confused as part of this issue is called out
- [ ] Out of scope doesn't contradict acceptance criteria

### Boundaries
- [ ] Clear line between what this issue does and doesn't do
- [ ] No ambiguous items that could be read either way

## Phase 6: Label Assessment

### Type Labels
- [ ] Every issue has exactly one type label
- [ ] Type label accurately reflects the issue content
- [ ] Available types: feature, bug, docs, infra

### Phase Labels
- [ ] Active issues have a phase label
- [ ] Phase label matches the current state of work
- [ ] No issue has multiple phase labels

### Consistency
- [ ] Label usage is consistent across issues
- [ ] No custom labels that duplicate standard ones

## Phase 7: Duplicate and Overlap Analysis

### Duplicate Detection
- [ ] No two issues have the same objective
- [ ] No two issues have substantially overlapping acceptance criteria
- [ ] Issues that address the same area are cross-referenced

### Dependency Detection
- [ ] Issues that depend on other issues are linked
- [ ] No circular dependencies
- [ ] Dependency order is logical

## Phase 8: Cross-Cutting Concerns

### Staleness
- [ ] No issues open for an unreasonable period without activity
- [ ] Stale issues are either still relevant or should be closed

### Consistency
- [ ] Writing style is consistent across issues
- [ ] Level of detail is consistent
- [ ] No issues that are dramatically more or less specified than others

### Actionability
- [ ] Every issue could be picked up by a new team without additional context
- [ ] No issues that require tribal knowledge to understand
