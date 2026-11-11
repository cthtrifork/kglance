package memory

import (
	"errors"
	"strings"
	"testing"
)

const sample = `MemTotal:       16384000 kB
MemFree:         1000000 kB
MemAvailable:   10240000 kB
Buffers:          200000 kB
Cached:          5000000 kB
SwapCached:            0 kB
SwapTotal:       2097152 kB
SwapFree:        1048576 kB
HugePages_Total:       0
Not a line
`

func TestParse(t *testing.T) {
	got, err := parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	want := State{
		TotalBytes:    16384000 * 1024,
		UsedBytes:     (16384000 - 10240000) * 1024,
		SwapUsedBytes: 1048576 * 1024,
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseNoSwap(t *testing.T) {
	got, err := parse(strings.NewReader("MemTotal: 1000 kB\nMemAvailable: 400 kB\nSwapTotal: 0 kB\nSwapFree: 0 kB\n"))
	if err != nil || got.SwapUsedBytes != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestParseClampsAvailableToTotal(t *testing.T) {
	got, err := parse(strings.NewReader("MemTotal: 1000 kB\nMemAvailable: 2000 kB\n"))
	if err != nil || got.UsedBytes != 0 {
		t.Fatalf("got %+v, %v; used memory must not underflow", got, err)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	for _, in := range []string{"", "MemFree: 5 kB\n", "MemTotal: 0 kB\n", "MemTotal: lots kB\n", "MemTotal: 1000 kB\nMemFree: 100 kB\n"} {
		if _, err := parse(strings.NewReader(in)); err == nil {
			t.Errorf("parse(%q) succeeded, want an error", in)
		}
	}
}

func TestServiceReturnsZeroStateOnError(t *testing.T) {
	s := NewService(func() (State, error) { return State{TotalBytes: 1}, errors.New("boom") })
	if got := s.State(); got != (State{}) {
		t.Fatalf("State() = %+v, want the zero state", got)
	}
	ok := NewService(func() (State, error) { return State{TotalBytes: 7}, nil })
	if got := ok.State(); got.TotalBytes != 7 {
		t.Fatalf("State() = %+v", got)
	}
}
