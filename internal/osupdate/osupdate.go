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
	// Known is false when rpm-ostree could not be asked or reported no usable deployment.
	Known bool `json:"known"`
	// Staged is true when a new image is staged and applies on the next reboot.
	Staged bool `json:"staged"`
	// Version is the version of the staged image, or of the booted one when none is staged, for example
	// "latest.20261111"; "" when the image carries none.
	Version string `json:"version"`
	// Timestamp is when that image was built, in Unix seconds; 0 when unknown.
	Timestamp int64 `json:"timestamp"`
}

// parse reads the output of `rpm-ostree status --json`: the staged deployment if there is one, else the booted one.
func parse(out []byte) (State, error) {
	var status struct {
		Deployments []struct {
			Booted    bool   `json:"booted"`
			Staged    bool   `json:"staged"`
			Version   string `json:"version"`
			Timestamp int64  `json:"timestamp"`
		} `json:"deployments"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return State{}, fmt.Errorf("osupdate: %w", err)
	}
	var booted *State
	for _, d := range status.Deployments {
		st := State{Known: true, Staged: d.Staged, Version: d.Version, Timestamp: d.Timestamp}
		if d.Staged {
			return st, nil
		}
		if d.Booted {
			booted = &st
		}
	}
	if booted == nil {
		return State{}, nil
	}
	return *booted, nil
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

// State returns the staged-update state; the zero State (not known) when it cannot be read.
func (s *Service) State() State {
	st, err := s.read(context.Background())
	if err != nil {
		return State{}
	}
	return st
}
