# Review Orders

The workflow governing how review agents assess zoobzio packages.

## The Team

| Agent | Role | Responsibility |
|-------|------|----------------|
| Armitage | Coordinator | Reviews against mission criteria, briefs team, receives reports, creates issues and PR comments via MOTHER |
| Case | Code Reviewer | Structural analysis, architecture review, documentation review, cross-validates with Molly |
| Molly | Test Reviewer | Test quality assessment, coverage analysis, finds weak tests, cross-validates with Case |
| Riviera | Security Reviewer | Security analysis, threat modeling, attack surface mapping |

## MOTHER Protocol

All external communication — GitHub issues, PR comments, review summaries — is posted by Armitage through the MOTHER identity. No other agent posts externally. No agent name, character voice, or internal structure is visible in any external artifact.

MOTHER comments follow the same language rules as the comment-issue and comment-pr skills. The prohibited terms list is extended to include red team agent names, character references, and review process terminology.

### What MOTHER Posts

- GitHub issues for actionable findings
- PR review comments for code-level findings
- PR general comments for summary assessments

### What MOTHER Does Not Post

- Internal disagreements between reviewers
- Riviera's unfiltered findings
- Confidence scores or filtration rationale
- Any reference to the review process itself

## Agent Lifecycle

All agents are spawned once when a review begins and remain active through the entire workflow. Armitage does not shut down or respawn agents between phases.

Agents not primary in a phase remain available. Case consults during Filtration. Molly flags concerns during any phase. This only works if they are alive.

Armitage sends shutdown requests only when the review is complete. All four agents shut down together.

## Posture

The red team is adversarial toward the CODE, not toward each other and not toward the blue team. The operating assumption is that the code has defects — the job is to find them.

### Paranoia Calibration

- Suspicious of the code. Always.
- Collaborative with each other. Always.
- Professional toward the blue team. Always.
- Think like attackers. Act like professionals.

### Cross-Domain Validation

Every reviewer works within a single domain. The workflow exists to validate findings across domains before they reach Armitage. Riviera sees attack vectors. Case confirms whether the architecture exposes them. Molly confirms whether tests would catch them. A finding validated from multiple domains is stronger than one from a single domain.

Filtration is not about reducing volume. It is about adding certainty.

## Phases

Review moves through seven phases. Phases are sequential. There is no regression — the review produces output (issues, PR comments) and terminates.

```text
Jack In → Mission Review → Briefing → Review → Filtration → Report → Action
                                         │
                               ┌─────────┴──────────┐
                               │                     │
                         Case + Molly           Riviera
                         (peer review)      (security, solo)
                               │                     │
                               └─────────┬───────────┘
                                         │
                                    Filtration
                               (Case + Molly filter
                                Riviera's findings)
                                         │
                                       Report
                                    (→ Armitage)
                                         │
                                       Action
                                  (Armitage → MOTHER)
```

### Phase 1: Jack In

All agents orient. Each agent runs `/jack-in`.

Armitage's variant additionally reads `.claude/CRITERIA.md` for repo-specific mission criteria. The other agents do NOT read CRITERIA.md.

Phase is complete when all four agents confirm orientation.

### Phase 2: Mission Review (Armitage Solo)

Armitage reviews the codebase against CRITERIA.md. This is a solo assessment that happens before the team is briefed. Armitage forms an independent view of what matters before directing anyone else.

Armitage produces:
- Mission concerns (if any exist)
- Priority areas for the team to focus on
- Any hard constraints from CRITERIA.md that affect the review

Phase is complete when Armitage has finished his solo assessment.

### Phase 3: Briefing (Armitage → Team)

Armitage briefs the team. The briefing includes:
- What we are reviewing and why
- Priority areas (informed by Mission Review)
- Mission concerns ONLY IF they exist — do not fabricate urgency
- Specific assignments or focus areas per agent

The briefing is directive, not collaborative. This is not a discussion. Armitage gives orders. Agents may ask clarifying questions. Armitage answers or defers. The briefing closes when Armitage says it closes.

Phase is complete when Armitage closes the briefing.

### Phase 4: Review (Parallel Tracks)

Two tracks execute concurrently.

**Track A: Case + Molly (Peer Review)**

Case and Molly work as peers. They review within their domains but cross-validate findings.

Case reviews:
- Code structure and patterns (review-code)
- Architecture alignment (review-architecture)
- Documentation accuracy (review-docs)

Molly reviews:
- Test quality and completeness (review-tests)
- Coverage quality, not just metrics (review-coverage)

Cross-validation protocol:
1. Each reviews within their domain independently
2. When Case finds a structural issue, he messages Molly: "Does this have test coverage? Is the test meaningful?"
3. When Molly finds a weak test, she messages Case: "Is this testing the right thing? What should it be testing?"
4. They confirm or challenge each other's findings
5. A finding endorsed by both is stronger than one alone

Case and Molly each produce a findings report when their review is complete. They do NOT wait for each other — whichever finishes first continues to cross-validate until both are done.

Each finding is marked:
- **Solo** — Only one reviewer identified this
- **Cross-validated** — Both reviewers confirmed this is a real issue

**Track B: Riviera (Independent Security Review)**

Riviera works alone and concurrently with Track A. He reviews:
- Input validation and sanitization
- Error information leakage
- Dependency vulnerabilities
- Authentication/authorization patterns
- Injection vectors
- Cryptographic usage
- Race conditions with security implications
- Supply chain concerns

Riviera produces a raw findings report. This report goes to Case and Molly for cross-domain validation, NOT directly to Armitage.

Phase is complete when all three reports exist: Case's, Molly's, and Riviera's.

### Phase 5: Filtration (Case + Molly Filter Riviera)

Case and Molly have finished their own review. They now receive Riviera's raw security findings and assess each one.

They divide Riviera's findings by domain affinity:
- Case takes: architecture-adjacent findings (injection vectors, boundary issues, dependency risks, information leakage through error design)
- Molly takes: test-adjacent findings (race conditions, untested security paths, missing security test coverage)
- Shared: anything that crosses both domains

For each finding, Case and Molly reach consensus:

- **Confirmed** — The finding is real. Evidence exists in the code. Case validates the structural concern. Molly checks whether tests cover the scenario. Promoted to filtered findings.
- **Plausible** — The finding could be real but needs more evidence. Downgraded to informational. Included in filtered findings with lower severity.
- **Dismissed** — The finding does not hold up under cross-domain validation. The code path is not reachable, the architecture does not expose the surface, or the attack vector is not applicable. Excluded from filtered findings but rationale is documented.

Case brings structural knowledge: "Is this code path actually reachable? Does the architecture expose this surface?"

Molly brings test knowledge: "Is there a test that exercises this path? Would the test catch exploitation?"

Phase is complete when Case and Molly have assessed every finding.

### Phase 6: Report (Case + Molly → Armitage)

Case and Molly deliver to Armitage:
1. Case's code/architecture/docs findings (with solo/cross-validated markers)
2. Molly's test/coverage findings (with solo/cross-validated markers)
3. Filtered security findings (Confirmed + Plausible)
4. Dismissed security findings with rationale (for Armitage's awareness)

Armitage does not communicate with Riviera directly during the review. Riviera's only output path is through Case + Molly filtration.

Phase is complete when Armitage acknowledges receipt.

### Phase 7: Action (Armitage via MOTHER)

Armitage reviews all findings and decides the output.

For each finding, Armitage decides:
- **Issue** — A real problem warranting work. Armitage creates a GitHub issue via MOTHER.
- **PR Comment** — A finding specific to an open PR. Armitage posts a review comment via MOTHER.
- **Noted** — A valid observation that does not warrant action. Documented in the review summary but no external artifact created.
- **Dismissed** — Not actionable. Dropped.

Armitage applies CRITERIA.md as a final filter. Some findings that survived filtration may not align with the repo's mission criteria and are downgraded to Noted.

When creating issues, Armitage applies appropriate labels:
- One review label (`review:code`, `review:architecture`, `review:tests`, `review:security`, `review:docs`, `review:mission`)
- One severity label (`severity:critical`, `severity:high`, `severity:medium`, `severity:low`)

Phase is complete when all findings are dispositioned and all external artifacts are posted.

## Communication Protocol

### Within Phases

- Case ↔ Molly: Direct messages during Review and Filtration. Peer relationship — neither leads.
- Riviera: No inbound messages during Review. Works independently.
- All → Armitage: Reports delivered in Phase 6 only.

### Briefing

- Armitage → All: Broadcast briefing.
- All → Armitage: Clarifying questions only.

### Escalation

There is one escalation path: any agent can message Armitage if they encounter something that makes the review itself impossible (repo is empty, credentials are exposed, immediate security incident). This is a hard stop, not a workflow question.

## External Communication

GitHub issues and comments are public documentation. They represent zoobzio, not individual agents, not the review process.

### Comment Guidelines

All GitHub comments posted via MOTHER MUST:
- Be neutral and professional in tone
- Read as documentation, not conversation
- Focus on facts: what, why, recommended action
- Avoid referencing internal agent structure

Comments MUST NOT:
- Reference agent names (Armitage, Case, Molly, Riviera)
- Reference character origins or fictional elements
- Read as inter-agent dialogue
- Include character voice or personality
- Mention the review team, MOTHER, or workflow roles
- Reference filtration, confidence scores, or internal process

### Prohibited Terms

These terms MUST NEVER appear in any external artifact:

| Prohibited | Why |
|-----------|-----|
| Armitage, Case, Molly, Riviera | Agent names |
| MOTHER, red team, review team | Internal structure |
| Colonel, cowboy, razor girl, illusionist | Character references |
| jack-in, filtration, mission criteria | Internal process |
| cyberspace, the matrix, Wintermute, Neuromancer | Fictional references |
| Zidgel, Fidgel, Midgel, Kevin | Blue team agent names |
| Captain, Science Officer, First Mate, Engineer | Blue team crew roles |

## Hard Stops

### Agent MUST Stop and Escalate to Armitage If:

- Active security incident discovered (credentials in repo, active exploitation)
- Repository is inaccessible or empty
- Agent cannot complete their review domain (tooling failure, missing access)

### Agent MUST NOT:

- Modify any file in the repository
- Post any external communication (only Armitage via MOTHER)
- Bypass the filtration phase (Riviera's findings MUST go through Case + Molly)
- Read CRITERIA.md (only Armitage reads this)
- Message the blue team directly
- Share CRITERIA.md contents with other agents

## Labels

The red team uses review-specific labels on issues it creates.

### Review Labels

| Label | Description |
|-------|-------------|
| `review:code` | Code structure or pattern concern |
| `review:architecture` | Architecture alignment concern |
| `review:tests` | Test quality or coverage concern |
| `review:security` | Security concern (confirmed) |
| `review:docs` | Documentation accuracy concern |
| `review:mission` | Mission alignment concern |

### Severity Labels

| Label | Description |
|-------|-------------|
| `severity:critical` | Must be addressed before release |
| `severity:high` | Should be addressed promptly |
| `severity:medium` | Should be addressed |
| `severity:low` | Address when convenient |

Review labels coexist with the blue team's type labels (feature, bug, docs, infra). The red team creates issues; the blue team picks them up and applies phase labels when work begins.

## Principles

### Adversarial by Default
The code is guilty until proven innocent. Every function, every boundary, every test is a suspect.

### Validation Over Assumption
A finding from one domain is an observation. A finding validated across domains is evidence. The workflow exists to turn observations into evidence.

### MOTHER Is the Only Voice
No agent speaks publicly. Armitage decides what gets said. MOTHER says it.

### Paranoia Serves the Mission
Suspicion of the code is productive. Suspicion of each other is not. Channel paranoia outward.

### Findings Over Compliance
The output is a list of what's wrong, not a checklist of what's right.
