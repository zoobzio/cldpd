// Command cldpd dispatches Claude Code agent teams to Docker containers.
//
// Usage:
//
//	cldpd start <pod> --issue <url>
//	cldpd resume <pod> --prompt <text>
//
// Pods are defined as directories under ~/.cldpd/pods/<name>/ containing
// a Dockerfile and an optional pod.json configuration file.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/zoobzio/cldpd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx)
	stop()
	os.Exit(code)
}

// run dispatches the subcommand and returns the process exit code.
func run(ctx context.Context) int {
	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	switch os.Args[1] {
	case "start":
		return runStart(ctx, os.Args[2:])
	case "resume":
		return runResume(ctx, os.Args[2:])
	case "help", "--help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "cldpd: unknown subcommand %q\n\n", os.Args[1])
		printUsage()
		return 1
	}
}

func runStart(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	issue := fs.String("issue", "", "GitHub issue URL (required)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "cldpd start: pod name required")
		return 1
	}
	if *issue == "" {
		fmt.Fprintln(os.Stderr, "cldpd start: --issue is required")
		return 1
	}
	podName := fs.Arg(0)

	runner := &cldpd.DockerRunner{}
	if err := runner.Preflight(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
		return 1
	}

	podsDir, err := cldpd.DefaultPodsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
		return 1
	}

	d := cldpd.NewDispatcher(podsDir, runner)
	session, err := d.Start(ctx, podName, *issue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
		return 1
	}

	return consumeSession(ctx, session)
}

func runResume(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("resume", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	prompt := fs.String("prompt", "", "Follow-up guidance for the running pod (required)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "cldpd resume: pod name required")
		return 1
	}
	if *prompt == "" {
		fmt.Fprintln(os.Stderr, "cldpd resume: --prompt is required")
		return 1
	}
	podName := fs.Arg(0)

	podsDir, err := cldpd.DefaultPodsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
		return 1
	}

	runner := &cldpd.DockerRunner{}
	d := cldpd.NewDispatcher(podsDir, runner)
	session, err := d.Resume(ctx, podName, *prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
		return 1
	}

	return consumeSession(ctx, session)
}

// consumeSession ranges over session events, printing output to stdout and
// errors to stderr. On interrupt (ctx cancellation), it calls session.Stop
// for graceful shutdown. Returns the container's exit code.
func consumeSession(ctx context.Context, session *cldpd.Session) int {
	// Handle interrupt: stop the session gracefully.
	go func() {
		<-ctx.Done()
		_ = session.Stop(context.Background())
	}()

	for event := range session.Events() {
		switch event.Type {
		case cldpd.EventOutput:
			fmt.Println(event.Data)
		case cldpd.EventError:
			fmt.Fprintf(os.Stderr, "cldpd: %s\n", event.Data)
		}
	}

	code, _ := session.Wait()
	return code
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  cldpd start <pod> --issue <url>")
	fmt.Fprintln(os.Stderr, "  cldpd resume <pod> --prompt <text>")
}
