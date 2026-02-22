# Documentation Audit Checklist

## Phase 1: Inventory

### File Presence
- [ ] Does `docs/` directory exist?
- [ ] Check for `1.learn/1.overview.md`
- [ ] Check for `1.learn/2.quickstart.md`
- [ ] Check for `1.learn/3.concepts.md`
- [ ] Check for `1.learn/4.architecture.md`
- [ ] Check for `2.guides/1.testing.md`
- [ ] Check for `2.guides/2.troubleshooting.md`
- [ ] Check for `4.reference/1.api.md`
- [ ] Check for `4.reference/2.types.md`
- [ ] List any additional guides in `2.guides/` (should be numbered: `3.[name].md`, etc.)
- [ ] Check if `3.integrations/` exists and files are numbered

### Numbering Convention
- [ ] All directories use numeric prefix: `1.learn/`, `2.guides/`, etc.
- [ ] All files use numeric prefix: `1.overview.md`, `2.quickstart.md`, etc.
- [ ] Numbers are sequential within each directory
- [ ] No gaps in numbering sequence
- [ ] All cross-reference links use full numbered paths (e.g., `../2.guides/1.testing.md`)

### Frontmatter Verification
- [ ] Does each file have required frontmatter (title, description, author, published, updated, tags)?
- [ ] Are dates in correct format (YYYY-MM-DD)?
- [ ] Are descriptions meaningful (not placeholder text)?

## Phase 2: Overview Assessment

- [ ] Does it answer "What is this?" in one sentence?
- [ ] Does it explain what motivated the package?
- [ ] Does it list what the package provides?
- [ ] Does it explain what you can build with it?
- [ ] Does it link to next steps?
- [ ] Is it free of installation instructions? (those belong in Quickstart)
- [ ] Is it free of API details? (those belong in Reference)
- [ ] Can a reader decide in 60 seconds if this package suits their needs?

## Phase 3: Quickstart Assessment

- [ ] Does it include installation instructions?
- [ ] Is there a complete, copy-paste-runnable basic usage example?
- [ ] Does it introduce foundational topics with brief explanations?
- [ ] Does it link to guides for deeper coverage?
- [ ] Does it link to concepts, architecture, reference?
- [ ] Is it free of exhaustive API coverage?
- [ ] Is it free of edge cases and advanced configuration?

## Phase 4: Concepts Assessment

- [ ] Does it define core abstractions?
- [ ] Does it explain why abstractions are shaped the way they are?
- [ ] Does it show how abstractions relate to each other?
- [ ] Does it cross-link to Reference for API elements?
- [ ] Is it free of complete type definitions? (those belong in Reference)
- [ ] Is it free of implementation details? (those belong in Architecture)
- [ ] Is it free of step-by-step tutorials? (those belong in Quickstart/Guides)

## Phase 5: Architecture Assessment

- [ ] Does it explain component structure and interactions?
- [ ] Does it describe data flow and key algorithms?
- [ ] Does it include design rationale as Q&A?
- [ ] Does it cover performance characteristics?
- [ ] Is it free of usage examples? (those belong in Quickstart/Guides)
- [ ] Does it link to Concepts rather than re-explaining them?
- [ ] Does it link to Reference rather than duplicating API docs?

## Phase 6: Guides Assessment

### Required Guides
- [ ] Does testing guide explain how to test code using this package?
- [ ] Does troubleshooting guide cover common errors and debugging?

### All Guides
- [ ] Does each guide have clear prerequisites stated?
- [ ] Are examples complete and runnable?
- [ ] Do guides link to related guides and reference?
- [ ] Is content appropriately scoped (not too broad, not too narrow)?

## Phase 7: Reference Assessment

### API Reference
- [ ] Does each function have: signature, description, panic/error conditions, example?
- [ ] Are functions organized by category?
- [ ] Does it link to Types Reference for return types and parameters?
- [ ] Is it complete (all public functions documented)?

### Types Reference
- [ ] Does each type have: definition, field table, usage notes?
- [ ] Are types organized logically?
- [ ] Is it complete (all public types documented)?

## Phase 8: Cross-Cutting Concerns

### Accuracy
- [ ] Do code examples compile/run?
- [ ] Do internal links use full numbered paths? (GitHub compatibility)
- [ ] Do links resolve correctly?
- [ ] Does documentation match actual package behavior?
- [ ] Are version requirements accurate?

### Consistency
- [ ] Is tone consistent across all files?
- [ ] Is terminology consistent?
- [ ] Do files follow the same structural patterns?
- [ ] All files follow `[number].[name].md` naming convention
- [ ] All directories follow `[number].[name]/` naming convention

### Navigation
- [ ] Can a reader easily find what they need?
- [ ] Are "Next Steps" sections present and helpful?
- [ ] Is cross-linking appropriate (not excessive, not missing)?
