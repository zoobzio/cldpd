//go:build testing

package cldpd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// makeTestPod creates a minimal valid pod directory in podsDir.
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

// drainSession drains a session's events and calls Wait, failing the test if
// either does not complete within the timeout.
func drainSession(t *testing.T, s *Session, timeout time.Duration) ([]Event, int, error) {
	t.Helper()
	events := collectEvents(t, s.Events(), timeout)
	code, err := waitForDone(t, s, timeout)
	return events, code, err
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

func TestNewSessionID_Format(t *testing.T) {
	re := regexp.MustCompile(`^myrepo-[0-9a-f]{8}$`)
	id := newSessionID("myrepo")
	if !re.MatchString(id) {
		t.Errorf("newSessionID: got %q, want format myrepo-<8 hex chars>", id)
	}
}

func TestNewSessionID_Unique(t *testing.T) {
	id1 := newSessionID("pod")
	id2 := newSessionID("pod")
	if id1 == id2 {
		t.Errorf("newSessionID: two calls returned same ID %q", id1)
	}
}

func TestDispatcher_Start_PodNotFound(t *testing.T) {
	podsDir := t.TempDir()
	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	_, err := d.Start(context.Background(), "ghost", "https://github.com/org/repo/issues/1")
	if !errors.Is(err, ErrPodNotFound) {
		t.Errorf("got %v, want ErrPodNotFound", err)
	}
}

func TestDispatcher_Start_InvalidPod(t *testing.T) {
	podsDir := t.TempDir()
	dir := filepath.Join(podsDir, "nodocker")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create dir: %v", err)
	}

	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	_, err := d.Start(context.Background(), "nodocker", "https://github.com/org/repo/issues/1")
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

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if builtTag != "cldpd-myrepo" {
		t.Errorf("image tag: got %q, want %q", builtTag, "cldpd-myrepo")
	}
}

func TestDispatcher_Start_CustomImageTag(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")
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

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

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

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if !errors.Is(err, ErrBuildFailed) {
		t.Errorf("got %v, want ErrBuildFailed", err)
	}
	if s != nil {
		t.Error("session should be nil on build failure")
		drainSession(t, s, 2*time.Second)
	}
}

func TestDispatcher_Start_RunOptions_Image(t *testing.T) {
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
	s, err := d.Start(context.Background(), "myrepo", issueURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

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

func TestDispatcher_Start_ContainerName_IsSessionID(t *testing.T) {
	// Container name is now the session ID, not cldpd-<podName>.
	// Session ID format: <podName>-<hex8>.
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	var capturedName string
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedName = opts.Name
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	re := regexp.MustCompile(`^myrepo-[0-9a-f]{8}$`)
	if !re.MatchString(capturedName) {
		t.Errorf("container name: got %q, want format myrepo-<8 hex chars>", capturedName)
	}
}

func TestDispatcher_Start_PreambleEvents(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := collectEvents(t, s.Events(), 2*time.Second)
	waitForDone(t, s, 2*time.Second)

	typeCount := make(map[EventType]int)
	for _, e := range events {
		typeCount[e.Type]++
	}
	if typeCount[EventBuildStarted] != 1 {
		t.Errorf("EventBuildStarted: got %d, want 1", typeCount[EventBuildStarted])
	}
	if typeCount[EventBuildComplete] != 1 {
		t.Errorf("EventBuildComplete: got %d, want 1", typeCount[EventBuildComplete])
	}
	if typeCount[EventContainerStarted] != 1 {
		t.Errorf("EventContainerStarted: got %d, want 1", typeCount[EventContainerStarted])
	}
	if typeCount[EventContainerExited] != 1 {
		t.Errorf("EventContainerExited: got %d, want 1", typeCount[EventContainerExited])
	}
	// BuildStarted must come before BuildComplete which must come before ContainerStarted.
	var order []EventType
	for _, e := range events {
		order = append(order, e.Type)
	}
	if order[0] != EventBuildStarted {
		t.Errorf("first event: got %d, want EventBuildStarted", order[0])
	}
	if order[1] != EventBuildComplete {
		t.Errorf("second event: got %d, want EventBuildComplete", order[1])
	}
	if order[2] != EventContainerStarted {
		t.Errorf("third event: got %d, want EventContainerStarted", order[2])
	}
}

func TestDispatcher_Start_OutputEvents(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{
		runFn: func(_ context.Context, _ RunOptions, stdout io.Writer) (int, error) {
			fmt.Fprintln(stdout, "hello from container")
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := collectEvents(t, s.Events(), 2*time.Second)
	waitForDone(t, s, 2*time.Second)

	var outputEvents []Event
	for _, e := range events {
		if e.Type == EventOutput {
			outputEvents = append(outputEvents, e)
		}
	}
	if len(outputEvents) != 1 {
		t.Fatalf("output events: got %d, want 1", len(outputEvents))
	}
	if outputEvents[0].Data != "hello from container" {
		t.Errorf("output data: got %q, want %q", outputEvents[0].Data, "hello from container")
	}
}

func TestDispatcher_Start_NonZeroExit_ViaSession(t *testing.T) {
	// Non-zero exit code is delivered through the session, not as a Start error.
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	r := &mockRunner{
		runFn: func(_ context.Context, _ RunOptions, _ io.Writer) (int, error) {
			return 2, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("Start returned unexpected error: %v", err)
	}

	events, code, waitErr := drainSession(t, s, 2*time.Second)
	if waitErr != nil {
		t.Errorf("Wait error: got %v, want nil", waitErr)
	}
	if code != 2 {
		t.Errorf("exit code: got %d, want 2", code)
	}

	var exitEvent *Event
	for i := range events {
		if events[i].Type == EventContainerExited {
			exitEvent = &events[i]
		}
	}
	if exitEvent == nil {
		t.Fatal("no ContainerExited event")
	}
	if exitEvent.Code != 2 {
		t.Errorf("ContainerExited.Code: got %d, want 2", exitEvent.Code)
	}
}

func TestDispatcher_Start_InheritEnv_MergedIntoRunOptions(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")
	dir := filepath.Join(podsDir, "myrepo")
	if err := os.WriteFile(filepath.Join(dir, "pod.json"),
		[]byte(`{"inheritEnv": ["TEST_DISPATCH_VAR"]}`), 0644); err != nil {
		t.Fatalf("write pod.json: %v", err)
	}

	t.Setenv("TEST_DISPATCH_VAR", "dispatch-value")

	var capturedOpts RunOptions
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedOpts = opts
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if capturedOpts.Env["TEST_DISPATCH_VAR"] != "dispatch-value" {
		t.Errorf("InheritEnv: TEST_DISPATCH_VAR not merged into Env: %v", capturedOpts.Env)
	}
}

func TestDispatcher_Start_InheritEnv_EmptyHostVar_NotMerged(t *testing.T) {
	// If the host env var is empty/unset, it should not appear in Env.
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")
	dir := filepath.Join(podsDir, "myrepo")
	if err := os.WriteFile(filepath.Join(dir, "pod.json"),
		[]byte(`{"inheritEnv": ["DEFINITELY_NOT_SET_XYZ123"]}`), 0644); err != nil {
		t.Fatalf("write pod.json: %v", err)
	}
	os.Unsetenv("DEFINITELY_NOT_SET_XYZ123")

	var capturedOpts RunOptions
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedOpts = opts
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if _, ok := capturedOpts.Env["DEFINITELY_NOT_SET_XYZ123"]; ok {
		t.Error("unset InheritEnv var should not appear in RunOptions.Env")
	}
}

func TestDispatcher_Start_Mounts_PassedThrough(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")
	dir := filepath.Join(podsDir, "myrepo")
	if err := os.WriteFile(filepath.Join(dir, "pod.json"),
		[]byte(`{"mounts": [{"source": "/host/keys", "target": "/root/.ssh", "readOnly": true}]}`), 0644); err != nil {
		t.Fatalf("write pod.json: %v", err)
	}

	var capturedOpts RunOptions
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedOpts = opts
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if len(capturedOpts.Mounts) != 1 {
		t.Fatalf("Mounts: got %d, want 1", len(capturedOpts.Mounts))
	}
	if capturedOpts.Mounts[0].Source != "/host/keys" {
		t.Errorf("Mount.Source: got %q, want %q", capturedOpts.Mounts[0].Source, "/host/keys")
	}
	if !capturedOpts.Mounts[0].ReadOnly {
		t.Error("Mount.ReadOnly: got false, want true")
	}
}

func TestDispatcher_Start_ConcurrentCalls_UniqueContainerNames(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	var names []string
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			names = append(names, opts.Name)
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	// Start twice sequentially; names must differ.
	s1, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("first Start: %v", err)
	}
	drainSession(t, s1, 2*time.Second)

	s2, err := d.Start(context.Background(), "myrepo", "https://github.com/org/repo/issues/1")
	if err != nil {
		t.Fatalf("second Start: %v", err)
	}
	drainSession(t, s2, 2*time.Second)

	if len(names) != 2 {
		t.Fatalf("expected 2 container names, got %d", len(names))
	}
	if names[0] == names[1] {
		t.Errorf("two Start calls produced same container name: %q", names[0])
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

	s, err := d.Resume(context.Background(), "myrepo", "do more work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

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

	s, err := d.Resume(context.Background(), "myrepo", "do more work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

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

func TestDispatcher_Resume_PreambleIsContainerStartedOnly(t *testing.T) {
	podsDir := t.TempDir()

	r := &mockRunner{}
	d := NewDispatcher(podsDir, r)

	s, err := d.Resume(context.Background(), "myrepo", "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := collectEvents(t, s.Events(), 2*time.Second)
	waitForDone(t, s, 2*time.Second)

	typeCount := make(map[EventType]int)
	for _, e := range events {
		typeCount[e.Type]++
	}
	if typeCount[EventBuildStarted] != 0 {
		t.Errorf("EventBuildStarted: got %d, want 0 (Resume does not build)", typeCount[EventBuildStarted])
	}
	if typeCount[EventBuildComplete] != 0 {
		t.Errorf("EventBuildComplete: got %d, want 0", typeCount[EventBuildComplete])
	}
	if typeCount[EventContainerStarted] != 1 {
		t.Errorf("EventContainerStarted: got %d, want 1", typeCount[EventContainerStarted])
	}
}

func TestDispatcher_Resume_ExecError_ViaSession(t *testing.T) {
	// ErrSessionNotFound from runner.Exec comes through the session event stream.
	podsDir := t.TempDir()

	r := &mockRunner{
		execFn: func(_ context.Context, container string, _ []string, _ io.Writer) (int, error) {
			return -1, fmt.Errorf("%w: %s", ErrSessionNotFound, container)
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Resume(context.Background(), "ghost", "guidance")
	if err != nil {
		t.Fatalf("Resume returned unexpected error: %v", err)
	}

	events := collectEvents(t, s.Events(), 2*time.Second)
	_, waitErr := waitForDone(t, s, 2*time.Second)

	if !errors.Is(waitErr, ErrSessionNotFound) {
		t.Errorf("Wait err: got %v, want ErrSessionNotFound", waitErr)
	}

	var errEvent *Event
	for i := range events {
		if events[i].Type == EventError {
			errEvent = &events[i]
		}
	}
	if errEvent == nil {
		t.Error("no EventError in session stream for exec failure")
	}
}

func TestDispatcher_Resume_OutputEvents(t *testing.T) {
	podsDir := t.TempDir()

	r := &mockRunner{
		execFn: func(_ context.Context, _ string, _ []string, stdout io.Writer) (int, error) {
			fmt.Fprintln(stdout, "resume output line")
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Resume(context.Background(), "myrepo", "guidance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := collectEvents(t, s.Events(), 2*time.Second)
	waitForDone(t, s, 2*time.Second)

	var outputEvents []Event
	for _, e := range events {
		if e.Type == EventOutput {
			outputEvents = append(outputEvents, e)
		}
	}
	if len(outputEvents) != 1 {
		t.Fatalf("output events: got %d, want 1", len(outputEvents))
	}
	if outputEvents[0].Data != "resume output line" {
		t.Errorf("output: got %q, want %q", outputEvents[0].Data, "resume output line")
	}
}

// makeTestPodWithTemplate creates a pod directory with a Dockerfile and a template.md.
func makeTestPodWithTemplate(t *testing.T, podsDir, name, templateContent string) {
	t.Helper()
	makeTestPod(t, podsDir, name)
	dir := filepath.Join(podsDir, name)
	if err := os.WriteFile(filepath.Join(dir, "template.md"), []byte(templateContent), 0644); err != nil {
		t.Fatalf("write template.md: %v", err)
	}
}

func TestDispatcher_Start_Prompt_WithTemplate(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPodWithTemplate(t, podsDir, "myrepo", "# Standing Orders\n\nEnsure origin is up to date.")

	var capturedCmd []string
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedCmd = opts.Cmd
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	issueURL := "https://github.com/org/repo/issues/99"
	s, err := d.Start(context.Background(), "myrepo", issueURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if len(capturedCmd) < 3 {
		t.Fatalf("Cmd too short: %v", capturedCmd)
	}
	// The prompt is the last element of the claude -p <prompt> command.
	prompt := capturedCmd[len(capturedCmd)-1]
	wantPrefix := "# Standing Orders\n\nEnsure origin is up to date."
	wantSuffix := "Work on this GitHub issue: " + issueURL
	if !strings.HasPrefix(prompt, wantPrefix) {
		t.Errorf("prompt does not start with template:\ngot:  %q\nwant prefix: %q", prompt, wantPrefix)
	}
	if !strings.HasSuffix(prompt, wantSuffix) {
		t.Errorf("prompt does not end with base prompt:\ngot:  %q\nwant suffix: %q", prompt, wantSuffix)
	}
	wantFull := wantPrefix + "\n\n" + wantSuffix
	if prompt != wantFull {
		t.Errorf("prompt:\ngot:  %q\nwant: %q", prompt, wantFull)
	}
}

func TestDispatcher_Start_Prompt_WithoutTemplate(t *testing.T) {
	podsDir := t.TempDir()
	makeTestPod(t, podsDir, "myrepo")

	var capturedCmd []string
	r := &mockRunner{
		runFn: func(_ context.Context, opts RunOptions, _ io.Writer) (int, error) {
			capturedCmd = opts.Cmd
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	issueURL := "https://github.com/org/repo/issues/7"
	s, err := d.Start(context.Background(), "myrepo", issueURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if len(capturedCmd) < 3 {
		t.Fatalf("Cmd too short: %v", capturedCmd)
	}
	prompt := capturedCmd[len(capturedCmd)-1]
	want := "Work on this GitHub issue: " + issueURL
	if prompt != want {
		t.Errorf("prompt:\ngot:  %q\nwant: %q", prompt, want)
	}
}

func TestDispatcher_Resume_Prompt_NoTemplateUsed(t *testing.T) {
	// Resume passes the caller's prompt directly; no template is applied.
	podsDir := t.TempDir()

	var capturedCmd []string
	r := &mockRunner{
		execFn: func(_ context.Context, _ string, cmd []string, _ io.Writer) (int, error) {
			capturedCmd = cmd
			return 0, nil
		},
	}
	d := NewDispatcher(podsDir, r)

	s, err := d.Resume(context.Background(), "myrepo", "continue where you left off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drainSession(t, s, 2*time.Second)

	if len(capturedCmd) < 4 {
		t.Fatalf("cmd too short: %v", capturedCmd)
	}
	prompt := capturedCmd[len(capturedCmd)-1]
	want := "continue where you left off"
	if prompt != want {
		t.Errorf("resume prompt:\ngot:  %q\nwant: %q", prompt, want)
	}
}
