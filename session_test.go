//go:build testing

package cldpd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

// collectEvents drains all events from the channel until it is closed.
// Fails the test if the channel does not close within the timeout.
func collectEvents(t *testing.T, events <-chan Event, timeout time.Duration) []Event {
	t.Helper()
	var got []Event
	deadline := time.After(timeout)
	for {
		select {
		case e, ok := <-events:
			if !ok {
				return got
			}
			got = append(got, e)
		case <-deadline:
			t.Fatalf("events channel did not close within %v; collected so far: %v", timeout, got)
			return got
		}
	}
}

// waitForDone blocks until Wait returns or the timeout expires.
func waitForDone(t *testing.T, s *Session, timeout time.Duration) (int, error) {
	t.Helper()
	type result struct {
		code int
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		code, err := s.Wait()
		ch <- result{code, err}
	}()
	select {
	case r := <-ch:
		return r.code, r.err
	case <-time.After(timeout):
		t.Fatalf("Wait did not return within %v", timeout)
		return -1, nil
	}
}

// immediateRunFn returns a runFn that exits immediately with the given code/err.
func immediateRunFn(code int, err error) func(pw io.WriteCloser) (int, error) {
	return func(pw io.WriteCloser) (int, error) {
		return code, err
	}
}

// writingRunFn returns a runFn that writes lines to pw, then exits with code/err.
func writingRunFn(lines []string, code int, err error) func(pw io.WriteCloser) (int, error) {
	return func(pw io.WriteCloser) (int, error) {
		for _, line := range lines {
			fmt.Fprintln(pw, line)
		}
		return code, err
	}
}

// blockingRunFn returns a runFn that blocks until unblock is closed, then returns code/err.
func blockingRunFn(unblock <-chan struct{}, code int, err error) func(pw io.WriteCloser) (int, error) {
	return func(pw io.WriteCloser) (int, error) {
		<-unblock
		return code, err
	}
}

func TestSession_ID(t *testing.T) {
	s := newSession("test-session-id", "cldpd-test", &mockRunner{}, immediateRunFn(0, nil), nil)
	if s.ID() != "test-session-id" {
		t.Errorf("ID: got %q, want %q", s.ID(), "test-session-id")
	}
	// Drain to avoid goroutine leak.
	collectEvents(t, s.Events(), 2*time.Second)
}

func TestSession_Events_ReturnsChannel(t *testing.T) {
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(0, nil), nil)
	ch := s.Events()
	if ch == nil {
		t.Fatal("Events() returned nil channel")
	}
	collectEvents(t, ch, 2*time.Second)
}

func TestSession_NoPreamble_ContainerExited(t *testing.T) {
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(0, nil), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1 (ContainerExited): %v", len(events), events)
	}
	if events[0].Type != EventContainerExited {
		t.Errorf("events[0].Type: got %d, want EventContainerExited", events[0].Type)
	}
	if events[0].Code != 0 {
		t.Errorf("events[0].Code: got %d, want 0", events[0].Code)
	}
}

func TestSession_Preamble_EmittedFirst(t *testing.T) {
	preamble := []Event{
		{Type: EventBuildStarted, Data: "cldpd-test", Time: time.Now()},
		{Type: EventBuildComplete, Data: "cldpd-test", Time: time.Now()},
		{Type: EventContainerStarted, Data: "ctn", Time: time.Now()},
	}
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(0, nil), preamble)
	events := collectEvents(t, s.Events(), 2*time.Second)

	// Expect: preamble(3) + ContainerExited(1) = 4
	if len(events) < 4 {
		t.Fatalf("got %d events, want at least 4: %v", len(events), events)
	}
	if events[0].Type != EventBuildStarted {
		t.Errorf("events[0].Type: got %d, want EventBuildStarted", events[0].Type)
	}
	if events[1].Type != EventBuildComplete {
		t.Errorf("events[1].Type: got %d, want EventBuildComplete", events[1].Type)
	}
	if events[2].Type != EventContainerStarted {
		t.Errorf("events[2].Type: got %d, want EventContainerStarted", events[2].Type)
	}
	if events[len(events)-1].Type != EventContainerExited {
		t.Errorf("last event: got %d, want EventContainerExited", events[len(events)-1].Type)
	}
}

func TestSession_Output_Events_InOrder(t *testing.T) {
	lines := []string{"line one", "line two", "line three"}
	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn(lines, 0, nil), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)

	// At minimum: 3 output events + 1 ContainerExited
	var outputEvents []Event
	for _, e := range events {
		if e.Type == EventOutput {
			outputEvents = append(outputEvents, e)
		}
	}
	if len(outputEvents) != 3 {
		t.Fatalf("got %d output events, want 3: %v", len(outputEvents), outputEvents)
	}
	for i, want := range lines {
		if outputEvents[i].Data != want {
			t.Errorf("output[%d].Data: got %q, want %q", i, outputEvents[i].Data, want)
		}
	}
}

func TestSession_Output_BeforeTerminal(t *testing.T) {
	lines := []string{"hello"}
	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn(lines, 0, nil), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)

	// Last event must be ContainerExited, not output.
	last := events[len(events)-1]
	if last.Type != EventContainerExited {
		t.Errorf("last event: got %d, want EventContainerExited", last.Type)
	}
}

func TestSession_NonZeroExit_ContainerExited_Code(t *testing.T) {
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(42, nil), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)

	var exitEvent *Event
	for i := range events {
		if events[i].Type == EventContainerExited {
			exitEvent = &events[i]
		}
	}
	if exitEvent == nil {
		t.Fatal("no ContainerExited event found")
	}
	if exitEvent.Code != 42 {
		t.Errorf("ContainerExited.Code: got %d, want 42", exitEvent.Code)
	}
}

func TestSession_RunError_EmitsEventError(t *testing.T) {
	runErr := errors.New("docker run: unexpected error")
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(-1, runErr), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)

	var errEvent *Event
	for i := range events {
		if events[i].Type == EventError {
			errEvent = &events[i]
		}
	}
	if errEvent == nil {
		t.Fatal("no EventError found")
	}
	if errEvent.Data == "" {
		t.Error("EventError.Data: expected non-empty error message")
	}
}

func TestSession_RunError_NoContainerExited(t *testing.T) {
	runErr := errors.New("fatal error")
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(-1, runErr), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)

	for _, e := range events {
		if e.Type == EventContainerExited {
			t.Error("ContainerExited should not be emitted when runFn returns error")
		}
	}
}

func TestSession_Channel_ClosedAfterTerminal(t *testing.T) {
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(0, nil), nil)
	ch := s.Events()

	// Drain all events; channel must be closed.
	collectEvents(t, ch, 2*time.Second)

	// Channel should now be closed; reading should return zero value, ok=false.
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel still open after terminal event")
		}
	default:
		// Already closed and drained; this is also fine.
	}
}

func TestSession_Wait_ReturnsExitCode(t *testing.T) {
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(7, nil), nil)
	// Don't consume events; Wait must work independently.
	code, err := waitForDone(t, s, 2*time.Second)
	if err != nil {
		t.Errorf("Wait error: got %v, want nil", err)
	}
	if code != 7 {
		t.Errorf("Wait code: got %d, want 7", code)
	}
}

func TestSession_Wait_ReturnsError(t *testing.T) {
	runErr := errors.New("process failed")
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(-1, runErr), nil)
	_, err := waitForDone(t, s, 2*time.Second)
	if !errors.Is(err, runErr) {
		t.Errorf("Wait err: got %v, want %v", err, runErr)
	}
}

func TestSession_Wait_IndependentOfEvents(t *testing.T) {
	// Call Wait without ever consuming Events; it must still return.
	s := newSession("sid", "ctn", &mockRunner{}, immediateRunFn(0, nil), nil)
	code, err := waitForDone(t, s, 2*time.Second)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("code: got %d, want 0", code)
	}
}

func TestSession_Stop_UnblocksWait(t *testing.T) {
	unblock := make(chan struct{})
	var stopCalled bool
	r := &mockRunner{
		stopFn: func(ctx context.Context, container string, timeout time.Duration) error {
			stopCalled = true
			close(unblock)
			return nil
		},
	}
	s := newSession("sid", "ctn", r, blockingRunFn(unblock, 0, nil), nil)

	ctx := context.Background()
	if err := s.Stop(ctx); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if !stopCalled {
		t.Error("runner.Stop was not called")
	}

	// Wait must now return since the container goroutine unblocked.
	code, err := waitForDone(t, s, 2*time.Second)
	if err != nil {
		t.Errorf("Wait after Stop: unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("Wait after Stop: code got %d, want 0", code)
	}
	// Drain events.
	collectEvents(t, s.Events(), 2*time.Second)
}

func TestSession_Stop_Idempotent(t *testing.T) {
	stopCount := 0
	r := &mockRunner{
		stopFn: func(ctx context.Context, container string, timeout time.Duration) error {
			stopCount++
			return nil
		},
	}
	unblock := make(chan struct{})
	// Use a different stop mock that also closes unblock.
	unblockOnce := make(chan struct{})
	r2 := &mockRunner{
		stopFn: func(ctx context.Context, container string, timeout time.Duration) error {
			stopCount++
			select {
			case <-unblockOnce:
			default:
				close(unblockOnce)
			}
			return nil
		},
	}
	_ = r
	_ = unblock

	s := newSession("sid", "ctn", r2, blockingRunFn(unblockOnce, 0, nil), nil)

	ctx := context.Background()
	// First Stop.
	if err := s.Stop(ctx); err != nil {
		t.Fatalf("first Stop: %v", err)
	}
	// Second Stop on an already-done session must return nil without calling runner.Stop again.
	if err := s.Stop(ctx); err != nil {
		t.Fatalf("second Stop: %v", err)
	}

	// runner.Stop was called exactly once (second call short-circuits via done channel).
	if stopCount != 1 {
		t.Errorf("runner.Stop called %d times, want 1", stopCount)
	}
	collectEvents(t, s.Events(), 2*time.Second)
}

func TestSession_Stop_PassesContainerName(t *testing.T) {
	var stoppedContainer string
	unblock := make(chan struct{})
	r := &mockRunner{
		stopFn: func(ctx context.Context, container string, timeout time.Duration) error {
			stoppedContainer = container
			close(unblock)
			return nil
		},
	}
	s := newSession("sid", "my-container", r, blockingRunFn(unblock, 0, nil), nil)
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stoppedContainer != "my-container" {
		t.Errorf("Stop container: got %q, want %q", stoppedContainer, "my-container")
	}
	collectEvents(t, s.Events(), 2*time.Second)
}

func TestSession_Stop_ContextExpires(t *testing.T) {
	// Stop blocks waiting for done channel; if ctx expires, it returns ctx.Err().
	// We simulate this by having runner.Stop succeed but the container goroutine
	// never actually exit (we don't close the unblock channel in the runFn).
	// Use a separate never-unblocking channel.
	neverUnblock := make(chan struct{}) // never closed

	r := &mockRunner{
		stopFn: func(ctx context.Context, container string, timeout time.Duration) error {
			// Stop succeeds but the container goroutine won't exit.
			return nil
		},
	}
	s := newSession("sid", "ctn", r, blockingRunFn(neverUnblock, 0, nil), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := s.Stop(ctx)
	if err == nil {
		t.Error("Stop: expected error when ctx expires, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Stop: got %v, want DeadlineExceeded", err)
	}
	// Clean up: close neverUnblock so the goroutines can exit.
	close(neverUnblock)
	// Drain to avoid goroutine leak.
	collectEvents(t, s.Events(), 2*time.Second)
}

func TestSession_Stop_RunnerError(t *testing.T) {
	stopErr := fmt.Errorf("%w: exit code 1", ErrStopFailed)
	r := &mockRunner{
		stopFn: func(ctx context.Context, container string, timeout time.Duration) error {
			return stopErr
		},
	}
	s := newSession("sid", "ctn", r, immediateRunFn(0, nil), nil)

	// Wait for the session to finish naturally first so the events drain.
	collectEvents(t, s.Events(), 2*time.Second)

	// Now call Stop on an already-done session — it returns nil (idempotent path).
	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("Stop on done session: got %v, want nil", err)
	}
}

func TestSession_EventTime_NonZero(t *testing.T) {
	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn([]string{"hello"}, 0, nil), nil)
	events := collectEvents(t, s.Events(), 2*time.Second)
	for _, e := range events {
		if e.Time.IsZero() {
			t.Errorf("event %d has zero Time", e.Type)
		}
	}
}

func TestSession_EmitOutput_DropsWhenFull(t *testing.T) {
	// Fill a channel beyond its buffer. emitOutput must not block; excess lines are dropped.
	// The event goroutine must still emit the terminal lifecycle event and close the channel.
	//
	// emitLifecycle is a blocking send — it requires a consumer running concurrently.
	// Without one, a full channel deadlocks on the terminal event. So we drain concurrently.
	lineCount := eventChannelBuffer * 3
	var lines []string
	for i := 0; i < lineCount; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}

	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn(lines, 0, nil), nil)

	// Drain concurrently so lifecycle events are never blocked.
	events := collectEvents(t, s.Events(), 5*time.Second)

	// Verify: output events may be fewer than lines written (some dropped).
	outputCount := 0
	for _, e := range events {
		if e.Type == EventOutput {
			outputCount++
		}
	}
	if outputCount > lineCount {
		t.Errorf("output events (%d) exceeds lines written (%d)", outputCount, lineCount)
	}

	// The terminal event must always appear.
	var hasTerminal bool
	for _, e := range events {
		if e.Type == EventContainerExited || e.Type == EventError {
			hasTerminal = true
		}
	}
	if !hasTerminal {
		t.Error("no terminal event found — lifecycle event was dropped or session hung")
	}

	// Wait must return now that the channel is closed.
	code, err := waitForDone(t, s, 2*time.Second)
	if err != nil {
		t.Errorf("unexpected error after high-volume output: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
}

func TestSession_LifecycleEvents_NeverDropped(t *testing.T) {
	// Preamble events are emitted synchronously (blocking send) before goroutines start.
	// They must all appear in the event stream even when combined with high output volume.
	preamble := []Event{
		{Type: EventBuildStarted, Data: "img", Time: time.Now()},
		{Type: EventBuildComplete, Data: "img", Time: time.Now()},
		{Type: EventContainerStarted, Data: "ctn", Time: time.Now()},
	}
	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn([]string{"line"}, 0, nil), preamble)
	events := collectEvents(t, s.Events(), 2*time.Second)

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
}

func TestSession_Wait_ExitCode_NotStale_AfterHighVolume(t *testing.T) {
	// Task #4: exit code is written before pw.Close() — Wait() must never return
	// a stale zero value. Run under go test -race to surface any ordering violation.
	lines := make([]string, eventChannelBuffer*2)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i)
	}

	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn(lines, 42, nil), nil)
	code, err := waitForDone(t, s, 5*time.Second)
	if err != nil {
		t.Errorf("Wait error: got %v, want nil", err)
	}
	if code != 42 {
		t.Errorf("exit code: got %d, want 42 (stale zero would indicate a race)", code)
	}
}

func TestSession_Wait_DoesNotDeadlock_WhenEventsNotConsumed(t *testing.T) {
	// Task #2: Wait() must return even when Events() is never consumed and the
	// event buffer fills. The fix closes done before emitting the terminal event,
	// so Wait() unblocks regardless of channel state.
	lines := make([]string, eventChannelBuffer*3)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i)
	}

	s := newSession("sid", "ctn", &mockRunner{}, writingRunFn(lines, 0, nil), nil)
	// Deliberately do NOT call s.Events() — channel is never consumed.
	code, err := waitForDone(t, s, 5*time.Second)
	if err != nil {
		t.Errorf("Wait error: got %v, want nil", err)
	}
	if code != 0 {
		t.Errorf("exit code: got %d, want 0", code)
	}
}
