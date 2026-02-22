//go:build testing

package cldpd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// makePodDir creates a pod directory with a Dockerfile inside podsDir.
func makePodDir(t *testing.T, podsDir, name string) string {
	t.Helper()
	dir := filepath.Join(podsDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create pod dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	return dir
}

// writePodJSON writes a pod.json file into the given pod directory.
func writePodJSON(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "pod.json"), []byte(content), 0644); err != nil {
		t.Fatalf("write pod.json: %v", err)
	}
}

// writeTemplate writes a template.md file into the given pod directory.
func writeTemplate(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "template.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write template.md: %v", err)
	}
}

func TestDiscoverPod_NotFound(t *testing.T) {
	podsDir := t.TempDir()
	_, err := DiscoverPod(podsDir, "ghost")
	if !errors.Is(err, ErrPodNotFound) {
		t.Errorf("got %v, want ErrPodNotFound", err)
	}
}

func TestDiscoverPod_MissingDockerfile(t *testing.T) {
	podsDir := t.TempDir()
	dir := filepath.Join(podsDir, "nodocker")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	_, err := DiscoverPod(podsDir, "nodocker")
	if !errors.Is(err, ErrInvalidPod) {
		t.Errorf("got %v, want ErrInvalidPod", err)
	}
}

func TestDiscoverPod_NoPodJSON(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "mypod")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Name != "mypod" {
		t.Errorf("Name: got %q, want %q", pod.Name, "mypod")
	}
	if pod.Config.Image != "" {
		t.Errorf("Config.Image: got %q, want empty", pod.Config.Image)
	}
	if pod.Config.Env != nil {
		t.Errorf("Config.Env: got %v, want nil", pod.Config.Env)
	}
	if pod.Config.BuildArgs != nil {
		t.Errorf("Config.BuildArgs: got %v, want nil", pod.Config.BuildArgs)
	}
	if pod.Config.Workdir != "" {
		t.Errorf("Config.Workdir: got %q, want empty", pod.Config.Workdir)
	}
}

func TestDiscoverPod_EmptyPodJSON(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Image != "" {
		t.Errorf("Config.Image: got %q, want empty", pod.Config.Image)
	}
}

func TestDiscoverPod_WithConfig(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{
		"image": "myimage:latest",
		"env": {"FOO": "bar"},
		"buildArgs": {"ARG1": "val1"},
		"workdir": "/app"
	}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Image != "myimage:latest" {
		t.Errorf("Config.Image: got %q, want %q", pod.Config.Image, "myimage:latest")
	}
	if pod.Config.Env["FOO"] != "bar" {
		t.Errorf("Config.Env[FOO]: got %q, want %q", pod.Config.Env["FOO"], "bar")
	}
	if pod.Config.BuildArgs["ARG1"] != "val1" {
		t.Errorf("Config.BuildArgs[ARG1]: got %q, want %q", pod.Config.BuildArgs["ARG1"], "val1")
	}
	if pod.Config.Workdir != "/app" {
		t.Errorf("Config.Workdir: got %q, want %q", pod.Config.Workdir, "/app")
	}
}

func TestDiscoverPod_MalformedPodJSON(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{not valid json`)

	_, err := DiscoverPod(podsDir, "mypod")
	if err == nil {
		t.Fatal("expected error for malformed pod.json, got nil")
	}
}

func TestDiscoverPod_AbsolutePaths(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "mypod")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(pod.Dir) {
		t.Errorf("Dir is not absolute: %q", pod.Dir)
	}
	if !filepath.IsAbs(pod.Dockerfile) {
		t.Errorf("Dockerfile is not absolute: %q", pod.Dockerfile)
	}
}

func TestDiscoverPod_NameFromDirectory(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "myrepo")

	pod, err := DiscoverPod(podsDir, "myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Name != "myrepo" {
		t.Errorf("Name: got %q, want %q", pod.Name, "myrepo")
	}
}

func TestDiscoverAll_Empty(t *testing.T) {
	podsDir := t.TempDir()
	pods, err := DiscoverAll(podsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 0 {
		t.Errorf("got %d pods, want 0", len(pods))
	}
}

func TestDiscoverAll_SkipsNonDirectories(t *testing.T) {
	podsDir := t.TempDir()
	// A plain file — should be skipped
	if err := os.WriteFile(filepath.Join(podsDir, "notapod"), []byte(""), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	makePodDir(t, podsDir, "realpod")

	pods, err := DiscoverAll(podsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 1 {
		t.Errorf("got %d pods, want 1", len(pods))
	}
	if pods[0].Name != "realpod" {
		t.Errorf("pod name: got %q, want %q", pods[0].Name, "realpod")
	}
}

func TestDiscoverAll_SkipsMissingDockerfile(t *testing.T) {
	podsDir := t.TempDir()
	// Directory without Dockerfile — should be skipped, not error
	noDocker := filepath.Join(podsDir, "nodocker")
	if err := os.MkdirAll(noDocker, 0755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	makePodDir(t, podsDir, "goodpod")

	pods, err := DiscoverAll(podsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 1 {
		t.Errorf("got %d pods, want 1", len(pods))
	}
	if pods[0].Name != "goodpod" {
		t.Errorf("pod name: got %q, want %q", pods[0].Name, "goodpod")
	}
}

func TestDiscoverAll_SortedByName(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "zebra")
	makePodDir(t, podsDir, "alpha")
	makePodDir(t, podsDir, "middle")

	pods, err := DiscoverAll(podsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 3 {
		t.Fatalf("got %d pods, want 3", len(pods))
	}
	order := []string{"alpha", "middle", "zebra"}
	for i, want := range order {
		if pods[i].Name != want {
			t.Errorf("pods[%d].Name: got %q, want %q", i, pods[i].Name, want)
		}
	}
}

func TestDiscoverAll_MultiplePods(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "pod-a")
	makePodDir(t, podsDir, "pod-b")

	pods, err := DiscoverAll(podsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 2 {
		t.Errorf("got %d pods, want 2", len(pods))
	}
}

func TestDiscoverAll_InvalidPodsDir(t *testing.T) {
	_, err := DiscoverAll("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for invalid pods directory, got nil")
	}
}

func TestDiscoverPod_InheritEnv(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"inheritEnv": ["HOME", "PATH", "ANTHROPIC_API_KEY"]}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pod.Config.InheritEnv) != 3 {
		t.Fatalf("InheritEnv: got %d entries, want 3", len(pod.Config.InheritEnv))
	}
	want := []string{"HOME", "PATH", "ANTHROPIC_API_KEY"}
	for i, name := range want {
		if pod.Config.InheritEnv[i] != name {
			t.Errorf("InheritEnv[%d]: got %q, want %q", i, pod.Config.InheritEnv[i], name)
		}
	}
}

func TestDiscoverPod_InheritEnv_Absent(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "mypod")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.InheritEnv != nil {
		t.Errorf("InheritEnv: got %v, want nil", pod.Config.InheritEnv)
	}
}

func TestDiscoverPod_Mounts_ReadWrite(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "/host/path", "target": "/container/path"}]}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pod.Config.Mounts) != 1 {
		t.Fatalf("Mounts: got %d entries, want 1", len(pod.Config.Mounts))
	}
	m := pod.Config.Mounts[0]
	if m.Source != "/host/path" {
		t.Errorf("Mount.Source: got %q, want %q", m.Source, "/host/path")
	}
	if m.Target != "/container/path" {
		t.Errorf("Mount.Target: got %q, want %q", m.Target, "/container/path")
	}
	if m.ReadOnly {
		t.Error("Mount.ReadOnly: got true, want false")
	}
}

func TestDiscoverPod_Mounts_ReadOnly(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "/host/keys", "target": "/root/.ssh", "readOnly": true}]}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pod.Config.Mounts) != 1 {
		t.Fatalf("Mounts: got %d entries, want 1", len(pod.Config.Mounts))
	}
	if !pod.Config.Mounts[0].ReadOnly {
		t.Error("Mount.ReadOnly: got false, want true")
	}
}

func TestDiscoverPod_Mounts_Multiple(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{
		"mounts": [
			{"source": "/a", "target": "/b", "readOnly": false},
			{"source": "/c", "target": "/d", "readOnly": true}
		]
	}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pod.Config.Mounts) != 2 {
		t.Fatalf("Mounts: got %d entries, want 2", len(pod.Config.Mounts))
	}
	if pod.Config.Mounts[0].Source != "/a" {
		t.Errorf("Mounts[0].Source: got %q, want %q", pod.Config.Mounts[0].Source, "/a")
	}
	if pod.Config.Mounts[1].ReadOnly != true {
		t.Error("Mounts[1].ReadOnly: got false, want true")
	}
}

func TestDiscoverPod_Mounts_Absent(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "mypod")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Mounts != nil {
		t.Errorf("Mounts: got %v, want nil", pod.Config.Mounts)
	}
}

func TestDiscoverPod_NoPodJSON_InheritEnvAndMountsNil(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "mypod")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.InheritEnv != nil {
		t.Errorf("InheritEnv: got %v, want nil (no pod.json)", pod.Config.InheritEnv)
	}
	if pod.Config.Mounts != nil {
		t.Errorf("Mounts: got %v, want nil (no pod.json)", pod.Config.Mounts)
	}
}

func TestDiscoverPod_Template_Absent(t *testing.T) {
	podsDir := t.TempDir()
	makePodDir(t, podsDir, "mypod")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Template != "" {
		t.Errorf("Template: got %q, want empty string", pod.Template)
	}
}

func TestDiscoverPod_Template_Present(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writeTemplate(t, dir, "# Team Lead Instructions\n\nEnsure origin is up to date.\n")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "# Team Lead Instructions\n\nEnsure origin is up to date.\n"
	if pod.Template != want {
		t.Errorf("Template: got %q, want %q", pod.Template, want)
	}
}

func TestDiscoverPod_Template_Empty(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writeTemplate(t, dir, "")

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Template != "" {
		t.Errorf("Template: got %q, want empty string for empty file", pod.Template)
	}
}

func TestDiscoverPod_Template_Unreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; permission checks do not apply")
	}

	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writeTemplate(t, dir, "some content")
	if err := os.Chmod(filepath.Join(dir, "template.md"), 0000); err != nil {
		t.Fatalf("chmod template.md: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup can remove the file.
		_ = os.Chmod(filepath.Join(dir, "template.md"), 0644)
	})

	_, err := DiscoverPod(podsDir, "mypod")
	if err == nil {
		t.Fatal("expected error for unreadable template.md, got nil")
	}
}

func TestDiscoverPod_Mount_TildeExpanded(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "~/keys", "target": "/root/.ssh/id_ed25519", "readOnly": true}]}`)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home dir: %v", err)
	}

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, "keys")
	if pod.Config.Mounts[0].Source != want {
		t.Errorf("Mount.Source: got %q, want %q", pod.Config.Mounts[0].Source, want)
	}
}

func TestDiscoverPod_Mount_TildeAloneExpanded(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "~", "target": "/root/home"}]}`)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get home dir: %v", err)
	}

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Mounts[0].Source != home {
		t.Errorf("Mount.Source: got %q, want %q", pod.Config.Mounts[0].Source, home)
	}
}

func TestDiscoverPod_Mount_AbsolutePathUnchanged(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "/absolute/path", "target": "/target"}]}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Mounts[0].Source != "/absolute/path" {
		t.Errorf("Mount.Source: got %q, want %q", pod.Config.Mounts[0].Source, "/absolute/path")
	}
}

func TestDiscoverPod_Mount_NoTildeUnchanged(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "relative/path", "target": "/target"}]}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Mounts[0].Source != "relative/path" {
		t.Errorf("Mount.Source: got %q, want %q", pod.Config.Mounts[0].Source, "relative/path")
	}
}

func TestDiscoverPod_Mount_TildeUsernameNotExpanded(t *testing.T) {
	// ~username form is not supported; the source is returned unchanged.
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "mypod")
	writePodJSON(t, dir, `{"mounts": [{"source": "~alice/keys", "target": "/root/.ssh/id_ed25519"}]}`)

	pod, err := DiscoverPod(podsDir, "mypod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Config.Mounts[0].Source != "~alice/keys" {
		t.Errorf("Mount.Source: got %q, want %q (tilde-username must not be expanded)", pod.Config.Mounts[0].Source, "~alice/keys")
	}
}

func TestDiscoverAll_Template_IncludedForPodsWithTemplate(t *testing.T) {
	podsDir := t.TempDir()
	dir := makePodDir(t, podsDir, "podwithtemplate")
	writeTemplate(t, dir, "standing orders")
	makePodDir(t, podsDir, "podwithouttemplate")

	pods, err := DiscoverAll(podsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pods) != 2 {
		t.Fatalf("got %d pods, want 2", len(pods))
	}
	// pods are sorted by name: "podwithouttemplate" < "podwithtemplate"
	if pods[0].Name != "podwithouttemplate" {
		t.Errorf("pods[0].Name: got %q, want %q", pods[0].Name, "podwithouttemplate")
	}
	if pods[0].Template != "" {
		t.Errorf("pods[0].Template: got %q, want empty", pods[0].Template)
	}
	if pods[1].Name != "podwithtemplate" {
		t.Errorf("pods[1].Name: got %q, want %q", pods[1].Name, "podwithtemplate")
	}
	if pods[1].Template != "standing orders" {
		t.Errorf("pods[1].Template: got %q, want %q", pods[1].Template, "standing orders")
	}
}
