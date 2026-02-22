//go:build testing

package cldpd

import (
	"testing"
	"time"
)

func TestEventType_Constants(t *testing.T) {
	// All six event types must be distinct.
	types := []EventType{
		EventBuildStarted,
		EventBuildComplete,
		EventContainerStarted,
		EventOutput,
		EventContainerExited,
		EventError,
	}
	seen := make(map[EventType]bool)
	for _, et := range types {
		if seen[et] {
			t.Errorf("duplicate EventType value: %d", et)
		}
		seen[et] = true
	}
}

func TestEvent_Fields(t *testing.T) {
	now := time.Now()
	e := Event{
		Type: EventOutput,
		Data: "hello from container",
		Code: 0,
		Time: now,
	}
	if e.Type != EventOutput {
		t.Errorf("Type: got %d, want %d", e.Type, EventOutput)
	}
	if e.Data != "hello from container" {
		t.Errorf("Data: got %q, want %q", e.Data, "hello from container")
	}
	if e.Code != 0 {
		t.Errorf("Code: got %d, want 0", e.Code)
	}
	if !e.Time.Equal(now) {
		t.Errorf("Time: got %v, want %v", e.Time, now)
	}
}

func TestEvent_ExitCode(t *testing.T) {
	e := Event{
		Type: EventContainerExited,
		Code: 137,
	}
	if e.Code != 137 {
		t.Errorf("Code: got %d, want 137", e.Code)
	}
}

func TestEvent_ErrorType(t *testing.T) {
	e := Event{
		Type: EventError,
		Data: "build failed: exit code 1",
	}
	if e.Type != EventError {
		t.Errorf("Type: got %d, want EventError", e.Type)
	}
	if e.Data == "" {
		t.Error("Data: expected non-empty error message")
	}
}

func TestEventType_BuildSequence(t *testing.T) {
	// Verify the documented ordering values make sense for iota assignment.
	// BuildStarted must be first (zero value).
	if EventBuildStarted != 0 {
		t.Errorf("EventBuildStarted: got %d, want 0", EventBuildStarted)
	}
	// Each subsequent value must be greater than the previous.
	sequence := []EventType{
		EventBuildStarted,
		EventBuildComplete,
		EventContainerStarted,
		EventOutput,
		EventContainerExited,
		EventError,
	}
	for i := 1; i < len(sequence); i++ {
		if sequence[i] <= sequence[i-1] {
			t.Errorf("EventType sequence broken at index %d: %d <= %d", i, sequence[i], sequence[i-1])
		}
	}
}
