# Mission Audit Checklist

## Phase 1: Mission Document Assessment

### MISSION.md Exists
- [ ] MISSION.md is present in `.claude/`
- [ ] MISSION.md is non-empty

### Required Sections
- [ ] Purpose section exists and contains a clear statement
- [ ] "What This Package Contains" section exists
- [ ] "What This Package Does NOT Contain" section exists
- [ ] Success Criteria section exists
- [ ] Non-Goals section exists

### Section Quality
- [ ] Purpose is a single, focused statement (not a paragraph of qualifications)
- [ ] Contains list is concrete — items are verifiable, not vague
- [ ] Exclusions list is concrete — items are specific things, not "bad stuff"
- [ ] Success criteria are step-by-step and testable
- [ ] Non-goals are specific design decisions, not obvious negatives

## Phase 2: Package Surface Inventory

### Exported API
- [ ] List all exported types
- [ ] List all exported functions
- [ ] List all exported constants and variables
- [ ] List all packages/subpackages

### Capabilities
- [ ] Describe what the package actually does (from code, not docs)
- [ ] Identify the core workflow a user follows
- [ ] Identify secondary/supporting capabilities

## Phase 3: Purpose Alignment

### Core Alignment
- [ ] The package's actual behavior matches the stated purpose
- [ ] The package does not do significantly more than the purpose states
- [ ] The package does not do significantly less than the purpose states

### Feature Classification
For each exported capability:
- [ ] Classify as: aligned, adjacent, contradictory, or orphaned
- [ ] Document reasoning for non-aligned classifications

### Adjacent Features
- [ ] List features that are related but not explicitly covered by mission
- [ ] For each: should the mission expand or the feature be removed?

### Contradictory Features
- [ ] List features that contradict non-goals or exclusions
- [ ] For each: is the feature wrong or is the mission outdated?

### Orphaned Features
- [ ] List features that serve no apparent mission purpose
- [ ] For each: useful addition (update mission) or dead weight (remove)?

## Phase 4: Contains/Excludes Verification

### "What This Package Contains"
For each listed item:
- [ ] Item actually exists in the package
- [ ] Item is functional, not just present
- [ ] Item matches the description

### "What This Package Does NOT Contain"
For each listed exclusion:
- [ ] The excluded thing is genuinely absent
- [ ] No workarounds or partial implementations of excluded items
- [ ] Exclusion is still a reasonable boundary

### Unlisted Items
- [ ] Identify anything the package contains that isn't listed in either section
- [ ] Classify each unlisted item as: belongs in contains, belongs in excludes, or shouldn't exist

## Phase 5: Success Criteria Verification

For each success criterion:

### Achievability
- [ ] The criterion can be achieved with the current implementation
- [ ] Steps described are actually possible to follow
- [ ] No missing functionality blocks the criterion

### Testability
- [ ] The criterion describes a concrete scenario
- [ ] Success/failure is objectively determinable
- [ ] No subjective judgment required ("should feel fast")

### Walk-Through
- [ ] Attempt to follow each criterion step-by-step
- [ ] Document where the walk-through succeeds
- [ ] Document where the walk-through fails or gets stuck

## Phase 6: Non-Goals Verification

For each non-goal:

### Respected
- [ ] The implementation does not pursue this goal
- [ ] No partial or accidental implementation of the non-goal
- [ ] No API surface that implies the non-goal is supported

### Relevance
- [ ] The non-goal is still relevant (not outdated)
- [ ] The non-goal distinguishes this package from alternatives
- [ ] The non-goal reflects a deliberate design decision

## Phase 7: Philosophy Alignment

### Dependency Policy
- [ ] Production dependencies are minimal
- [ ] Each dependency is justified (stdlib and zoobzio checked first)
- [ ] Provider-specific deps are in submodules

### Type Safety
- [ ] Generics used where appropriate
- [ ] No `interface{}` / `any` in public API where type parameters would work

### Boundaries
- [ ] Data transformations happen at identifiable boundaries
- [ ] Boundaries are explicit, not implicit

### Composition
- [ ] Interfaces are small and composable
- [ ] Clear separation between processors, connectors, and values

### Errors
- [ ] Semantic errors defined for expected failure modes
- [ ] Errors carry context

### Context
- [ ] I/O operations accept `context.Context`

## Phase 8: MISSION.md Currency

### Is the Mission Current?
- [ ] Mission reflects the package as it exists today
- [ ] No sections that describe aspirational features as present
- [ ] No sections that describe removed features as present
- [ ] Success criteria are achievable with current code

### Recommendations
- [ ] If mission is outdated: list specific updates needed
- [ ] If implementation has drifted: list features to align or remove
- [ ] If both: prioritize which to fix first
