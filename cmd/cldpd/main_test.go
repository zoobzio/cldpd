//go:build testing

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
