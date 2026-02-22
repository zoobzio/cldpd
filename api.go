// Package cldpd dispatches Claude Code agent teams to Docker containers.
//
// Each zoobzio repository carries its own agent workflows, standing orders,
// and skills. cldpd spawns pods — Docker containers running Claude Code —
// pointed at GitHub issues, then streams the team leader's output back
// to the caller.
package cldpd
