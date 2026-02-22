# Release Checklist

## Phase 1: Pre-Release Verification

### Branch State
- [ ] Currently on `develop` branch
- [ ] Working directory is clean (`git status`)
- [ ] develop is up to date with origin (`git pull origin develop`)

### CI Status
- [ ] CI passing on develop: `gh run list --branch develop --limit 1`
- [ ] No failing checks

### Changes Since Last Release
- [ ] Review commits: `git log main..develop --oneline`
- [ ] Confirm changes are release-worthy
- [ ] No WIP or incomplete features

### Version Preview
Check what version will be created:
```bash
# If svu is installed locally
svu next

# Or check recent tags
git describe --tags --abbrev=0
```

- [ ] Expected version bump makes sense (major/minor/patch)

## Phase 2: Create Release PR

### Option A: Via gh CLI

```bash
gh pr create \
  --base main \
  --head develop \
  --title "Release: merge develop into main" \
  --body "$(cat <<'EOF'
## Release

Merging develop into main to trigger release.

### Changes
$(git log main..develop --oneline)

### Version
Version will be auto-calculated from conventional commits.

---
*This PR will trigger auto-versioning and release workflows when merged.*
EOF
)"
```

### Option B: Manual Steps

1. Go to GitHub repository
2. Click "Compare & pull request" or "New pull request"
3. Set base: `main`, compare: `develop`
4. Title: "Release: merge develop into main"
5. Create PR

- [ ] PR created from develop to main
- [ ] PR URL noted

## Phase 3: Verify PR

### Automated Checks
- [ ] CI runs on the PR
- [ ] Version preview comment appears (shows next version)
- [ ] No merge conflicts

### Review Version Preview
The version-preview workflow should comment:
```
## 📋 Version Preview
**Current:** v1.2.3
**Next:** v1.3.0
```

- [ ] Version preview is correct
- [ ] Version bump matches expectations

### Final Check
- [ ] All status checks pass
- [ ] Code owner approval (if required by self, approve or get approval)

## Phase 4: Merge

### Option A: Via gh CLI

```bash
gh pr merge --squash --auto
```

Or if ready to merge immediately:
```bash
gh pr merge --squash
```

### Option B: Via GitHub UI

1. Click "Squash and merge"
2. Confirm merge

- [ ] PR merged

## Phase 5: Verify Release

### Tag Creation
Wait for auto-version workflow (~30 seconds):

```bash
# Refresh tags
git fetch --tags

# Check latest tag
git describe --tags --abbrev=0
```

- [ ] New tag created

### Release Workflow
Check release workflow status:

```bash
gh run list --workflow=release.yml --limit 1
```

Or watch in real-time:
```bash
gh run watch
```

- [ ] Release workflow triggered
- [ ] Release workflow completed

### GitHub Release
```bash
gh release view $(git describe --tags --abbrev=0)
```

Or open in browser:
```bash
gh release view $(git describe --tags --abbrev=0) --web
```

- [ ] GitHub release created
- [ ] Changelog generated
- [ ] Assets attached (if applicable)

## Phase 6: Post-Release

### Sync Branches
Ensure develop has the merge commit:

```bash
git checkout develop
git pull origin develop
git checkout main
git pull origin main
```

- [ ] Both branches up to date locally

### Announce (Optional)
- [ ] Post release notes to relevant channels
- [ ] Update any external documentation

### Summary

```markdown
## Release Complete

**Version:** [version]
**Release URL:** [url]

### Changes Included
- [change 1]
- [change 2]

### Next Steps
- Continue development on feature branches
- PRs should target `develop`
```

## Troubleshooting

### Auto-version didn't run
- Check if merge was a squash merge (required)
- Check workflow permissions
- Check if commit message contains "Merge pull request"

### Version not bumped
- Ensure commits follow conventional commit format
- Check svu is calculating correctly
- May need manual tag if no feat/fix commits

### Release workflow failed
- Check release workflow logs
- Verify GITHUB_TOKEN permissions
- Check goreleaser configuration
