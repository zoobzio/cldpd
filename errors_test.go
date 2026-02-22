//go:build testing

package cldpd

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_NonNil(t *testing.T) {
	sentinels := []error{
		ErrPodNotFound,
		ErrInvalidPod,
		ErrBuildFailed,
		ErrContainerFailed,
		ErrSessionNotFound,
		ErrDockerUnavailable,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel error is nil: %v", err)
		}
	}
}

func TestSentinelErrors_Messages(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{ErrPodNotFound, "pod not found"},
		{ErrInvalidPod, "invalid pod: Dockerfile not found"},
		{ErrBuildFailed, "image build failed"},
		{ErrContainerFailed, "container exited with error"},
		{ErrSessionNotFound, "no running session for pod"},
		{ErrDockerUnavailable, "docker is not available"},
	}
	for _, tc := range cases {
		if tc.err.Error() != tc.want {
			t.Errorf("error message: got %q, want %q", tc.err.Error(), tc.want)
		}
	}
}

func TestSentinelErrors_Distinct(t *testing.T) {
	sentinels := []error{
		ErrPodNotFound,
		ErrInvalidPod,
		ErrBuildFailed,
		ErrContainerFailed,
		ErrSessionNotFound,
		ErrDockerUnavailable,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinel[%d] == sentinel[%d]: %v and %v should be distinct", i, j, a, b)
			}
		}
	}
}

func TestSentinelErrors_WrappedIs(t *testing.T) {
	cases := []error{
		ErrPodNotFound,
		ErrInvalidPod,
		ErrBuildFailed,
		ErrContainerFailed,
		ErrSessionNotFound,
		ErrDockerUnavailable,
	}
	for _, sentinel := range cases {
		wrapped := fmt.Errorf("some context: %w", sentinel)
		if !errors.Is(wrapped, sentinel) {
			t.Errorf("errors.Is failed for wrapped %v", sentinel)
		}
	}
}
