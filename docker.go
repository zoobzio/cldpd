package cldpd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner is the interface over Docker CLI operations.
// All methods block until the operation completes and stream output to the
// provided io.Writer where applicable.
type Runner interface {
	// Build builds a Docker image tagged with tag from the Dockerfile in dir.
	// buildArgs are passed as --build-arg K=V flags.
	// Returns ErrBuildFailed if the build exits with a non-zero status.
	Build(ctx context.Context, tag string, dir string, buildArgs map[string]string) error

	// Run starts a container with the given options, streams its stdout to the
	// provided writer, blocks until the container exits, and returns the exit code.
	// A non-zero exit code is not itself an error — the caller interprets it.
	Run(ctx context.Context, opts RunOptions, stdout io.Writer) (int, error)

	// Exec runs a command in an already-running container, streams its stdout
	// to the provided writer, blocks until the command exits, and returns the exit code.
	// Returns ErrSessionNotFound if the container is not running.
	Exec(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error)
}

// RunOptions configures a docker run invocation.
type RunOptions struct {
	Image   string            // Docker image to run
	Name    string            // container name (--name); used for deterministic resume
	Cmd     []string          // command and arguments to run inside the container
	Env     map[string]string // environment variables (-e K=V)
	Workdir string            // working directory inside the container (-w)
	Remove  bool              // remove the container after it exits (--rm)
}

// DockerRunner implements Runner using the Docker CLI via os/exec.
type DockerRunner struct{}

// Preflight checks that the Docker daemon is reachable by running docker info.
// Returns ErrDockerUnavailable if the daemon cannot be contacted.
func (d *DockerRunner) Preflight(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %w", ErrDockerUnavailable, err)
	}
	return nil
}

// buildCmdArgs returns the docker CLI arguments for a build invocation.
func buildCmdArgs(tag string, dir string, buildArgs map[string]string) []string {
	args := []string{"build", "-t", tag}
	for k, v := range buildArgs {
		args = append(args, "--build-arg", k+"="+v)
	}
	args = append(args, dir)
	return args
}

// runCmdArgs returns the docker CLI arguments for a run invocation.
func runCmdArgs(opts RunOptions) []string {
	args := []string{"run"}
	if opts.Remove {
		args = append(args, "--rm")
	}
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	for k, v := range opts.Env {
		args = append(args, "-e", k+"="+v)
	}
	if opts.Workdir != "" {
		args = append(args, "-w", opts.Workdir)
	}
	args = append(args, opts.Image)
	args = append(args, opts.Cmd...)
	return args
}

// execCmdArgs returns the docker CLI arguments for an exec invocation.
func execCmdArgs(container string, cmd []string) []string {
	return append([]string{"exec", container}, cmd...)
}

// Build builds a Docker image tagged with tag from the Dockerfile in dir.
func (d *DockerRunner) Build(ctx context.Context, tag string, dir string, buildArgs map[string]string) error {
	args := buildCmdArgs(tag, dir, buildArgs)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = io.Discard
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("%w: exit code %d: %s", ErrBuildFailed, exitErr.ExitCode(), stderr.String())
		}
		return fmt.Errorf("%w: %w", ErrBuildFailed, err)
	}
	return nil
}

// Run starts a container with the given options, streams stdout, and blocks
// until the container exits. Returns the container's exit code.
func (d *DockerRunner) Run(ctx context.Context, opts RunOptions, stdout io.Writer) (int, error) {
	args := runCmdArgs(opts)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return -1, fmt.Errorf("docker run: %w", err)
	}
	return 0, nil
}

// Exec runs a command in an already-running container and streams its stdout.
// Returns ErrSessionNotFound if the container does not exist or is not running.
// For all other non-zero exits the exit code is returned with a nil error.
func (d *DockerRunner) Exec(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error) {
	// Preflight: verify the container exists and is running.
	// docker inspect exits non-zero if the container does not exist.
	inspect := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Running}}", container)
	out, err := inspect.Output()
	if err != nil {
		return -1, fmt.Errorf("%s: %w", container, ErrSessionNotFound)
	}
	if strings.TrimSpace(string(out)) != "true" {
		return -1, fmt.Errorf("%s: %w", container, ErrSessionNotFound)
	}

	args := execCmdArgs(container, cmd)
	c := exec.CommandContext(ctx, "docker", args...)
	c.Stdout = stdout
	c.Stderr = io.Discard

	err = c.Run()
	if err == nil {
		return 0, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}

	// Non-ExitError: context cancelled or other process failure.
	return -1, err
}
