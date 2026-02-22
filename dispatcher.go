package cldpd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Dispatcher coordinates pod discovery, image building, and container lifecycle.
// Use NewDispatcher to create one.
type Dispatcher struct {
	podsDir string
	runner  Runner
}

// NewDispatcher returns a Dispatcher that discovers pods from podsDir and
// executes Docker operations via runner.
func NewDispatcher(podsDir string, runner Runner) *Dispatcher {
	return &Dispatcher{
		podsDir: podsDir,
		runner:  runner,
	}
}

// DefaultPodsDir returns the conventional pods directory: ~/.cldpd/pods/.
func DefaultPodsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".cldpd", "pods"), nil
}

// Start builds the pod's Docker image, runs a container with a prompt pointing
// Claude Code at the given GitHub issue URL, streams container stdout to the
// provided writer, and blocks until the container exits.
//
// Returns the container's exit code. A non-zero exit code also returns
// ErrContainerFailed.
func (d *Dispatcher) Start(ctx context.Context, podName string, issueURL string, stdout io.Writer) (int, error) {
	pod, err := DiscoverPod(d.podsDir, podName)
	if err != nil {
		return -1, err
	}

	tag := pod.Config.Image
	if tag == "" {
		tag = "cldpd-" + podName
	}

	if err := d.runner.Build(ctx, tag, pod.Dir, pod.Config.BuildArgs); err != nil {
		return -1, err
	}

	prompt := "Work on this GitHub issue: " + issueURL

	opts := RunOptions{
		Image:   tag,
		Name:    containerName(podName),
		Cmd:     []string{"claude", "-p", prompt},
		Env:     pod.Config.Env,
		Workdir: pod.Config.Workdir,
		Remove:  true,
	}

	code, err := d.runner.Run(ctx, opts, stdout)
	if err != nil {
		return -1, err
	}
	if code != 0 {
		return code, fmt.Errorf("%w: exit code %d", ErrContainerFailed, code)
	}
	return code, nil
}

// Resume execs a follow-up command into an already-running container for the
// named pod. The prompt is passed to claude --resume. Streams output to the
// provided writer and blocks until the command exits.
//
// Returns ErrSessionNotFound if no container named cldpd-<podName> is running.
func (d *Dispatcher) Resume(ctx context.Context, podName string, prompt string, stdout io.Writer) (int, error) {
	name := containerName(podName)
	cmd := []string{"claude", "--resume", "-p", prompt}

	code, err := d.runner.Exec(ctx, name, cmd, stdout)
	if err != nil {
		return -1, err
	}
	return code, nil
}

// containerName returns the deterministic Docker container name for a pod.
func containerName(podName string) string {
	return "cldpd-" + podName
}
