# README Audit

Evaluate an existing README against quality standards and provide actionable recommendations.

## Principles

1. **Essence over problem** — MUST lead with what makes this library unique
2. **Show, don't frame** — Code MUST speak louder than prose
3. **Voice from nature** — Tone MUST match the library's character
4. **Structure as scaffolding** — Sections MUST provide rhythm without feeling scripted

## Execution

1. Read `checklist.md` in this skill directory
2. Work through each phase systematically
3. Compile findings into structured report

## Specifications

### Required Badges

README MUST include exactly 8 badges after the title:

| # | Badge | URL Pattern |
|---|-------|-------------|
| 1 | CI Status | `github.com/.../workflows/CI/badge.svg` |
| 2 | Coverage | `codecov.io/gh/.../graph/badge.svg` |
| 3 | Report Card | `goreportcard.com/badge/...` |
| 4 | CodeQL | `github.com/.../workflows/CodeQL/badge.svg` |
| 5 | Go Reference | `pkg.go.dev/badge/...` |
| 6 | License | `img.shields.io/github/license/...` |
| 7 | Go Version | `img.shields.io/github/go-mod-go-version/...` |
| 8 | Release | `img.shields.io/github/v/release/...` |

### Required Sections

| # | Section | Requirements |
|---|---------|--------------|
| 1 | Header | Title, badges (all 8), tagline |
| 2 | Hook | Library-specific name (NOT generic) |
| 3 | Install | `go get` command |
| 4 | Quick Start | Complete runnable example |
| 5 | Capabilities | Feature table with doc links |
| 6 | Why [Name]? | 3-5 bullet points |
| 7 | Ecosystem | Related projects (if applicable) |
| 8 | Documentation | Learn / Guides / Reference links |
| 9 | Contributing | Link to CONTRIBUTING.md |
| 10 | License | Single line |

### Anti-Patterns

MUST flag if present:
- Template voice (sounds like any library)
- Problem-first framing ("The Problem / The Solution")
- Feature laundry lists
- Duplicate examples between Hook and Quick Start
- API reference that belongs in docs

### Hook Section

Hook section name MUST be library-specific.

PROHIBITED names: "Overview", "Introduction", "How It Works", "Getting Started", "About"

## Output

### Report Structure

```markdown
## Summary
[One paragraph: overall impression and primary recommendation]

## Badge Assessment

| Badge | Present | Working |
|-------|---------|---------|
| CI Status | [✓/✗] | [✓/✗] |
| Coverage | [✓/✗] | [✓/✗] |
| Report Card | [✓/✗] | [✓/✗] |
| CodeQL | [✓/✗] | [✓/✗] |
| Go Reference | [✓/✗] | [✓/✗] |
| License | [✓/✗] | [✓/✗] |
| Go Version | [✓/✗] | [✓/✗] |
| Release | [✓/✗] | [✓/✗] |

## Structure Assessment

| Section | Present | Quality |
|---------|---------|---------|
| Header + Essence | [✓/✗] | [Strong/Adequate/Weak] |
| Hook | [✓/✗] | [Strong/Adequate/Weak] |
| Install | [✓/✗] | [Strong/Adequate/Weak] |
| Quick Start | [✓/✗] | [Strong/Adequate/Weak] |
| Capabilities | [✓/✗] | [Strong/Adequate/Weak] |
| Why [Name]? | [✓/✗] | [Strong/Adequate/Weak] |
| Documentation | [✓/✗] | [Strong/Adequate/Weak] |

## Anti-Patterns Detected
[List any anti-patterns found]

## Strengths
[What the README does well]

## Issues
[Prioritized list with recommendations]

## Suggested Hook Section Name
[If current name is generic]

## Suggested Tagline
[If essence is unclear]
```
