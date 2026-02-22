// Package cldpd is an async pod lifecycle library for Claude Code agent teams.
//
// Each zoobzio repository carries its own agent workflows, standing orders,
// and skills. cldpd spawns pods — Docker containers running Claude Code —
// pointed at GitHub issues, and returns a Session handle for non-blocking
// lifecycle management.
//
// # Basic usage
//
//	d := cldpd.NewDispatcher(podsDir, &cldpd.DockerRunner{})
//
//	session, err := d.Start(ctx, "myrepo", "https://github.com/org/repo/issues/42")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for event := range session.Events() {
//	    if event.Type == cldpd.EventOutput {
//	        fmt.Println(event.Data)
//	    }
//	}
//
//	code, err := session.Wait()
//
// # Session lifecycle
//
// Dispatcher.Start builds the pod's Docker image synchronously, then returns
// a *Session immediately. The container runs in the background. The Session
// emits typed events on its Events() channel:
//
//	BuildStarted → BuildComplete → ContainerStarted → Output* → ContainerExited
//
// Call session.Stop(ctx) for graceful shutdown (SIGTERM with timeout, then
// SIGKILL). Call session.Wait() to block until the container exits.
//
// # Pod definitions
//
// Pods are directories under ~/.cldpd/pods/<name>/ containing:
//   - Dockerfile (required) — the container image definition
//   - pod.json (optional) — configuration: env, mounts, image tag, etc.
//   - template.md (optional) — standing orders prepended to the prompt on Start
//
// See PodConfig for available configuration fields including credential
// passthrough (InheritEnv, Mounts). Mount source paths beginning with ~/
// are expanded to the user's home directory.
//
// # Constraints
//
// cldpd uses the Docker CLI via os/exec and has no external dependencies
// beyond the Go standard library.
package cldpd
