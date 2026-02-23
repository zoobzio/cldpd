//go:build testing

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zoobzio/cldpd"
)

// testRunner implements cldpd.Runner for use in CLI tests.
type testRunner struct {
	preflightFn func(ctx context.Context) error
	buildFn     func(ctx context.Context, tag string, dir string, buildArgs map[string]string) error
	runFn       func(ctx context.Context, opts cldpd.RunOptions, stdout io.Writer) (int, error)
	execFn      func(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error)
	stopFn      func(ctx context.Context, container string, timeout time.Duration) error
}

func (r *testRunner) Preflight(ctx context.Context) error {
	if r.preflightFn != nil {
		return r.preflightFn(ctx)
	}
	return nil
}

func (r *testRunner) Build(ctx context.Context, tag string, dir string, buildArgs map[string]string) error {
	if r.buildFn != nil {
		return r.buildFn(ctx, tag, dir, buildArgs)
	}
	return nil
}

func (r *testRunner) Run(ctx context.Context, opts cldpd.RunOptions, stdout io.Writer) (int, error) {
	if r.runFn != nil {
		return r.runFn(ctx, opts, stdout)
	}
	return 0, nil
}

func (r *testRunner) Exec(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error) {
	if r.execFn != nil {
		return r.execFn(ctx, container, cmd, stdout)
	}
	return 0, nil
}

func (r *testRunner) Stop(ctx context.Context, container string, timeout time.Duration) error {
	if r.stopFn != nil {
		return r.stopFn(ctx, container, timeout)
	}
	return nil
}

// makeSessionPod creates a minimal valid pod directory and returns a Dispatcher backed by runner.
func makeSessionPod(t *testing.T, runner cldpd.Runner) (*cldpd.Dispatcher, string) {
	t.Helper()
	podsDir := t.TempDir()
	dir := filepath.Join(podsDir, "testpod")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create pod dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	return cldpd.NewDispatcher(podsDir, runner), "testpod"
}

// buildCLI compiles the cldpd binary into a temp dir and returns the path.
// The binary is removed when the test ends.
func buildCLI(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "cldpd")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/zoobzio/cldpd/cmd/cldpd")
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	return bin
}

// runCLI executes the binary with args and returns stdout, stderr, and exit code.
func runCLI(t *testing.T, bin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run CLI: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), code
}

func TestCLI_NoArgs(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin)
	if code != 1 {
		t.Errorf("exit code: got %d, want 1", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("stderr should contain usage, got: %q", stderr)
	}
}

func TestCLI_UnknownSubcommand(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "launch")
	if code != 1 {
		t.Errorf("exit code: got %d, want 1", code)
	}
	if !strings.Contains(stderr, "unknown subcommand") {
		t.Errorf("stderr should mention unknown subcommand, got: %q", stderr)
	}
}

func TestCLI_Start_MissingPodName(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "start", "--issue", "https://github.com/org/repo/issues/1")
	if code != 1 {
		t.Errorf("exit code: got %d, want 1", code)
	}
	if !strings.Contains(stderr, "pod name required") {
		t.Errorf("stderr should mention pod name required, got: %q", stderr)
	}
}

func TestCLI_Start_MissingIssueFlag(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "start", "myrepo")
	if code != 1 {
		t.Errorf("exit code: got %d, want 1", code)
	}
	if !strings.Contains(stderr, "--issue is required") {
		t.Errorf("stderr should mention --issue required, got: %q", stderr)
	}
}

func TestCLI_Resume_MissingPodName(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "resume", "--prompt", "do more")
	if code != 1 {
		t.Errorf("exit code: got %d, want 1", code)
	}
	if !strings.Contains(stderr, "pod name required") {
		t.Errorf("stderr should mention pod name required, got: %q", stderr)
	}
}

func TestCLI_Resume_MissingPromptFlag(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "resume", "myrepo")
	if code != 1 {
		t.Errorf("exit code: got %d, want 1", code)
	}
	if !strings.Contains(stderr, "--prompt is required") {
		t.Errorf("stderr should mention --prompt required, got: %q", stderr)
	}
}

// TestRunStart_MissingArgs tests runStart directly (same package).
func TestRunStart_MissingArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"no issue flag", []string{"myrepo"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Redirect stderr to discard for clean test output
			old := os.Stderr
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() { os.Stderr = old }()

			code := runStart(context.Background(), tc.args)
			if code != 1 {
				t.Errorf("exit code: got %d, want 1", code)
			}
		})
	}
}

// TestRunResume_MissingArgs tests runResume directly (same package).
func TestRunResume_MissingArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"no prompt flag", []string{"myrepo"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := os.Stderr
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() { os.Stderr = old }()

			code := runResume(context.Background(), tc.args)
			if code != 1 {
				t.Errorf("exit code: got %d, want 1", code)
			}
		})
	}
}

// TestRunStart_ErrorsGoToStderr verifies errors are written to stderr, not stdout.
func TestRunStart_ErrorsGoToStderr(t *testing.T) {
	bin := buildCLI(t)
	stdout, stderr, code := runCLI(t, bin, "start", "myrepo", "--issue", "https://github.com/org/repo/issues/1")
	// Docker preflight will fail in test environment without Docker.
	// What matters: stdout is empty, stderr has the error.
	if code == 0 {
		t.Skip("Docker available — skipping error path test")
	}
	if stdout != "" {
		t.Errorf("errors should not appear on stdout, got: %q", stdout)
	}
	if stderr == "" {
		t.Error("error should appear on stderr")
	}
}

// TestRunStart_PodNotFound exercises the path through Preflight, DefaultPodsDir,
// and d.Start — all of which succeed or fail gracefully with a nonexistent pod.
func TestRunStart_PodNotFound(t *testing.T) {
	// Redirect stderr to suppress noise.
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer devnull.Close()
	old := os.Stderr
	os.Stderr = devnull
	defer func() { os.Stderr = old }()

	// Flags must come before positional args for flag.Parse to see them.
	// If Docker is unavailable, runStart exits at Preflight with code 1.
	// If Docker is available, it exits at d.Start (pod not found) with code 1.
	// Either way, code must be non-zero.
	code := runStart(context.Background(), []string{"--issue", "https://github.com/org/repo/issues/1", "__nonexistent_test_pod__"})
	if code == 0 {
		t.Errorf("exit code: got 0, want non-zero")
	}
}

// TestRunResume_SessionNotFound exercises the path through DefaultPodsDir and
// d.Resume with a nonexistent container.
func TestRunResume_SessionNotFound(t *testing.T) {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer devnull.Close()
	old := os.Stderr
	os.Stderr = devnull
	defer func() { os.Stderr = old }()

	// No running container named cldpd-__nonexistent__ — Resume returns non-zero.
	code := runResume(context.Background(), []string{"--prompt", "do something", "__nonexistent_test_pod__"})
	// code may be -1 or 1 depending on the error; either way, non-zero.
	if code == 0 {
		t.Errorf("exit code: got 0, want non-zero")
	}
}

// TestCLI_Help verifies that the help subcommand exits 0 and prints usage.
func TestCLI_Help(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "help")
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("stderr should contain usage, got: %q", stderr)
	}
}

// TestCLI_HelpFlag verifies that --help exits 0 and prints usage.
func TestCLI_HelpFlag(t *testing.T) {
	bin := buildCLI(t)
	_, stderr, code := runCLI(t, bin, "--help")
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("stderr should contain usage, got: %q", stderr)
	}
}

// consumeSession tests below construct real *cldpd.Session values via Dispatcher
// backed by a testRunner, allowing in-process testing of the event loop.

func TestConsumeSession_OutputToStdout(t *testing.T) {
	r := &testRunner{
		runFn: func(_ context.Context, _ cldpd.RunOptions, stdout io.Writer) (int, error) {
			fmt.Fprintln(stdout, "output line one")
			fmt.Fprintln(stdout, "output line two")
			return 0, nil
		},
	}
	d, pod := makeSessionPod(t, r)
	session, err := d.Start(context.Background(), pod, "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Redirect stdout to capture consumeSession output.
	pr, pw, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = pw

	code := consumeSession(context.Background(), session)

	pw.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, pr) //nolint:errcheck
	pr.Close()

	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	out := buf.String()
	if !strings.Contains(out, "output line one") {
		t.Errorf("stdout missing 'output line one': %q", out)
	}
	if !strings.Contains(out, "output line two") {
		t.Errorf("stdout missing 'output line two': %q", out)
	}
}

func TestConsumeSession_ErrorToStderr(t *testing.T) {
	runErr := fmt.Errorf("container process error")
	r := &testRunner{
		runFn: func(_ context.Context, _ cldpd.RunOptions, _ io.Writer) (int, error) {
			return -1, runErr
		},
	}
	d, pod := makeSessionPod(t, r)
	session, err := d.Start(context.Background(), pod, "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Redirect stderr to capture error output.
	pr, pw, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = pw

	consumeSession(context.Background(), session)

	pw.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, pr) //nolint:errcheck
	pr.Close()

	errOut := buf.String()
	if !strings.Contains(errOut, "container process error") {
		t.Errorf("stderr missing error message: %q", errOut)
	}
}

func TestConsumeSession_ReturnsExitCode(t *testing.T) {
	r := &testRunner{
		runFn: func(_ context.Context, _ cldpd.RunOptions, _ io.Writer) (int, error) {
			return 5, nil
		},
	}
	d, pod := makeSessionPod(t, r)
	session, err := d.Start(context.Background(), pod, "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Discard stderr to suppress any noise.
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() { os.Stderr = oldStderr }()

	code := consumeSession(context.Background(), session)
	if code != 5 {
		t.Errorf("exit code: got %d, want 5", code)
	}
}

func TestConsumeSession_InterruptCallsStop(t *testing.T) {
	stopCalled := make(chan struct{})
	unblock := make(chan struct{})

	r := &testRunner{
		runFn: func(_ context.Context, _ cldpd.RunOptions, _ io.Writer) (int, error) {
			<-unblock
			return 0, nil
		},
		stopFn: func(_ context.Context, _ string, _ time.Duration) error {
			close(stopCalled)
			close(unblock)
			return nil
		},
	}
	d, pod := makeSessionPod(t, r)
	session, err := d.Start(context.Background(), pod, "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan int, 1)
	go func() {
		done <- consumeSession(ctx, session)
	}()

	// Cancel context to simulate interrupt.
	cancel()

	select {
	case <-stopCalled:
		// Stop was triggered by the interrupt goroutine.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop was not called within 2s after interrupt")
	}

	select {
	case <-done:
		// consumeSession returned.
	case <-time.After(2 * time.Second):
		t.Fatal("consumeSession did not return within 2s after Stop")
	}
}

func TestRun_Dispatch(t *testing.T) {
	// run() reads os.Args directly. We restore it after each subtest.
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Redirect stderr to suppress noise from run() calls.
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer devnull.Close()

	cases := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "no args",
			args:     []string{"cldpd"},
			wantCode: 1,
		},
		{
			name:     "unknown subcommand",
			args:     []string{"cldpd", "launch"},
			wantCode: 1,
		},
		{
			name:     "help subcommand",
			args:     []string{"cldpd", "help"},
			wantCode: 0,
		},
		{
			name:     "--help flag",
			args:     []string{"cldpd", "--help"},
			wantCode: 0,
		},
		{
			name:     "start missing pod name",
			args:     []string{"cldpd", "start", "--issue", "https://github.com/org/repo/issues/1"},
			wantCode: 1,
		},
		{
			name:     "resume missing pod name",
			args:     []string{"cldpd", "resume", "--prompt", "do something"},
			wantCode: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args
			old := os.Stderr
			os.Stderr = devnull
			code := run(context.Background())
			os.Stderr = old
			if code != tc.wantCode {
				t.Errorf("run(%v): got code %d, want %d", tc.args, code, tc.wantCode)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Redirect stderr and verify printUsage writes something containing "Usage:".
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = pw

	printUsage()

	pw.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, pr) //nolint:errcheck
	pr.Close()

	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("printUsage output missing 'Usage:': %q", buf.String())
	}
}
