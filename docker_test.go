//go:build testing

package cldpd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

// mockRunner is a test double for Runner.
type mockRunner struct {
	buildFn func(ctx context.Context, tag string, dir string, buildArgs map[string]string) error
	runFn   func(ctx context.Context, opts RunOptions, stdout io.Writer) (int, error)
	execFn  func(ctx context.Context, container string, cmd []string, stdout io.Writer) (int, error)
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

// Verify mockRunner satisfies Runner interface at compile time.
var _ Runner = (*mockRunner)(nil)

// Verify DockerRunner satisfies Runner interface at compile time.
var _ Runner = (*DockerRunner)(nil)

func TestRunner_InterfaceSatisfied(t *testing.T) {
	// Compile-time assertions above; this test documents the requirement.
	var _ Runner = (*DockerRunner)(nil)
	var _ Runner = (*mockRunner)(nil)
}

func TestMockRunner_Build_Success(t *testing.T) {
	called := false
	r := &mockRunner{
		buildFn: func(ctx context.Context, tag string, dir string, buildArgs map[string]string) error {
			called = true
			if tag != "myimage" {
				t.Errorf("tag: got %q, want %q", tag, "myimage")
			}
			if dir != "/some/dir" {
				t.Errorf("dir: got %q, want %q", dir, "/some/dir")
			}
			return nil
		},
	}
	err := r.Build(context.Background(), "myimage", "/some/dir", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("Build was not called")
	}
}

func TestMockRunner_Build_Error(t *testing.T) {
	r := &mockRunner{
		buildFn: func(_ context.Context, _ string, _ string, _ map[string]string) error {
			return ErrBuildFailed
		},
	}
	err := r.Build(context.Background(), "tag", "/dir", nil)
	if !errors.Is(err, ErrBuildFailed) {
		t.Errorf("got %v, want ErrBuildFailed", err)
	}
}

func TestMockRunner_Run_PassesThroughOutput(t *testing.T) {
	var buf bytes.Buffer
	r := &mockRunner{
		runFn: func(_ context.Context, _ RunOptions, stdout io.Writer) (int, error) {
			_, _ = stdout.Write([]byte("hello from container"))
			return 0, nil
		},
	}
	code, err := r.Run(context.Background(), RunOptions{Image: "img", Name: "c"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	if buf.String() != "hello from container" {
		t.Errorf("output: got %q, want %q", buf.String(), "hello from container")
	}
}

func TestMockRunner_Run_NonZeroExitCode(t *testing.T) {
	r := &mockRunner{
		runFn: func(_ context.Context, _ RunOptions, _ io.Writer) (int, error) {
			return 2, nil
		},
	}
	code, err := r.Run(context.Background(), RunOptions{Image: "img"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 2 {
		t.Errorf("exit code: got %d, want 2", code)
	}
}

func TestMockRunner_Exec_PassesThroughOutput(t *testing.T) {
	var buf bytes.Buffer
	r := &mockRunner{
		execFn: func(_ context.Context, container string, cmd []string, stdout io.Writer) (int, error) {
			if container != "cldpd-myrepo" {
				t.Errorf("container: got %q, want %q", container, "cldpd-myrepo")
			}
			_, _ = stdout.Write([]byte("exec output"))
			return 0, nil
		},
	}
	code, err := r.Exec(context.Background(), "cldpd-myrepo", []string{"echo", "hi"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
	if buf.String() != "exec output" {
		t.Errorf("output: got %q, want %q", buf.String(), "exec output")
	}
}

func TestMockRunner_Exec_SessionNotFound(t *testing.T) {
	r := &mockRunner{
		execFn: func(_ context.Context, container string, _ []string, _ io.Writer) (int, error) {
			return -1, ErrSessionNotFound
		},
	}
	_, err := r.Exec(context.Background(), "cldpd-missing", []string{"claude"}, io.Discard)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}

func TestRunOptions_Fields(t *testing.T) {
	opts := RunOptions{
		Image:   "myimage:latest",
		Name:    "cldpd-myrepo",
		Cmd:     []string{"claude", "-p", "prompt"},
		Env:     map[string]string{"KEY": "val"},
		Workdir: "/workspace",
		Remove:  true,
	}
	if opts.Image != "myimage:latest" {
		t.Errorf("Image: got %q", opts.Image)
	}
	if opts.Name != "cldpd-myrepo" {
		t.Errorf("Name: got %q", opts.Name)
	}
	if len(opts.Cmd) != 3 {
		t.Errorf("Cmd len: got %d, want 3", len(opts.Cmd))
	}
	if opts.Env["KEY"] != "val" {
		t.Errorf("Env[KEY]: got %q", opts.Env["KEY"])
	}
	if opts.Workdir != "/workspace" {
		t.Errorf("Workdir: got %q", opts.Workdir)
	}
	if !opts.Remove {
		t.Error("Remove: got false, want true")
	}
}
