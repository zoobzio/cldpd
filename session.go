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

// newSession creates a Session and starts its goroutines.
//
// The goroutine sequence:
//  1. container goroutine: calls runFn, writes exitCode/exitErr under mutex, closes pipeWriter.
//  2. event goroutine: reads lines from pipeReader, emits EventOutput, closes done, then emits terminal event.
//
// done is closed before the terminal event is emitted, so Wait() never blocks on
// event consumption. preamble events are emitted synchronously before goroutines start.
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
		// Write results under mutex before closing the pipe. Closing pw signals
		// EOF to the event goroutine; by writing first, we guarantee the event
		// goroutine observes committed values when it reads after EOF.
		s.mu.Lock()
		s.exitCode = code
		s.exitErr = err
		s.mu.Unlock()
		// PipeWriter.Close always returns nil, but the error is checked to satisfy errcheck.
		_ = pw.Close()
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

		// Read the result stored by the container goroutine. EOF guarantees the
		// container goroutine has already committed exitCode/exitErr under its mutex.
		s.mu.Lock()
		code := s.exitCode
		err := s.exitErr
		s.mu.Unlock()

		// Signal Wait BEFORE emitting the terminal event. This ensures Wait()
		// never deadlocks even if the event channel is full.
		s.once.Do(func() { close(s.done) })

		// Emit terminal event with a non-blocking send. If the channel is full,
		// the event is lost, but Wait() has already been unblocked. Callers who
		// consume Events() will see the channel close as the terminal signal.
		var terminal Event
		if err != nil {
			terminal = Event{
				Type: EventError,
				Data: err.Error(),
				Time: time.Now(),
			}
		} else {
			terminal = Event{
				Type: EventContainerExited,
				Code: code,
				Time: time.Now(),
			}
		}
		select {
		case s.events <- terminal:
		default:
		}

		close(s.events)
	}()

	return s
}

// emitLifecycle sends a lifecycle event to the channel, blocking until delivered.
// Used only for preamble events emitted synchronously before goroutines start,
// when the channel buffer is empty and blocking is safe.
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
// Consuming Events() is optional. Wait() returns as soon as the container exits,
// independent of whether Events() is consumed. Under high output volume, output
// events may be dropped if the buffer fills; the terminal event may also be
// dropped if the buffer is full when the container exits, but the channel is
// always closed as the definitive terminal signal.
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

	// Wait for the event goroutine to finish (done channel closes, then terminal
	// event emitted, then events channel closed).
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
