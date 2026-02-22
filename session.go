package cldpd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

const (
	// sessionStopTimeout is the default timeout passed to runner.Stop.
	sessionStopTimeout = 10 * time.Second

	// eventChannelBuffer is the size of the event channel buffer.
	// Lifecycle events block until delivered. Output events may be dropped
	// under sustained backpressure.
	eventChannelBuffer = 256
)

// Session represents an active pod lifecycle. It is returned by Dispatcher.Start
// and Dispatcher.Resume. The caller owns the Session and is responsible for
// calling Stop or Wait.
//
// Events and Wait are independent consumption paths — neither requires the other.
// Stop is idempotent.
type Session struct {
	runner    Runner
	exitErr   error
	events    chan Event
	done      chan struct{}
	id        string
	container string
	// mu guards exitCode and exitErr.
	mu       sync.Mutex
	once     sync.Once // guards done channel close
	exitCode int
}

// newSession creates a Session and starts its goroutines. pipeReader is the read
// end of a pipe connected to the container's stdout. The caller must ensure the
// container is already running (or starting) before calling newSession.
//
// The goroutine sequence:
//  1. container goroutine: calls runFn, stores result, closes pipeWriter.
//  2. event goroutine: reads lines from pipeReader, emits EventOutput, then emits terminal event.
//
// preamble events are emitted synchronously before goroutines start.
func newSession(
	id string,
	container string,
	runner Runner,
	runFn func(pw io.WriteCloser) (int, error),
	preamble []Event,
) *Session {
	s := &Session{
		id:        id,
		container: container,
		runner:    runner,
		events:    make(chan Event, eventChannelBuffer),
		done:      make(chan struct{}),
	}

	// Emit preamble lifecycle events synchronously before spawning goroutines.
	for _, e := range preamble {
		s.emitLifecycle(e)
	}

	pr, pw := io.Pipe()

	// Container goroutine: runs the container, stores result, closes the pipe.
	go func() {
		code, err := runFn(pw)
		// Close the write end of the pipe; signals EOF to the event goroutine.
		// PipeWriter.Close always returns nil, but the error is checked to satisfy errcheck.
		_ = pw.Close()
		s.mu.Lock()
		s.exitCode = code
		s.exitErr = err
		s.mu.Unlock()
	}()

	// Event goroutine: reads lines from pipeReader, emits events, then closes channel.
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			s.emitOutput(Event{
				Type: EventOutput,
				Data: scanner.Text(),
				Time: time.Now(),
			})
		}
		// pipeReader is exhausted (EOF). Pipe closure is normal termination.
		// PipeReader.Close always returns nil, but the error is checked to satisfy errcheck.
		_ = pr.Close()

		// Read the result stored by the container goroutine. The pipe is closed,
		// so the container goroutine has already written its result.
		s.mu.Lock()
		code := s.exitCode
		err := s.exitErr
		s.mu.Unlock()

		if err != nil {
			s.emitLifecycle(Event{
				Type: EventError,
				Data: err.Error(),
				Time: time.Now(),
			})
		} else {
			s.emitLifecycle(Event{
				Type: EventContainerExited,
				Code: code,
				Time: time.Now(),
			})
		}

		close(s.events)
		// Signal Wait that the session is done.
		s.once.Do(func() { close(s.done) })
	}()

	return s
}

// emitLifecycle sends a lifecycle event to the channel, blocking until delivered.
// Lifecycle events must never be dropped.
func (s *Session) emitLifecycle(e Event) {
	s.events <- e
}

// emitOutput sends an output event to the channel. If the channel is full,
// the event is dropped to avoid blocking the event goroutine indefinitely.
func (s *Session) emitOutput(e Event) {
	select {
	case s.events <- e:
	default:
		// Channel full; drop this output event.
	}
}

// ID returns the unique session identifier.
func (s *Session) ID() string {
	return s.id
}

// Events returns a receive-only channel of typed events. The channel is closed
// after the terminal event (ContainerExited or Error). Callers may range over
// this channel to consume the full event stream.
//
// Warning: if the caller never consumes Events(), the channel buffer (256) will
// fill under high output volume. Once full, the terminal lifecycle event blocks
// until space is available, which means Wait() will also block indefinitely.
// Either consume Events() or use Stop/Wait without consuming events only when
// container output is known to be low-volume.
func (s *Session) Events() <-chan Event {
	return s.events
}

// Stop initiates graceful shutdown of the container. It calls runner.Stop with
// a 10-second SIGTERM timeout, then blocks until the container goroutine exits
// or ctx expires.
//
// Stop is idempotent: calling it on an already-stopped session returns nil immediately.
func (s *Session) Stop(ctx context.Context) error {
	// If already done, return immediately.
	select {
	case <-s.done:
		return nil
	default:
	}

	if err := s.runner.Stop(ctx, s.container, sessionStopTimeout); err != nil {
		return fmt.Errorf("stop session %s: %w", s.id, err)
	}

	// Wait for the container goroutine to finish (pipe closes, event goroutine
	// emits terminal event, done channel closes).
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wait blocks until the container exits and returns its exit code and any
// process-level error. A non-zero exit code does not itself produce an error
// here — check the returned code.
//
// Wait is independent of Events: it can be called without consuming the event channel.
func (s *Session) Wait() (int, error) {
	<-s.done
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode, s.exitErr
}
