// Package osupdate reports whether a new OS image is staged and waiting for a reboot (bootc/ostree systems).
//
// It asks `rpm-ostree status --json`, which talks to the rpm-ostree daemon over D-Bus and so works for an
// unprivileged user, unlike `bootc status`, which insists on root. On the homeserver image the nightly
// `bootc upgrade` stages the new image; a staged deployment is therefore the "update available" signal.
package osupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// timeout bounds one rpm-ostree call, so a stuck daemon cannot hang the page's request.
const timeout = 5 * time.Second

// State is what the popup shows about OS updates. Its JSON form is the return value of Service.State.
type State struct {
	// Staged is true when a new image is staged and applies on the next reboot.
	Staged bool `json:"staged"`
	// Version is the staged image's version, for example "latest.20261111"; "" when there is none.
	Version string `json:"version"`
}

// parse reads the output of `rpm-ostree status --json`.
func parse(out []byte) (State, error) {
	var status struct {
		Deployments []struct {
			Staged  bool   `json:"staged"`
			Version string `json:"version"`
		} `json:"deployments"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return State{}, fmt.Errorf("osupdate: %w", err)
	}
	for _, d := range status.Deployments {
		if d.Staged {
			return State{Staged: true, Version: d.Version}, nil
		}
	}
	return State{}, nil
}

// Read runs rpm-ostree and returns the current state. It fails when rpm-ostree is missing or errors, for
// example on a system that is not image-based.
func Read(ctx context.Context) (State, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "rpm-ostree", "status", "--json").Output()
	if err != nil {
		return State{}, fmt.Errorf("osupdate: rpm-ostree status: %w", err)
	}
	return parse(out)
}

// Service is bound to the page, which asks for the state each time the popup is shown.
type Service struct {
	read func(context.Context) (State, error)
}

// NewService returns a Service that reads the state with read (Read in the app, a fake in tests). It is a
// constructor argument and not a method because every exported method of a service is bound to the page.
func NewService(read func(context.Context) (State, error)) *Service { return &Service{read: read} }

// State returns the staged-update state; the zero State (nothing staged) when it cannot be read.
func (s *Service) State() State {
	st, err := s.read(context.Background())
	if err != nil {
		return State{}
	}
	return st
}
