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
	defer stop()

	if len(os.Args) < 2 {
		usageAndExit()
	}

	switch os.Args[1] {
	case "start":
		os.Exit(runStart(ctx, os.Args[2:]))
	case "resume":
		os.Exit(runResume(ctx, os.Args[2:]))
	case "help", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "cldpd: unknown subcommand %q\n\n", os.Args[1])
		usageAndExit()
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
	code, err := d.Start(ctx, podName, *issue, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
	}
	return code
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
	code, err := d.Resume(ctx, podName, *prompt, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cldpd: %v\n", err)
	}
	return code
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  cldpd start <pod> --issue <url>")
	fmt.Fprintln(os.Stderr, "  cldpd resume <pod> --prompt <text>")
}

func usageAndExit() {
	printUsage()
	os.Exit(1)
}
