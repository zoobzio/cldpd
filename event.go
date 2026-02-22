package cldpd

import "time"

// EventType identifies the kind of event emitted by a Session.
type EventType int

const (
	// EventBuildStarted is emitted when the Docker image build begins.
	// Data contains the image tag.
	EventBuildStarted EventType = iota

	// EventBuildComplete is emitted when the Docker image build succeeds.
	// Data contains the image tag.
	EventBuildComplete

	// EventContainerStarted is emitted when the container begins running.
	// Data contains the container name.
	EventContainerStarted

	// EventOutput is emitted for each line of container stdout.
	// Data contains the line content.
	EventOutput

	// EventContainerExited is emitted when the container exits normally.
	// Code contains the container's exit code.
	EventContainerExited

	// EventError is emitted when a fatal error terminates the session.
	// Data contains the error message.
	EventError
)

// Event is a lifecycle or output event emitted by a Session.
//
// Temporal ordering guarantees:
//   - Successful start: BuildStarted → BuildComplete → ContainerStarted → Output* → ContainerExited
//   - Build failure:    BuildStarted → Error
//   - Runtime failure:  BuildStarted → BuildComplete → ContainerStarted → Output* → Error
//
// After the terminal event (ContainerExited or Error), the channel is closed.
type Event struct {
	Time time.Time
	Data string
	Type EventType
	Code int
}
