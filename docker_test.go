//go:build testing

package cldpd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// mockRunner is a test double for Runner.
type mockRunner struct {
	preflightFn func(ctx context.Context) error
	buildFn     func(ctx context.Context, tag string, dir string, buildArgs map[string]string) error
	runFn       func(ctx context.Context, opts RunOptions, stdout io.Writer) (int, error)
	execFn      func(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error)
	stopFn      func(ctx context.Context, container string, timeout time.Duration) error
}

func (m *mockRunner) Preflight(ctx context.Context) error {
	if m.preflightFn != nil {
		return m.preflightFn(ctx)
	}
	return nil
}

func (m *mockRunner) Build(ctx context.Context, tag string, dir string, buildArgs map[string]string) error {
	if m.buildFn != nil {
		return m.buildFn(ctx, tag, dir, buildArgs)
	}
	return nil
}

func (m *mockRunner) Run(ctx context.Context, opts RunOptions, stdout io.Writer) (int, error) {
	if m.runFn != nil {
		return m.runFn(ctx, opts, stdout)
	}
	return 0, nil
}

func (m *mockRunner) Exec(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error) {
	if m.execFn != nil {
		return m.execFn(ctx, container, cmd, stdout)
	}
	return 0, nil
}

func (m *mockRunner) Stop(ctx context.Context, container string, timeout time.Duration) error {
	if m.stopFn != nil {
		return m.stopFn(ctx, container, timeout)
	}
	return nil
}

// Compile-time interface assertions.
var _ Runner = (*DockerRunner)(nil)
var _ Runner = (*mockRunner)(nil)

func TestBuildCmdArgs_Minimal(t *testing.T) {
	args := buildCmdArgs("myimage:latest", "/some/dir", nil)
	want := []string{"build", "-t", "myimage:latest", "/some/dir"}
	if len(args) != len(want) {
		t.Fatalf("args: got %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d]: got %q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildCmdArgs_WithBuildArgs(t *testing.T) {
	args := buildCmdArgs("img", "/dir", map[string]string{"KEY": "val"})
	// Must contain --build-arg KEY=val before the dir.
	var foundBuildArg bool
	for i, a := range args {
		if a == "--build-arg" && i+1 < len(args) && args[i+1] == "KEY=val" {
			foundBuildArg = true
		}
	}
	if !foundBuildArg {
		t.Errorf("args missing --build-arg KEY=val: %v", args)
	}
	if args[len(args)-1] != "/dir" {
		t.Errorf("last arg should be dir, got %q", args[len(args)-1])
	}
}

func TestRunCmdArgs_Minimal(t *testing.T) {
	opts := RunOptions{Image: "myimage"}
	args := runCmdArgs(opts)
	if args[0] != "run" {
		t.Errorf("args[0]: got %q, want %q", args[0], "run")
	}
	// Last arg should be the image.
	if args[len(args)-1] != "myimage" {
		t.Errorf("last arg: got %q, want %q", args[len(args)-1], "myimage")
	}
}

func TestRunCmdArgs_WithAllOptions(t *testing.T) {
	opts := RunOptions{
		Image:   "myimage:latest",
		Name:    "cldpd-myrepo",
		Cmd:     []string{"claude", "-p", "prompt"},
		Env:     map[string]string{"FOO": "bar"},
		Workdir: "/workspace",
		Remove:  true,
	}
	args := runCmdArgs(opts)

	var hasRM, hasName, hasEnv, hasWorkdir bool
	for i, a := range args {
		switch a {
		case "--rm":
			hasRM = true
		case "--name":
			if i+1 < len(args) && args[i+1] == "cldpd-myrepo" {
				hasName = true
			}
		case "-e":
			if i+1 < len(args) && args[i+1] == "FOO=bar" {
				hasEnv = true
			}
		case "-w":
			if i+1 < len(args) && args[i+1] == "/workspace" {
				hasWorkdir = true
			}
		}
	}
	if !hasRM {
		t.Error("missing --rm")
	}
	if !hasName {
		t.Error("missing --name cldpd-myrepo")
	}
	if !hasEnv {
		t.Error("missing -e FOO=bar")
	}
	if !hasWorkdir {
		t.Error("missing -w /workspace")
	}
	// Image and cmd should be at the end.
	var imageIdx int
	for i, a := range args {
		if a == "myimage:latest" {
			imageIdx = i
		}
	}
	if imageIdx == 0 {
		t.Error("image not found in args")
	}
	// Cmd elements should follow image.
	if len(args) < imageIdx+4 {
		t.Errorf("cmd not found after image: %v", args)
	}
}

func TestExecCmdArgs(t *testing.T) {
	args := execCmdArgs("cldpd-myrepo", []string{"claude", "--resume", "-p", "prompt"})
	want := []string{"exec", "cldpd-myrepo", "claude", "--resume", "-p", "prompt"}
	if len(args) != len(want) {
		t.Fatalf("args: got %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d]: got %q, want %q", i, args[i], want[i])
		}
	}
}

func TestRunCmdArgs_NoRemove(t *testing.T) {
	opts := RunOptions{Image: "img", Remove: false}
	args := runCmdArgs(opts)
	for _, a := range args {
		if a == "--rm" {
			t.Error("--rm should not be present when Remove=false")
		}
	}
}

func TestRunCmdArgs_NoName(t *testing.T) {
	opts := RunOptions{Image: "img"}
	args := runCmdArgs(opts)
	for i, a := range args {
		if a == "--name" {
			t.Errorf("--name should not be present when Name is empty, found at %d", i)
		}
	}
}

func TestRunCmdArgs_NoWorkdir(t *testing.T) {
	opts := RunOptions{Image: "img"}
	args := runCmdArgs(opts)
	for i, a := range args {
		if a == "-w" {
			t.Errorf("-w should not be present when Workdir is empty, found at %d", i)
		}
	}
}

func TestRunCmdArgs_InheritEnv_BareFlag(t *testing.T) {
	// A name in InheritEnv but not in Env should be emitted as bare -e NAME.
	opts := RunOptions{
		Image:      "img",
		InheritEnv: []string{"HOME", "PATH"},
	}
	args := runCmdArgs(opts)

	foundHome, foundPath := false, false
	for i, a := range args {
		if a == "-e" && i+1 < len(args) {
			switch args[i+1] {
			case "HOME":
				foundHome = true
			case "PATH":
				foundPath = true
			}
		}
	}
	if !foundHome {
		t.Errorf("args missing bare -e HOME: %v", args)
	}
	if !foundPath {
		t.Errorf("args missing bare -e PATH: %v", args)
	}
}

func TestRunCmdArgs_InheritEnv_SkipsDuplicates(t *testing.T) {
	// A name present in both Env and InheritEnv must not be emitted twice.
	opts := RunOptions{
		Image:      "img",
		Env:        map[string]string{"HOME": "/root"},
		InheritEnv: []string{"HOME"},
	}
	args := runCmdArgs(opts)

	count := 0
	for i, a := range args {
		if a == "-e" && i+1 < len(args) && args[i+1] == "HOME=/root" {
			count++
		}
		if a == "-e" && i+1 < len(args) && args[i+1] == "HOME" {
			t.Errorf("bare -e HOME emitted when HOME already in Env: %v", args)
		}
	}
	if count != 1 {
		t.Errorf("-e HOME=/root appears %d times, want exactly 1: %v", count, args)
	}
}

func TestRunCmdArgs_Mounts_ReadWrite(t *testing.T) {
	opts := RunOptions{
		Image: "img",
		Mounts: []Mount{
			{Source: "/host/path", Target: "/container/path", ReadOnly: false},
		},
	}
	args := runCmdArgs(opts)

	found := false
	for i, a := range args {
		if a == "-v" && i+1 < len(args) && args[i+1] == "/host/path:/container/path" {
			found = true
		}
	}
	if !found {
		t.Errorf("args missing -v /host/path:/container/path: %v", args)
	}
}

func TestRunCmdArgs_Mounts_ReadOnly(t *testing.T) {
	opts := RunOptions{
		Image: "img",
		Mounts: []Mount{
			{Source: "/host/keys", Target: "/root/.ssh", ReadOnly: true},
		},
	}
	args := runCmdArgs(opts)

	found := false
	for i, a := range args {
		if a == "-v" && i+1 < len(args) && args[i+1] == "/host/keys:/root/.ssh:ro" {
			found = true
		}
	}
	if !found {
		t.Errorf("args missing -v /host/keys:/root/.ssh:ro: %v", args)
	}
}

func TestRunCmdArgs_Mounts_Multiple(t *testing.T) {
	opts := RunOptions{
		Image: "img",
		Mounts: []Mount{
			{Source: "/a", Target: "/b", ReadOnly: false},
			{Source: "/c", Target: "/d", ReadOnly: true},
		},
	}
	args := runCmdArgs(opts)

	foundRW, foundRO := false, false
	for i, a := range args {
		if a == "-v" && i+1 < len(args) {
			switch args[i+1] {
			case "/a:/b":
				foundRW = true
			case "/c:/d:ro":
				foundRO = true
			}
		}
	}
	if !foundRW {
		t.Errorf("args missing -v /a:/b: %v", args)
	}
	if !foundRO {
		t.Errorf("args missing -v /c:/d:ro: %v", args)
	}
}

func TestRunCmdArgs_NoMounts(t *testing.T) {
	opts := RunOptions{Image: "img"}
	args := runCmdArgs(opts)
	for i, a := range args {
		if a == "-v" {
			t.Errorf("-v should not be present when Mounts is empty, found at %d", i)
		}
	}
}

func TestRunCmdArgs_NoInheritEnv(t *testing.T) {
	// With no InheritEnv, only Env entries appear.
	opts := RunOptions{Image: "img", Env: map[string]string{"FOO": "bar"}}
	args := runCmdArgs(opts)
	// Check no bare -e NAME (without =) for FOO.
	for i, a := range args {
		if a == "-e" && i+1 < len(args) && args[i+1] == "FOO" {
			t.Errorf("bare -e FOO should not appear when FOO is in Env: %v", args)
		}
	}
}

func TestMount_Struct(t *testing.T) {
	m := Mount{Source: "/src", Target: "/tgt", ReadOnly: true}
	if m.Source != "/src" {
		t.Errorf("Source: got %q, want %q", m.Source, "/src")
	}
	if m.Target != "/tgt" {
		t.Errorf("Target: got %q, want %q", m.Target, "/tgt")
	}
	if !m.ReadOnly {
		t.Error("ReadOnly: got false, want true")
	}
}

// dockerAvailable reports whether the Docker daemon is reachable.
func dockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func TestDockerRunner_Preflight_Available(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	r := &DockerRunner{}
	err := r.Preflight(context.Background())
	if err != nil {
		t.Errorf("Preflight failed with Docker available: %v", err)
	}
}

func TestDockerRunner_Preflight_ContextCancelled(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := &DockerRunner{}
	err := r.Preflight(ctx)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
	if !errors.Is(err, ErrDockerUnavailable) {
		t.Errorf("got %v, want ErrDockerUnavailable", err)
	}
}

func TestDockerRunner_Build_InvalidDir(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	r := &DockerRunner{}
	err := r.Build(context.Background(), "cldpd-test-build-invalid", "/nonexistent/path/that/does/not/exist", nil)
	if err == nil {
		t.Error("expected error building from nonexistent dir, got nil")
	}
	if !errors.Is(err, ErrBuildFailed) {
		t.Errorf("got %v, want ErrBuildFailed", err)
	}
}

func TestDockerRunner_Run_HelloWorld(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	var buf bytes.Buffer
	r := &DockerRunner{}
	opts := RunOptions{
		Image:  "alpine:latest",
		Name:   "cldpd-test-unit-run-hello",
		Cmd:    []string{"echo", "hello"},
		Remove: true,
	}
	code, err := r.Run(context.Background(), opts, &buf)
	// Clean up in case --rm didn't fire.
	exec.Command("docker", "rm", "-f", "cldpd-test-unit-run-hello").Run() //nolint:errcheck
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
}

func TestDockerRunner_Run_NonZeroExit(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	r := &DockerRunner{}
	opts := RunOptions{
		Image:  "alpine:latest",
		Name:   "cldpd-test-unit-run-exit2",
		Cmd:    []string{"sh", "-c", "exit 2"},
		Remove: true,
	}
	code, err := r.Run(context.Background(), opts, io.Discard)
	exec.Command("docker", "rm", "-f", "cldpd-test-unit-run-exit2").Run() //nolint:errcheck
	if err != nil {
		t.Fatalf("unexpected process error: %v", err)
	}
	if code != 2 {
		t.Errorf("exit code: got %d, want 2", code)
	}
}

func TestDockerRunner_Exec_ContainerNotFound(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	r := &DockerRunner{}
	_, err := r.Exec(context.Background(), "cldpd-test-unit-nonexistent", []string{"echo", "hi"}, io.Discard)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}

func TestDockerRunner_Exec_ContainerNotRunning(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	// Create a stopped container.
	containerName := "cldpd-test-unit-stopped"
	// Run and let it exit immediately.
	create := exec.Command("docker", "run", "--name", containerName, "alpine:latest", "true")
	create.Stdout = io.Discard
	create.Stderr = io.Discard
	if err := create.Run(); err != nil {
		t.Skipf("could not create stopped container: %v", err)
	}
	defer exec.Command("docker", "rm", "-f", containerName).Run() //nolint:errcheck

	r := &DockerRunner{}
	_, err := r.Exec(context.Background(), containerName, []string{"echo", "hi"}, io.Discard)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}

func TestDockerRunner_Run_WithEnvAndWorkdir(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	var buf bytes.Buffer
	r := &DockerRunner{}
	opts := RunOptions{
		Image:   "alpine:latest",
		Name:    "cldpd-test-unit-run-env",
		Cmd:     []string{"sh", "-c", "echo $TESTVAR"},
		Env:     map[string]string{"TESTVAR": "hello-env"},
		Workdir: "/tmp",
		Remove:  true,
	}
	code, err := r.Run(context.Background(), opts, &buf)
	exec.Command("docker", "rm", "-f", "cldpd-test-unit-run-env").Run() //nolint:errcheck
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected output from env var, got empty")
	}
}

func TestDockerRunner_Stop_RunningContainer(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	// Start a long-running container in detached mode.
	containerName := "cldpd-test-unit-stop-running"
	start := exec.Command("docker", "run", "-d", "--name", containerName, "alpine:latest", "sleep", "60")
	start.Stdout = io.Discard
	start.Stderr = io.Discard
	if err := start.Run(); err != nil {
		t.Skipf("could not start container: %v", err)
	}
	defer exec.Command("docker", "rm", "-f", containerName).Run() //nolint:errcheck

	r := &DockerRunner{}
	err := r.Stop(context.Background(), containerName, 5*time.Second)
	if err != nil {
		t.Errorf("Stop running container: got %v, want nil", err)
	}

	// Verify container is stopped (not running). docker stop doesn't remove — it stops.
	out, inspectErr := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", containerName).Output()
	if inspectErr != nil {
		// Container was removed (e.g. by --rm on start) — also acceptable.
		return
	}
	if strings.TrimSpace(string(out)) != "false" {
		t.Errorf("container still running after Stop; State.Running = %q", strings.TrimSpace(string(out)))
	}
}

func TestDockerRunner_Stop_NoSuchContainer(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	// Stopping a nonexistent container must return nil, not ErrStopFailed.
	r := &DockerRunner{}
	err := r.Stop(context.Background(), "cldpd-test-unit-stop-nonexistent", 5*time.Second)
	if err != nil {
		t.Errorf("Stop nonexistent container: got %v, want nil", err)
	}
}

func TestDockerRunner_Stop_ZeroTimeout_UsesFloor(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	// A zero timeout must be clamped to 1 second minimum.
	// We test this by stopping a nonexistent container with 0 timeout —
	// the call must complete (not hang) and return nil.
	r := &DockerRunner{}
	err := r.Stop(context.Background(), "cldpd-test-unit-stop-zero-timeout", 0)
	if err != nil {
		t.Errorf("Stop with zero timeout: got %v, want nil (nonexistent container)", err)
	}
}

func TestDockerRunner_Stop_ContextCancelled(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("Docker not available")
	}
	// A pre-cancelled context causes exec.CommandContext to fail before the
	// process starts, returning a non-ExitError. Stop must return ErrStopFailed.
	r := &DockerRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Stop(ctx, "cldpd-test-unit-stop-cancelled", 10*time.Second)
	if !errors.Is(err, ErrStopFailed) {
		t.Errorf("Stop with cancelled context: got %v, want ErrStopFailed", err)
	}
}
