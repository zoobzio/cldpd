# Mission: cldpd

Pod dispatcher for Claude Code agent teams.

## Purpose

Spawn Docker containers ("pods") that run Claude Code against zoobzio repositories. Each repo carries its own agent workflows, standing orders, and skills. cldpd dispatches work to these self-sufficient teams and monitors their output.

## What This Package Does

- Discovers pod definitions from a conventional directory structure (`~/.cldpd/pods/<name>/`)
- Builds Docker images from user-provided Dockerfiles
- Starts containers with a structured prompt pointing Claude Code at a GitHub issue
- Streams the team leader's output back to the caller as the direct line
- Supports follow-up commands to running pods via session resume
- Provides a CLI that Claude Code in the user's terminal can invoke as a tool

## What This Package Does NOT Do

- Orchestrate agent workflows — the repos already define those
- Parse or interpret Claude Code's output — the team leader narrates; cldpd relays
- Manage custom reporting protocols — stdout is the direct line
- Persist state between runs — pods are ephemeral

## How It Works

1. The caller (user or Claude) invokes `cldpd start <pod> --issue <url>`
2. cldpd builds the pod's Docker image and starts a container
3. The container runs Claude Code headlessly (`claude -p`) with the repo's existing `.claude/` configuration
4. The crew (agents, skills, standing orders already in the repo) works the issue autonomously
5. The team leader's narration streams back via container stdout
6. Escalation routes through GitHub (issue comments) rather than interactive prompts
7. The container exits when the task is complete

## Success Criteria

A user can:
1. Define a pod for any zoobzio repo (Dockerfile + optional config)
2. Point it at a GitHub issue with guidance
3. Walk away while the crew works
4. Read the team leader's output stream to monitor progress
5. Find the completed work (PR, commits, issue comments) on GitHub

## Non-Goals

- Daemon or persistent background service (the CLI blocks for the duration of a run)
- Custom inter-process communication (no MCP servers, no HTTP applets)
- Telemetry infrastructure (OTEL integration is a future concern)
- Multi-pod orchestration from a single process (run multiple terminals)
- External dependencies beyond the Go standard library
