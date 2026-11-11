// Package memory reports how much memory is in use, read from /proc/meminfo.
package memory

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// State is what the popup shows about memory. Its JSON form is the return value of Service.State.
type State struct {
	// TotalBytes is the installed memory; 0 when it could not be read.
	TotalBytes uint64 `json:"totalBytes"`
	// UsedBytes is the memory programs cannot get without swapping: total minus "available" (the kernel's
	// own estimate, which counts reclaimable cache as free).
	UsedBytes uint64 `json:"usedBytes"`
	// SwapUsedBytes is the swap in use; 0 when there is none.
	SwapUsedBytes uint64 `json:"swapUsedBytes"`
}

// Read returns the current state from /proc/meminfo.
func Read() (State, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return State{}, err
	}
	defer func() { _ = f.Close() }()
	return parse(f)
}

// parse reads the format of /proc/meminfo: "Name:   1234 kB" per line.
func parse(r io.Reader) (State, error) {
	kb := map[string]uint64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		name, rest, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		kb[name] = n
	}
	if err := sc.Err(); err != nil {
		return State{}, err
	}

	total, ok := kb["MemTotal"]
	if !ok || total == 0 {
		return State{}, fmt.Errorf("memory: no MemTotal in /proc/meminfo")
	}
	// Fedora 44's kernel always reports MemAvailable, so there is no fallback for kernels that do not.
	available, ok := kb["MemAvailable"]
	if !ok {
		return State{}, fmt.Errorf("memory: no MemAvailable in /proc/meminfo")
	}
	available = min(available, total)

	s := State{TotalBytes: total * 1024, UsedBytes: (total - available) * 1024}
	if swapTotal := kb["SwapTotal"]; swapTotal > kb["SwapFree"] {
		s.SwapUsedBytes = (swapTotal - kb["SwapFree"]) * 1024
	}
	return s, nil
}

// Service is bound to the page, which asks for the state while the popup is shown.
type Service struct{ read func() (State, error) }

// NewService returns a Service that reads the state with read (Read in the app, a fake in tests). It is a
// constructor argument and not a method because every exported method of a service is bound to the page.
func NewService(read func() (State, error)) *Service { return &Service{read: read} }

// State returns the current memory state; the zero State (no total) when it cannot be read.
func (s *Service) State() State {
	st, err := s.read()
	if err != nil {
		return State{}
	}
	return st
}
