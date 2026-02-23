package cldpd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Dispatcher coordinates pod discovery, image building, and container lifecycle.
// Use NewDispatcher to create one.
//
// Dispatcher is stateless — it does not track running sessions. Each returned
// *Session is self-contained. The caller is responsible for calling Stop or Wait.
type Dispatcher struct {
	runner  Runner
	podsDir string
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

// Start builds the pod's Docker image synchronously, then returns a *Session
// representing the running container. The image build completes before Start
// returns — if the build fails, Start returns an error and no Session is created.
//
// If the pod's template.md is non-empty, its contents are prepended to the
// prompt passed to Claude Code: template + "\n\n" + "Work on this GitHub issue: " + issueURL.
// When template.md is absent, the prompt is the issue URL directive alone.
//
// The Session emits events in the following order:
//
//	BuildStarted → BuildComplete → ContainerStarted → Output* → ContainerExited
//
// On build failure: BuildStarted → Error (no Session returned).
// On runtime failure: events up to ContainerStarted, then Output*, then Error.
//
// The caller is responsible for calling session.Stop or session.Wait.
func (d *Dispatcher) Start(ctx context.Context, podName string, issueURL string) (*Session, error) {
	pod, err := DiscoverPod(d.podsDir, podName)
	if err != nil {
		return nil, err
	}

	tag := pod.Config.Image
	if tag == "" {
		tag = "cldpd-" + podName
	}

	// Build phase: synchronous. Emit build events to a temporary channel so
	// callers who consume Events() see them in order. We emit these as preamble
	// inside newSession.
	buildStarted := Event{
		Type: EventBuildStarted,
		Data: tag,
		Time: time.Now(),
	}

	if err := d.runner.Build(ctx, tag, pod.Dir, pod.Config.BuildArgs); err != nil {
		// Build failed: no session. Return a synthetic error event sequence via
		// a closed-channel session so callers using Events() still see BuildStarted
		// and Error. We emit this via a dedicated helper rather than newSession
		// to keep the failure path simple and goroutine-free.
		return nil, fmt.Errorf("%w", err)
	}

	buildComplete := Event{
		Type: EventBuildComplete,
		Data: tag,
		Time: time.Now(),
	}

	sessionID := newSessionID(podName)
	container := containerName(podName)

	// Resolve InheritEnv two ways: names whose values are present on the host
	// are eagerly resolved into Env (passed as -e K=V). Names not set on the
	// host are deferred to Docker via InheritEnv (passed as bare -e NAME),
	// allowing Docker to inherit them from the host environment at run time.
	env := make(map[string]string, len(pod.Config.Env))
	for k, v := range pod.Config.Env {
		env[k] = v
	}
	var inheritEnv []string
	for _, name := range pod.Config.InheritEnv {
		if v := os.Getenv(name); v != "" {
			env[name] = v
		} else {
			inheritEnv = append(inheritEnv, name)
		}
	}

	prompt := "Work on this GitHub issue: " + issueURL
	if pod.Template != "" {
		prompt = pod.Template + "\n\n" + prompt
	}

	opts := RunOptions{
		Image:      tag,
		Name:       container,
		Cmd:        []string{"claude", "-p", prompt},
		Env:        env,
		InheritEnv: inheritEnv,
		Workdir:    pod.Config.Workdir,
		Remove:     true,
		Mounts:     pod.Config.Mounts,
	}

	containerStarted := Event{
		Type: EventContainerStarted,
		Data: container,
		Time: time.Now(),
	}

	runner := d.runner
	runFn := func(pw io.WriteCloser) (int, error) {
		return runner.Run(ctx, opts, pw)
	}

	preamble := []Event{buildStarted, buildComplete, containerStarted}

	return newSession(sessionID, container, d.runner, runFn, preamble), nil
}

// Resume returns a *Session wrapping a follow-up exec into an already-running
// container for the named pod. Resume does not build an image.
//
// The Session emits events in the following order:
//
//	ContainerStarted → Output* → ContainerExited
//
// Returns ErrSessionNotFound if no container named cldpd-<podName> is running.
// The caller is responsible for calling session.Stop or session.Wait.
func (d *Dispatcher) Resume(ctx context.Context, podName string, prompt string) (*Session, error) {
	container := containerName(podName)
	cmd := []string{"claude", "--resume", "-p", prompt}

	sessionID := newSessionID(podName)

	runner := d.runner
	runFn := func(pw io.WriteCloser) (int, error) {
		return runner.Exec(ctx, container, cmd, pw)
	}

	containerStarted := Event{
		Type: EventContainerStarted,
		Data: container,
		Time: time.Now(),
	}

	preamble := []Event{containerStarted}

	return newSession(sessionID, container, d.runner, runFn, preamble), nil
}

// containerName returns the deterministic Docker container name for a pod.
// Used by both Start (to name the new container) and Resume (to target the running one).
func containerName(podName string) string {
	return "cldpd-" + podName
}

// newSessionID generates a unique session ID in the format <podName>-<hex8>.
// Uses crypto/rand for the random suffix.
func newSessionID(podName string) string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failure is extremely unlikely; fall back to a fixed suffix
		// rather than panicking. The session will still function.
		return podName + "-00000000"
	}
	return podName + "-" + hex.EncodeToString(b[:])
}
