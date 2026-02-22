# Review Criteria

Secret, repo-specific review criteria. Only Armitage reads this file.

## Mission Criteria

### What This Repo MUST Achieve

- Library must support non-blocking pod lifecycle (Start returns a Session handle immediately)
- Event-based output must distinguish lifecycle events from pod output
- Session identity must be explicit and deterministic for concurrent pod management
- Graceful shutdown must send SIGTERM with timeout before SIGKILL
- CLI must work as a trivial consumer of the library API

### What This Repo MUST NOT Contain

- No external dependencies beyond the Go standard library
- No Docker SDK — Docker CLI via os/exec only
- No GitHub awareness — cldpd is a pure container lifecycle library
- No secrets written to disk — credential passthrough via CLI flags only
- No application logic beyond pod dispatch

## Review Priorities

Ordered by importance. When findings conflict, higher-priority items take precedence.

1. Security: no credential leakage, no secrets on disk, no unsafe container defaults
2. Correctness: Session lifecycle is clean — goroutines exit, channels close, no leaks
3. API surface: Runner interface remains mockable, Session is testable without Docker
4. Completeness: all pod lifecycle phases (build, run, stop, resume) are covered
5. Quality: conventions are consistent, documentation matches implementation

## Severity Calibration

Guidance for how Armitage classifies finding severity for this specific repo.

| Condition | Severity |
|-----------|----------|
| Credential or secret written to disk or logged | Critical |
| Goroutine leak in Session lifecycle | Critical |
| Channel not closed on container exit | High |
| Runner interface change breaks mockability | High |
| Missing event type for a lifecycle transition | Medium |
| Godoc gap on exported symbol | Medium |
| Minor inconsistency in error sentinel naming | Low |
| Documentation wording does not match current API | Low |

## Standing Concerns

Persistent issues or areas of known weakness that should always be checked.

- Goroutine lifecycle in Session — every goroutine spawned by Start must exit cleanly on container exit, Stop, or context cancellation
- Channel backpressure — slow consumers must not block the event-writing goroutine indefinitely
- io.Pipe lifecycle — both ends must close correctly in all exit paths
- Docker CLI error parsing — os/exec output is unstructured and fragile

## Out of Scope

Things the red team should NOT flag for this repo, even if they look wrong.

- Interactive attach (bidirectional pty I/O) is explicitly deferred — absence is intentional
- The CLI is intentionally minimal — it is a thin consumer of the library, not a feature-rich tool
- Pod definitions are filesystem-based by design — no database or API backing is expected
- MISSION.md references future TUI orchestrator — this is context, not a deliverable for this repo
