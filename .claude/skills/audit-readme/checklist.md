# README Audit Checklist

## Phase 1: Context Gathering

### Package Understanding
- [ ] Read the existing README completely
- [ ] Examine the package's actual public API
- [ ] Review test files for usage patterns
- [ ] Check if examples/ directory exists and aligns with README examples

### Verification
- [ ] Do code examples compile/run?
- [ ] Are documented features actually present in the codebase?
- [ ] Do links resolve correctly?
- [ ] Is the installation command accurate?

## Phase 2: Anti-Pattern Detection

### Template Voice
- [ ] Does this README sound like it could describe any library?
- [ ] Are section names generic ("Overview", "Introduction")?
- [ ] Does the prose feel like fill-in-the-blank?

### Problem-First Framing
- [ ] Does it open with complaints or pain points?
- [ ] How many paragraphs before showing code?
- [ ] Is there a "The Problem / The Solution" structure?

### Feature Laundry Lists
- [ ] Are there long bullet-point lists that belong in docs?
- [ ] Is the README trying to be comprehensive instead of compelling?

### Duplicate Examples
- [ ] Do multiple code blocks cover the same ground?
- [ ] Is the hook distinct from quick start?

### Misplaced Content
- [ ] Is there API reference that belongs in docs?
- [ ] Are there tutorials that should be separate guides?

## Phase 3: Badge Verification

Check for standard badge set after title:

| Badge | Present | Working |
|-------|---------|---------|
| CI Status | | |
| codecov | | |
| Go Report Card | | |
| CodeQL | | |
| Go Reference | | |
| License | | |
| Go Version | | |
| Release | | |

- [ ] All badges present
- [ ] Badge URLs use correct package name
- [ ] Badges render correctly (not broken images)
- [ ] Badge links resolve to correct destinations

## Phase 4: Structure Assessment

Evaluate presence and quality of each section:

| Section | Present | Quality | Notes |
|---------|---------|---------|-------|
| Badges | | | |
| Header + Essence | | | |
| Hook (library-specific) | | | |
| Install | | | |
| Quick Start | | | |
| Capabilities | | | |
| Why [Name]? | | | |
| Ecosystem | | | |
| Documentation Links | | | |
| Contributing | | | |
| License | | | |

Quality ratings:
- **Strong**: Captures the spirit, draws readers in
- **Adequate**: Functional but generic
- **Weak**: Missing, misleading, or template-feeling
- **N/A**: Not applicable to this package

## Phase 5: Essence Evaluation

- [ ] After reading this README, can I explain what makes this library unique in one sentence?
- [ ] Does the hook section name tell me something about the library's core insight?
- [ ] Would I know why to choose this over alternatives?
