# Documentation Audit

Evaluate existing package documentation against quality standards and provide actionable recommendations.

## Principles

1. **Clear and direct** — Writing MUST be technical but accessible
2. **Example-driven** — MUST show, then explain
3. **Purpose-fit** — Each doc type MUST serve its distinct need
4. **Cohesive voice** — Docs MUST feel like they belong together

## Execution

1. Read `checklist.md` in this skill directory
2. Work through each phase systematically
3. Compile findings into structured report

## Specifications

### Required Directory Structure

```
docs/
├── 1.learn/
│   ├── 1.overview.md
│   ├── 2.quickstart.md
│   ├── 3.concepts.md
│   └── 4.architecture.md
├── 2.guides/
│   ├── 1.testing.md       (required)
│   ├── 2.troubleshooting.md (required)
│   └── 3.[feature].md     (as needed)
├── 3.integrations/        (optional)
└── 4.reference/
    ├── 1.api.md
    └── 2.types.md
```

### Naming Convention

- Directories: `[number].[name]/`
- Files: `[number].[name].md`
- Numbers control order, names describe content
- Numbers MUST be sequential with no gaps

### Cross-References

All internal links MUST use full numbered paths:
- Correct: `[Quickstart](../1.learn/2.quickstart.md)`
- Wrong: `[Quickstart](../learn/quickstart.md)`

### Required Frontmatter

Every file MUST have:

```yaml
---
title: Article Title
description: One-line description
author: [author]
published: YYYY-MM-DD
updated: YYYY-MM-DD
tags:
  - Tag1
  - Tag2
---
```

### File Requirements

See checklist for specific requirements per file type.

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall documentation health and primary recommendation]

## Coverage Assessment

| File | Exists | Quality | Primary Issue |
|------|--------|---------|---------------|
| overview.md | [✓/✗] | [Strong/Adequate/Weak] | |
| quickstart.md | [✓/✗] | [Strong/Adequate/Weak] | |
| concepts.md | [✓/✗] | [Strong/Adequate/Weak] | |
| architecture.md | [✓/✗] | [Strong/Adequate/Weak] | |
| testing.md | [✓/✗] | [Strong/Adequate/Weak] | |
| troubleshooting.md | [✓/✗] | [Strong/Adequate/Weak] | |
| api.md | [✓/✗] | [Strong/Adequate/Weak] | |
| types.md | [✓/✗] | [Strong/Adequate/Weak] | |

## Strengths
[What the documentation does well]

## Issues
[Prioritized list with recommendations]

## Missing Documentation
[Files that should exist but don't]

## Structural Issues
[Problems with organization, cross-linking, or navigation]
```
