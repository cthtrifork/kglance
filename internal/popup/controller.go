// Package popup owns the popup window: when it is shown and hidden, and how the page is told.
package popup

import (
	"sync"
	"time"
)

// Surface is what the controller drives: the real window in the app, a fake in tests.
type Surface interface {
	Show()
	Hide()
}

// focusGrace is how long after Show a focus-lost event is ignored. A compositor can hand focus to the
// new surface and take it back while the window is still appearing, and that must not close it.
const focusGrace = 300 * time.Millisecond

// Controller keeps the visible state and decides when to show and hide the surface.
// Its methods are safe to call from any goroutine (the shortcut listener, the window events and the
// frontend all call them).
type Controller struct {
	mu       sync.Mutex
	surface  Surface
	visible  bool
	shownAt  time.Time
	now      func() time.Time
	onChange func(visible bool)
}

// NewController returns a hidden controller. onChange, if set, runs after every change of state,
// outside the controller's lock.
func NewController(s Surface, onChange func(visible bool)) *Controller {
	return &Controller{surface: s, now: time.Now, onChange: onChange}
}

// Visible reports whether the popup is shown.
func (c *Controller) Visible() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.visible
}

// Show shows the popup. It does nothing if it is already shown.
func (c *Controller) Show() {
	c.mu.Lock()
	if c.visible {
		c.mu.Unlock()
		return
	}
	c.visible = true
	c.shownAt = c.now()
	c.mu.Unlock()

	c.surface.Show()
	c.notify(true)
}

// Hide hides the popup. It does nothing if it is already hidden.
func (c *Controller) Hide() {
	c.mu.Lock()
	if !c.visible {
		c.mu.Unlock()
		return
	}
	c.visible = false
	c.mu.Unlock()

	c.surface.Hide()
	c.notify(false)
}

// Toggle hides the popup when it is shown and shows it otherwise.
func (c *Controller) Toggle() {
	if c.Visible() {
		c.Hide()
		return
	}
	c.Show()
}

// FocusLost hides the popup when it lost focus, unless it was only just shown.
func (c *Controller) FocusLost() {
	c.mu.Lock()
	settled := c.visible && c.now().Sub(c.shownAt) >= focusGrace
	c.mu.Unlock()
	if settled {
		c.Hide()
	}
}

func (c *Controller) notify(visible bool) {
	if c.onChange != nil {
		c.onChange(visible)
	}
}
