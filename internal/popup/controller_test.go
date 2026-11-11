package popup

import (
	"slices"
	"testing"
	"time"
)

type fakeSurface struct{ calls []string }

func (f *fakeSurface) Show() { f.calls = append(f.calls, "show") }
func (f *fakeSurface) Hide() { f.calls = append(f.calls, "hide") }

func newTest(t *testing.T) (*Controller, *fakeSurface, *time.Time, *[]bool) {
	t.Helper()
	s := &fakeSurface{}
	var changes []bool
	c := NewController(s, func(v bool) { changes = append(changes, v) })
	now := time.Unix(1000, 0)
	c.now = func() time.Time { return now }
	return c, s, &now, &changes
}

func TestToggleShowsThenHides(t *testing.T) {
	c, s, _, changes := newTest(t)

	c.Toggle()
	if !c.Visible() {
		t.Fatal("first toggle should show the popup")
	}
	c.Toggle()
	if c.Visible() {
		t.Fatal("second toggle should hide the popup")
	}
	if want := []string{"show", "hide"}; !slices.Equal(s.calls, want) {
		t.Fatalf("surface calls = %v, want %v", s.calls, want)
	}
	if want := []bool{true, false}; !slices.Equal(*changes, want) {
		t.Fatalf("changes = %v, want %v", *changes, want)
	}
}

func TestShowAndHideAreIdempotent(t *testing.T) {
	c, s, _, changes := newTest(t)

	c.Hide() // already hidden
	c.Show()
	c.Show() // already shown
	c.Hide()
	c.Hide()

	if want := []string{"show", "hide"}; !slices.Equal(s.calls, want) {
		t.Fatalf("surface calls = %v, want %v", s.calls, want)
	}
	if want := []bool{true, false}; !slices.Equal(*changes, want) {
		t.Fatalf("changes = %v, want %v", *changes, want)
	}
}

func TestFocusLostIsIgnoredRightAfterShow(t *testing.T) {
	c, _, now, _ := newTest(t)

	c.Show()
	*now = now.Add(focusGrace - time.Millisecond)
	c.FocusLost()
	if !c.Visible() {
		t.Fatal("focus lost during the grace period must not hide the popup")
	}

	*now = now.Add(time.Millisecond)
	c.FocusLost()
	if c.Visible() {
		t.Fatal("focus lost after the grace period should hide the popup")
	}
}

func TestFocusLostWhileHiddenDoesNothing(t *testing.T) {
	c, s, now, changes := newTest(t)
	*now = now.Add(time.Hour)

	c.FocusLost()

	if len(s.calls) != 0 || len(*changes) != 0 {
		t.Fatalf("focus lost while hidden changed state: calls=%v changes=%v", s.calls, *changes)
	}
}
