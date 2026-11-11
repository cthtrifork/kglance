// Package shortcut registers KGlance's global shortcut with KDE's KGlobalAccel over D-Bus and calls a
// function when it is pressed.
//
// The action shows up in System Settings → Shortcuts under the component's friendly name, with the
// icon of the .desktop file of the same id (build/linux). A binding the user sets there is kept by
// KGlobalAccel and wins over the default on later starts.
package shortcut

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync/atomic"

	"github.com/godbus/dbus/v5"
)

// Identity of the action in KGlobalAccel.
const (
	// Component is the desktop file id; KDE uses it to find the app's name and icon.
	Component     = "io.github.cthtrifork.kglance.desktop"
	ComponentName = "KGlance"
	Action        = "toggle"
	ActionName    = "Toggle KGlance"

	// DefaultShortcut is used until the user binds another key.
	DefaultShortcut = "Meta+Alt+G"
)

const (
	accelName      = "org.kde.kglobalaccel"
	accelPath      = dbus.ObjectPath("/kglobalaccel")
	accelInterface = "org.kde.KGlobalAccel"

	componentInterface = "org.kde.kglobalaccel.Component"
	pressedMember      = "globalShortcutPressed"
)

// KGlobalAccel's setShortcut flags (KGlobalAccel::SetShortcutFlag).
const (
	flagSetPresent uint32 = 0x2
	flagIsDefault  uint32 = 0x8
)

// actionID is KGlobalAccel's four-string action identifier:
// component, action, component friendly name, action friendly name.
func actionID() []string { return []string{Component, Action, ComponentName, ActionName} }

// accel is the part of KGlobalAccel's D-Bus API the manager uses; tests replace it.
type accel interface {
	doRegister(id []string) error
	setShortcut(id []string, keys []int32, flags uint32) ([]int32, error)
}

// register announces the action, sets its default, and returns the keys that are active: the user's
// own binding if KGlobalAccel has one, otherwise the default.
//
// The default goes first, as in KDE's own client library: a freshly registered action loads its saved
// binding and, with none saved, falls back to its default keys. Setting the active keys first leaves a
// new action with no default to fall back to, so nothing is bound.
func register(a accel, def int32) ([]int32, error) {
	id := actionID()
	if err := a.doRegister(id); err != nil {
		return nil, fmt.Errorf("register action: %w", err)
	}
	if _, err := a.setShortcut(id, []int32{def}, flagSetPresent|flagIsDefault); err != nil {
		return nil, fmt.Errorf("set default shortcut: %w", err)
	}
	// Without NoAutoloading, KGlobalAccel replaces the keys passed here with the saved ones, if any.
	active, err := a.setShortcut(id, []int32{def}, flagSetPresent)
	if err != nil {
		return nil, fmt.Errorf("set shortcut: %w", err)
	}
	return active, nil
}

// Manager owns the registration and the listener.
type Manager struct {
	onPressed func()
	def       int32
	text      atomic.Pointer[string]
}

// New returns a manager that calls onPressed whenever the shortcut is pressed.
func New(onPressed func()) *Manager {
	def, err := ParseKeys(DefaultShortcut)
	if err != nil {
		panic(err) // a constant; keys_test.go covers it
	}
	m := &Manager{onPressed: onPressed, def: def}
	m.setText(DefaultShortcut)
	return m
}

// Text is the current shortcut as text such as "Meta+Alt+G"; empty when the user cleared it.
func (m *Manager) Text() string { return *m.text.Load() }

func (m *Manager) setText(s string) { m.text.Store(&s) }

func (m *Manager) setActive(keys []int32) {
	keys = slices.DeleteFunc(slices.Clone(keys), func(k int32) bool { return k == 0 })
	if len(keys) == 0 {
		m.setText("")
		return
	}
	m.setText(FormatKeys(keys[0]))
}

// Start connects to the session bus, registers the action and listens until ctx ends. It returns once
// the registration has been tried; an error means the shortcut is not available (no session bus, no
// KGlobalAccel), and the app still works through `kglance --toggle`.
func (m *Manager) Start(ctx context.Context) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("session bus: %w", err)
	}
	// Subscribe before registering so no press is missed. The sender is not matched: signals come from
	// the component's own object, and they are told apart by their arguments.
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(componentInterface),
		dbus.WithMatchMember(pressedMember),
	); err != nil {
		_ = conn.Close()
		return fmt.Errorf("match signal: %w", err)
	}
	// KGlobalAccel restarts with the Plasma session; register again when it comes back.
	if err := conn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.DBus"),
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
		dbus.WithMatchArg(0, accelName),
	); err != nil {
		_ = conn.Close()
		return fmt.Errorf("match signal: %w", err)
	}
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)

	client := &busAccel{obj: conn.Object(accelName, accelPath)}
	registerErr := m.registerWith(client)

	go func() {
		defer func() { _ = conn.Close() }()
		for {
			select {
			case <-ctx.Done():
				return
			case sig, ok := <-signals:
				if !ok {
					return
				}
				switch {
				case IsPressed(sig):
					m.onPressed()
				case isAccelRestarted(sig):
					if err := m.registerWith(client); err != nil {
						log.Printf("shortcut: re-register after KGlobalAccel restart: %v", err)
					}
				}
			}
		}
	}()
	return registerErr
}

func (m *Manager) registerWith(a accel) error {
	active, err := register(a, m.def)
	if err != nil {
		return err
	}
	m.setActive(active)
	return nil
}

// IsPressed reports whether sig is KGlobalAccel's "shortcut pressed" signal for our action.
func IsPressed(sig *dbus.Signal) bool {
	if sig == nil || sig.Name != componentInterface+"."+pressedMember || len(sig.Body) < 2 {
		return false
	}
	component, ok1 := sig.Body[0].(string)
	action, ok2 := sig.Body[1].(string)
	return ok1 && ok2 && component == Component && action == Action
}

// isAccelRestarted reports whether sig says KGlobalAccel got a new owner on the bus.
func isAccelRestarted(sig *dbus.Signal) bool {
	if sig == nil || sig.Name != "org.freedesktop.DBus.NameOwnerChanged" || len(sig.Body) < 3 {
		return false
	}
	name, _ := sig.Body[0].(string)
	newOwner, _ := sig.Body[2].(string)
	return name == accelName && newOwner != ""
}

// busAccel is accel over D-Bus.
type busAccel struct{ obj dbus.BusObject }

func (b *busAccel) doRegister(id []string) error {
	return b.obj.Call(accelInterface+".doRegister", 0, id).Err
}

func (b *busAccel) setShortcut(id []string, keys []int32, flags uint32) ([]int32, error) {
	// KGlobalAccel takes the keys as an array of ints (Qt key codes).
	ints := make([]int32, len(keys))
	copy(ints, keys)
	var active []int32
	call := b.obj.Call(accelInterface+".setShortcut", 0, id, ints, flags)
	if call.Err != nil {
		return nil, call.Err
	}
	if err := call.Store(&active); err != nil {
		return nil, errors.New("unexpected reply from setShortcut: " + err.Error())
	}
	return active, nil
}
