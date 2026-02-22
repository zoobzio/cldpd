# Release

Merge develop into main to trigger the auto-version and release workflow.

## Prerequisites

- Gitflow setup complete (main + develop branches)
- Auto-version workflow configured
- All CI checks passing on develop

## Flow

```
1. Verify develop is ready
2. Create PR: develop → main
3. Preview version (auto-commented)
4. Merge PR
5. Auto-version tags
6. Release workflow runs
```

## Execution

1. Read `checklist.md` in this skill directory
2. Verify develop is clean and CI passes
3. Create PR from develop to main
4. Confirm version preview
5. Merge (or instruct user to merge)

## Specifications

### Pre-Release Requirements

Before creating release PR, ALL must be true:
- Currently on `develop` branch
- Working directory is clean
- develop is up to date with origin
- CI is passing on develop
- No WIP or incomplete features in commits since last release

### Release PR Format

Title: `Release: merge develop into main`

Body:
```markdown
## Release

Merging develop into main to trigger release.

### Changes
[git log main..develop --oneline]

### Version
Version will be auto-calculated from conventional commits.

---
*This PR will trigger auto-versioning and release workflows when merged.*
```

### Merge Method

MUST use squash merge for clean history.

## Prohibitions

DO NOT:
- Create release PR if CI is failing on develop
- Create release PR with uncommitted changes
- Merge without confirming version preview
- Use merge commit (use squash only)

## Output

A merged PR that triggers:
- Auto-version workflow (calculates and pushes tag)
- Release workflow (creates GitHub release)
