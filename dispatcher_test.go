//go:build testing

package cldpd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeTestPod creates a minimal valid pod directory in podsDir and returns the dispatcher.
func makeTestPod(t *testing.T, podsDir, name string) {
	t.Helper()
	dir := filepath.Join(podsDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create pod dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
}

func TestContainerName(t *testing.T) {
	cases := []struct {
		podName string
		want    string
	}{
		{"myrepo", "cldpd-myrepo"},
		{"some-repo", "cldpd-some-repo"},
		{"a", "cldpd-a"},
	}
	for _, tc := range cases {
		got := containerName(tc.podName)
		if got != tc.want {
			t.Errorf("containerName(%q): got %q, want %q", tc.podName, got, tc.want)
		}
	}
}

func TestDefaultPodsDir(t *testing.T) {
	dir, err := DefaultPodsDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home dir: %v", err)
	}
	want := filepath.Join(home, ".cldpd", "pods")
	if dir != want {
		t.Errorf("DefaultPodsDir: got %q, want %q", dir, want)
	}
}

func TestNewDispatcher(t *testing.T) {
	r := &mockRunner{}
	d := NewDispatcher("/some/path", r)
	if d == nil {
		t.Fatal("NewDispatcher returned nil")
	}
	if d.podsDir != "/some/path" {
		t.Errorf("podsDir: got %q, want %q", d.podsDir, "/some/path")
	}
	if d.runner != r {
		t.Error("runner not stored correctly")
	}
}

func TestDispatcher_Start_PodNotFound(t *testing.T) {
	podsDir := t.TempDir()
	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	_, err := d.Start(context.Background(), "ghost", "https://github.com/org/repo/issues/1", io.Discard)
	if !errors.Is(err, ErrPodNotFound) {
		t.Errorf("got %v, want ErrPodNotFound", err)
	}
}

func TestDispatcher_Start_InvalidPod(t *testing.T) {
	podsDir := t.TempDir()
	// Directory with no Dockerfile
	dir := filepath.Join(podsDir, "nodocker")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create dir: %v", err)
	}

	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	_, err := d.Start(context.Background(), "nodocker", "https://github.com/org/repo/issues/1", io.Discard)
	if !errors.Is(err, ErrInvalidPod) {
		t.Errorf("got %v, want ErrInvalidPod", err)
	}
}

func TestDispatcher_Start_DefaultImageTag(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	var builtTag string
	r := &mockRunner{
		buildFn: func(_ context.Context, tag string, _ string, _ map[string]string) error {
			builtTag = tag
			return nil
		},
	}
	d := NewDispatcher(podsDir, r)

	_, _ = d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1", io.Discard)

	if builtTag != "cldpd-myrepo" {
		t.Errorf("image tag: got %q, want %q", builtTag, "cldpd-myrepo")
	}
}

func TestDispatcher_Start_CustomImageTag(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")
	// Write pod.json with custom image
	dir := filepath.Join(podsDir, "myrepo")
	if err := os.WriteFile(filepath.Join(dir, "pod.json"), []byte(`{"image":"custom:v1"}`), 0644); err != nil {
		t.Fatalf("write pod.json: %v", err)
	}

	var builtTag string
	r := &mockRunner{
		buildFn: func(_ context.Context, tag string, _ string, _ map[string]string) error {
			builtTag = tag
			return nil
		},
	}
	d := NewDispatcher(podsDir, r)

	_, _ = d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1", io.Discard)

	if builtTag != "custom:v1" {
		t.Errorf("image tag: got %q, want %q", builtTag, "custom:v1")
	}
}

func TestDispatcher_Start_BuildFailed(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{
		buildFn: func(_ context.Context, _ string, _ string, _ map[string]string) error {
			return fmt.Errorf("%w: exit code 1", ErrBuildFailed)
		},
	}
	d := NewDispatcher(podsDir, r)

	_, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1", io.Discard)
	if !errors.Is(err, ErrBuildFailed) {
		t.Errorf("got %v, want ErrBuildFailed", err)
	}
}

func TestDispatcher_Start_RunOptions(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	var capturedOpts RunOptions
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedOpts = opts
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	issueURL := "https://github.com/org/repo/issues/42"
	_, err := d.Start(context.Background(), "myrepo", issueURL, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedOpts.Name != "cldpd-myrepo" {
		t.Errorf("container name: got %q, want %q", capturedOpts.Name, "cldpd-myrepo")
	}
	if capturedOpts.Image != "cldpd-myrepo" {
		t.Errorf("image: got %q, want %q", capturedOpts.Image, "cldpd-myrepo")
	}
	if !capturedOpts.Remove {
		t.Error("Remove: got false, want true")
	}
	if len(capturedOpts.Cmd) < 3 {
		t.Fatalf("Cmd too short: %v", capturedOpts.Cmd)
	}
	if capturedOpts.Cmd[0] != "claude" {
		t.Errorf("Cmd[0]: got %q, want %q", capturedOpts.Cmd[0], "claude")
	}
	if !strings.Contains(strings.Join(capturedOpts.Cmd, " "), issueURL) {
		t.Errorf("Cmd does not contain issue URL %q: %v", issueURL, capturedOpts.Cmd)
	}
}

func TestDispatcher_Start_StreamsOutput(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{
		runFn: func(_ context.Context, _ RunOptions, stdout io.Writer) (int, error) {
			_, _ = stdout.Write([]byte("container output"))
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	var buf bytes.Buffer
	_, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "container output" {
		t.Errorf("output: got %q, want %q", buf.String(), "container output")
	}
}

func TestDispatcher_Start_NonZeroExitReturnsErrContainerFailed(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{
		runFn: func(_ context.Context, _ RunOptions, _ io.Writer) (int, error) {
			return 2, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	code, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1", io.Discard)
	if !errors.Is(err, ErrContainerFailed) {
		t.Errorf("got %v, want ErrContainerFailed", err)
	}
	if code != 2 {
		t.Errorf("exit code: got %d, want 2", code)
	}
}

func TestDispatcher_Start_SuccessReturnsZero(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	code, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1", io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
}

func TestDispatcher_Resume_ContainerName(t *testing.T) {
	podsDir := t.TempDir()

	var execContainer string
	r := &mockRunner{
		execFn: func(_ context.Context, container string, _ []string, _ io.Writer) (int, error) {
			execContainer = container
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	_, err := d.Resume(context.Background(), "myrepo", "do more work", io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if execContainer != "cldpd-myrepo" {
		t.Errorf("container: got %q, want %q", execContainer, "cldpd-myrepo")
	}
}

func TestDispatcher_Resume_Command(t *testing.T) {
	podsDir := t.TempDir()

	var execCmd []string
	r := &mockRunner{
		execFn: func(_ context.Context, _ string, cmd []string, _ io.Writer) (int, error) {
			execCmd = cmd
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	_, err := d.Resume(context.Background(), "myrepo", "do more work", io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"claude", "--resume", "-p", "do more work"}
	if len(execCmd) != len(want) {
		t.Fatalf("cmd: got %v, want %v", execCmd, want)
	}
	for i := range want {
		if execCmd[i] != want[i] {
			t.Errorf("cmd[%d]: got %q, want %q", i, execCmd[i], want[i])
		}
	}
}

func TestDispatcher_Resume_SessionNotFound(t *testing.T) {
	podsDir := t.TempDir()

	r := &mockRunner{
		execFn: func(_ context.Context, container string, _ []string, _ io.Writer) (int, error) {
			return -1, fmt.Errorf("%w: %s", ErrSessionNotFound, container)
		},
	}
	d := NewDispatcher(podsDir, r)

	_, err := d.Resume(context.Background(), "ghost", "guidance", io.Discard)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}

func TestDispatcher_Resume_StreamsOutput(t *testing.T) {
	podsDir := t.TempDir()

	r := &mockRunner{
		execFn: func(_ context.Context, _ string, _ []string, stdout io.Writer) (int, error) {
			_, _ = stdout.Write([]byte("resume output"))
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	var buf bytes.Buffer
	_, err := d.Resume(context.Background(), "myrepo", "guidance", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "resume output" {
		t.Errorf("output: got %q, want %q", buf.String(), "resume output")
	}
}

func TestDispatcher_Resume_ReturnsExitCode(t *testing.T) {
	podsDir := t.TempDir()

	r := &mockRunner{
		execFn: func(_ context.Context, _ string, _ []string, _ io.Writer) (int, error) {
			return 3, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	code, err := d.Resume(context.Background(), "myrepo", "guidance", io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 3 {
		t.Errorf("exit code: got %d, want 3", code)
	}
}
