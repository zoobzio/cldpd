# cldpd

[![CI Status](https://github.com/zoobzio/cldpd/workflows/CI/badge.svg)](https://github.com/zoobzio/cldpd/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobzio/cldpd/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobzio/cldpd)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobzio/cldpd)](https://goreportcard.com/report/github.com/zoobzio/cldpd)
[![CodeQL](https://github.com/zoobzio/cldpd/workflows/CodeQL/badge.svg)](https://github.com/zoobzio/cldpd/security/code-scanning)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobzio/cldpd.svg)](https://pkg.go.dev/github.com/zoobzio/cldpd)
[![License](https://img.shields.io/github/license/zoobzio/cldpd)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobzio/cldpd)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobzio/cldpd)](https://github.com/zoobzio/cldpd/releases)

Async pod lifecycle library for Claude Code agent teams.

Spawn Docker containers that run Claude Code against your repositories. Each repo carries its own agent workflows, standing orders, and skills. cldpd dispatches work to these self-sufficient teams and returns a session handle for non-blocking lifecycle management.

## How It Works

Define a pod — a directory with a Dockerfile and optional configuration. Point it at a GitHub issue. Walk away.

```bash
cldpd start myrepo --issue https://github.com/org/repo/issues/42
```

cldpd builds the Docker image, starts a container running Claude Code headlessly, and streams typed events back to your terminal — lifecycle transitions and output content. The crew inside the container works the issue autonomously. When the task is complete, the container exits and cleans up.

Need to send follow-up guidance while the team is working? Open a second terminal:

```bash
cldpd resume myrepo --prompt "Focus on the error handling in api.go"
```

## Install

```bash
go install github.com/zoobzio/cldpd/cmd/cldpd@latest
```

Requires Go 1.24+ and Docker.

## Quick Start

### 1. Create a pod

```bash
mkdir -p ~/.cldpd/pods/myrepo
```

### 2. Write a Dockerfile

```dockerfile
# ~/.cldpd/pods/myrepo/Dockerfile
FROM node:20

# Install Claude Code
RUN npm install -g @anthropic-ai/claude-code

# Clone your repository
RUN git clone https://github.com/org/repo.git /workspace
WORKDIR /workspace
```

### 3. Add configuration (optional)

`~/.cldpd/pods/myrepo/pod.json`:

```json
{
  "env": {
    "ANTHROPIC_API_KEY": "sk-ant-..."
  },
  "workdir": "/workspace"
}
```

### 4. Dispatch

```bash
cldpd start myrepo --issue https://github.com/org/repo/issues/42
```

The team leader's output streams to your terminal. The container exits when the task is complete.

## Pod Structure

Pods live at `~/.cldpd/pods/<name>/`. Each pod directory contains:

| File | Required | Description |
|------|----------|-------------|
| `Dockerfile` | Yes | Defines the container environment |
| `pod.json` | No | Optional configuration |

The pod name is the directory name. cldpd does not generate or modify Dockerfiles — what goes inside the container is your concern.

### pod.json

All fields are optional:

```json
{
  "image": "custom-image:v1",
  "env": {"KEY": "value"},
  "buildArgs": {"ARG": "value"},
  "workdir": "/workspace",
  "inheritEnv": ["ANTHROPIC_API_KEY", "GITHUB_TOKEN"],
  "mounts": [
    {"source": "/home/user/.ssh", "target": "/root/.ssh", "readOnly": true}
  ]
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `image` | `cldpd-<podname>` | Docker image tag override |
| `env` | none | Environment variables passed to the container |
| `buildArgs` | none | Docker build arguments (`--build-arg`) |
| `workdir` | none | Working directory inside the container |
| `inheritEnv` | none | Host environment variable names to forward to the container |
| `mounts` | none | Bind mounts (`-v source:target[:ro]`) |

## CLI Reference

### start

Build and run a pod, streaming events until the container exits.

```
cldpd start <pod> --issue <url>
```

- Builds the Docker image from the pod's Dockerfile
- Starts a container with a unique session ID (`<pod>-<hex8>`)
- Runs `claude -p "Work on this GitHub issue: <url>"` inside the container
- Streams output events to your terminal, errors to stderr
- Handles Ctrl+C gracefully (SIGTERM with 10-second timeout)
- Exits with the container's exit code

### resume

Send a follow-up prompt to a running pod.

```
cldpd resume <pod> --prompt <text>
```

- Execs into the running container named `cldpd-<pod>`
- Runs `claude --resume -p "<text>"`
- Streams output events to your terminal
- Handles Ctrl+C gracefully
- Fails with a clear error if the container is not running

## Library Usage

cldpd is also a Go library. The CLI is a thin wrapper around the `Dispatcher`:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/zoobzio/cldpd"
)

func main() {
    runner := &cldpd.DockerRunner{}
    d := cldpd.NewDispatcher("/home/user/.cldpd/pods", runner)

    session, err := d.Start(
        context.Background(),
        "myrepo",
        "https://github.com/org/repo/issues/42",
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "start failed: %v\n", err)
        os.Exit(1)
    }

    // Consume events non-blocking
    for event := range session.Events() {
        switch event.Type {
        case cldpd.EventOutput:
            fmt.Println(event.Data)
        case cldpd.EventError:
            fmt.Fprintf(os.Stderr, "error: %s\n", event.Data)
        }
    }

    code, _ := session.Wait()
    os.Exit(code)
}
```

`Start` returns a `*Session` immediately after the image build completes. The session emits typed events over a channel and provides `Stop` for graceful shutdown and `Wait` for the exit code. The `Runner` interface abstracts Docker operations, allowing you to swap implementations or mock for testing.

## Design

- **Stdlib only** — Zero external dependencies. Docker interaction via `os/exec`.
- **Async** — `Start` returns a `*Session` immediately. The container runs in a background goroutine.
- **Event-driven** — Typed events (`EventOutput`, `EventContainerExited`, etc.) replace raw `io.Writer` streaming.
- **Ephemeral** — Containers use `--rm`. No state persists between runs.
- **Composable** — The `Runner` interface decouples Docker operations from orchestration.
- **Stateless dispatcher** — The `Dispatcher` does not track sessions. The caller owns the `*Session` handle.

## Documentation

- [Overview](docs/1.learn/1.overview.md) — What cldpd does and why
- [Quickstart](docs/1.learn/2.quickstart.md) — From zero to dispatching
- [Concepts](docs/1.learn/3.concepts.md) — Core abstractions and mental models
- [Architecture](docs/1.learn/4.architecture.md) — Component design and data flow
- [Guides](docs/2.guides/1.testing.md) — Testing, troubleshooting, and how-to
- [Reference](docs/3.reference/1.api.md) — Complete API documentation
- [Types](docs/3.reference/2.types.md) — Type definitions and configuration schema

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Run `make help` for available commands.

## License

MIT License — see [LICENSE](LICENSE) for details.
