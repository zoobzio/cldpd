package cldpd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Pod is a discovered pod definition. It holds the pod name, the absolute path
// to its directory, the parsed configuration, and the absolute path to its Dockerfile.
type Pod struct {
	Name       string    // directory name, used as the pod identifier
	Dir        string    // absolute path to the pod directory
	Config     PodConfig // parsed from pod.json; zero-value if pod.json is absent
	Dockerfile string    // absolute path to the Dockerfile within Dir
}

// PodConfig holds the optional configuration parsed from a pod's pod.json file.
// All fields are optional; absent values use zero values (empty string, nil map).
type PodConfig struct {
	Image     string            `json:"image"`     // Docker image tag; defaults to cldpd-<name> if empty
	Env       map[string]string `json:"env"`       // environment variables passed to the container
	BuildArgs map[string]string `json:"buildArgs"` // --build-arg values passed to docker build
	Workdir   string            `json:"workdir"`   // working directory inside the container
}

// DiscoverPod loads a single pod by name from the given pods directory.
// It returns ErrPodNotFound if the pod directory does not exist, and
// ErrInvalidPod if the directory exists but contains no Dockerfile.
// If pod.json is absent the pod is returned with a zero-value PodConfig.
// If pod.json is present but malformed, an error is returned.
func DiscoverPod(podsDir, name string) (Pod, error) {
	dir := filepath.Join(podsDir, name)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return Pod{}, fmt.Errorf("%w: %s", ErrPodNotFound, name)
	} else if err != nil {
		return Pod{}, fmt.Errorf("stat pod directory: %w", err)
	}

	dockerfile := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
		return Pod{}, fmt.Errorf("%w: %s", ErrInvalidPod, name)
	} else if err != nil {
		return Pod{}, fmt.Errorf("stat Dockerfile: %w", err)
	}

	var config PodConfig
	configPath := filepath.Join(dir, "pod.json")
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return Pod{}, fmt.Errorf("read pod.json: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &config); err != nil {
			return Pod{}, fmt.Errorf("parse pod.json: %w", err)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return Pod{}, fmt.Errorf("resolve pod directory: %w", err)
	}

	return Pod{
		Name:       name,
		Dir:        absDir,
		Config:     config,
		Dockerfile: filepath.Join(absDir, "Dockerfile"),
	}, nil
}

// DiscoverAll loads all valid pods from the given pods directory.
// Entries that are not directories, or directories without a Dockerfile, are skipped.
// The returned slice is sorted by pod name.
func DiscoverAll(podsDir string) ([]Pod, error) {
	entries, err := os.ReadDir(podsDir)
	if err != nil {
		return nil, fmt.Errorf("read pods directory: %w", err)
	}

	var pods []Pod
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pod, err := DiscoverPod(podsDir, entry.Name())
		if err != nil {
			// Skip pods that exist but lack a Dockerfile.
			if isInvalidPod(err) {
				continue
			}
			return nil, err
		}
		pods = append(pods, pod)
	}

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})

	return pods, nil
}

// isInvalidPod reports whether err wraps ErrInvalidPod.
func isInvalidPod(err error) bool {
	return errors.Is(err, ErrInvalidPod)
}
