//go:build testing

package integration

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"testing"

	"github.com/zoobzio/cldpd"
)

// dockerAvailable reports whether a Docker daemon is reachable.
func dockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func TestDockerRunner_Preflight_Available(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}

	r := &cldpd.DockerRunner{}
	err := r.Preflight(context.Background())
	if err != nil {
		t.Errorf("Preflight failed with Docker available: %v", err)
	}
}

func TestDockerRunner_Preflight_ContextCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled immediately

	r := &cldpd.DockerRunner{}
	err := r.Preflight(ctx)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

func TestDockerRunner_Build_InvalidDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}

	r := &cldpd.DockerRunner{}
	err := r.Build(context.Background(), "cldpd-test-build-invalid", "/nonexistent/path", nil)
	if err == nil {
		t.Error("expected error building from nonexistent dir, got nil")
	}
}

func TestDockerRunner_Run_HelloWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}

	var buf bytes.Buffer
	r := &cldpd.DockerRunner{}
	opts := cldpd.RunOptions{
		Image:  "alpine:latest",
		Name:   "cldpd-test-run-hello",
		Cmd:    []string{"echo", "hello"},
		Remove: true,
	}
	code, err := r.Run(context.Background(), opts, &buf)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	// Clean up in case --rm didn't fire
	exec.Command("docker", "rm", "-f", "cldpd-test-run-hello").Run() //nolint:errcheck
}

func TestDockerRunner_Run_NonZeroExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}

	r := &cldpd.DockerRunner{}
	opts := cldpd.RunOptions{
		Image:  "alpine:latest",
		Name:   "cldpd-test-run-exit1",
		Cmd:    []string{"sh", "-c", "exit 2"},
		Remove: true,
	}
	code, err := r.Run(context.Background(), opts, io.Discard)
	if err != nil {
		t.Fatalf("unexpected process error: %v", err)
	}
	if code != 2 {
		t.Errorf("exit code: got %d, want 2", code)
	}
	exec.Command("docker", "rm", "-f", "cldpd-test-run-exit1").Run() //nolint:errcheck
}

func TestDockerRunner_Exec_NotRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}

	r := &cldpd.DockerRunner{}
	// Container does not exist — docker inspect preflight returns an error,
	// which Exec maps to ErrSessionNotFound.
	_, err := r.Exec(context.Background(), "cldpd-test-nonexistent-container", []string{"echo", "hi"}, io.Discard)
	if !errors.Is(err, cldpd.ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}
