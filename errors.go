package cldpd

import "errors"

// ErrPodNotFound is returned when a pod directory does not exist.
var ErrPodNotFound = errors.New("pod not found")

// ErrInvalidPod is returned when a pod directory exists but contains no Dockerfile.
var ErrInvalidPod = errors.New("invalid pod: Dockerfile not found")

// ErrBuildFailed is returned when the Docker image build exits with a non-zero status.
var ErrBuildFailed = errors.New("image build failed")

// ErrContainerFailed is returned when a container exits with a non-zero status.
var ErrContainerFailed = errors.New("container exited with error")

// ErrSessionNotFound is returned when no running session exists for the given pod name.
var ErrSessionNotFound = errors.New("no running session for pod")

// ErrDockerUnavailable is returned when the Docker daemon cannot be reached.
var ErrDockerUnavailable = errors.New("docker is not available")
