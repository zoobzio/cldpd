package cldpd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Pod is a discovered pod definition. It holds the pod name, the absolute path
// to its directory, the parsed configuration, the absolute path to its Dockerfile,
// and the optional template contents loaded from template.md.
type Pod struct {
	Name       string    // directory name, used as the pod identifier
	Dir        string    // absolute path to the pod directory
	Dockerfile string    // absolute path to the Dockerfile within Dir
	Template   string    // contents of template.md; empty string if absent
	Config     PodConfig // parsed from pod.json; zero-value if pod.json is absent
}

// PodConfig holds the optional configuration parsed from a pod's pod.json file.
// All fields are optional; absent values use zero values (empty string, nil map, nil slice).
type PodConfig struct {
	Env        map[string]string `json:"env"`        // environment variables passed to the container
	BuildArgs  map[string]string `json:"buildArgs"`  // --build-arg values passed to docker build
	Image      string            `json:"image"`      // Docker image tag; defaults to cldpd-<name> if empty
	Workdir    string            `json:"workdir"`    // working directory inside the container
	InheritEnv []string          `json:"inheritEnv"` // host env var names to forward to the container
	Mounts     []Mount           `json:"mounts"`     // bind mounts to pass to the container
}

// DiscoverPod loads a single pod by name from the given pods directory.
// It returns ErrPodNotFound if the pod directory does not exist, and
// ErrInvalidPod if the directory exists but contains no Dockerfile.
// If pod.json is absent the pod is returned with a zero-value PodConfig.
// If pod.json is present but malformed, an error is returned.
// Mount source paths beginning with ~ or ~/ are expanded to the user's home
// directory. ~user expansion is not supported.
// If template.md is absent, Pod.Template is an empty string.
// If template.md is present but cannot be read, an error is returned.
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
	//nolint:gosec // configPath is constructed from a trusted pods directory, not user input
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return Pod{}, fmt.Errorf("read pod.json: %w", err)
	}
	if len(data) > 0 {
		if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
			return Pod{}, fmt.Errorf("parse pod.json: %w", jsonErr)
		}
		// Expand ~ in mount source paths. Neither Go's os/exec nor Docker's -v
		// flag performs shell expansion, so a literal ~ would silently fail to mount.
		if len(config.Mounts) > 0 {
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return Pod{}, fmt.Errorf("resolve home directory: %w", homeErr)
			}
			for i := range config.Mounts {
				if config.Mounts[i].Source == "~" {
					config.Mounts[i].Source = home
				} else if strings.HasPrefix(config.Mounts[i].Source, "~/") {
					config.Mounts[i].Source = filepath.Join(home, config.Mounts[i].Source[2:])
				}
			}
		}
	}

	var template string
	templatePath := filepath.Join(dir, "template.md")
	//nolint:gosec // templatePath is constructed from a trusted pods directory, not user input
	templateData, err := os.ReadFile(templatePath)
	if err != nil && !os.IsNotExist(err) {
		return Pod{}, fmt.Errorf("read template.md: %w", err)
	}
	if len(templateData) > 0 {
		template = string(templateData)
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
		Template:   template,
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

	pods := make([]Pod, 0, len(entries))
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
