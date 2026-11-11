package power

import (
	"context"
	"fmt"
	"log"
	"maps"
	"sync"

	"github.com/godbus/dbus/v5"
)

// EventChanged is the Wails event sent to the page with the new State whenever it changes.
const EventChanged = "power:changed"

const (
	upowerName    = "org.freedesktop.UPower"
	displayDevice = dbus.ObjectPath("/org/freedesktop/UPower/devices/DisplayDevice")
	deviceIface   = "org.freedesktop.UPower.Device"
	propsIface    = "org.freedesktop.DBus.Properties"
)

// Monitor keeps the current State and calls onChange whenever it changes.
type Monitor struct {
	onChange func(State)

	mu    sync.Mutex
	props map[string]dbus.Variant
	state State
}

// NewMonitor returns a monitor; until Start succeeds its State is "no battery".
func NewMonitor(onChange func(State)) *Monitor {
	return &Monitor{onChange: onChange, state: fromProperties(nil)}
}

// State returns the latest state.
func (m *Monitor) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Start connects to UPower and follows its display device (the aggregate of all batteries, which also
// reports "not present" on a desktop) until ctx ends. An error means there is no UPower to ask.
func (m *Monitor) Start(ctx context.Context) error {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return fmt.Errorf("system bus: %w", err)
	}
	// Subscribe first, then read, so a change in between is not lost.
	if err := conn.AddMatchSignal(
		dbus.WithMatchSender(upowerName),
		dbus.WithMatchObjectPath(displayDevice),
		dbus.WithMatchInterface(propsIface),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		_ = conn.Close()
		return fmt.Errorf("match signal: %w", err)
	}
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)

	obj := conn.Object(upowerName, displayDevice)
	if err := m.reload(obj); err != nil {
		_ = conn.Close()
		return err
	}

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
				changed, invalidated, ok := parseChange(sig)
				if !ok {
					continue
				}
				if len(invalidated) > 0 {
					if err := m.reload(obj); err != nil {
						log.Printf("power: reload: %v", err)
					}
					continue
				}
				m.merge(changed)
			}
		}
	}()
	return nil
}

// reload reads all properties of the display device.
func (m *Monitor) reload(obj dbus.BusObject) error {
	var all map[string]dbus.Variant
	if err := obj.Call(propsIface+".GetAll", 0, deviceIface).Store(&all); err != nil {
		return fmt.Errorf("read UPower display device: %w", err)
	}
	m.mu.Lock()
	m.props = all
	m.mu.Unlock()
	m.update()
	return nil
}

// merge applies the properties a PropertiesChanged signal carries (it only lists what changed).
func (m *Monitor) merge(changed map[string]dbus.Variant) {
	m.mu.Lock()
	if m.props == nil {
		m.props = make(map[string]dbus.Variant, len(changed))
	}
	maps.Copy(m.props, changed)
	m.mu.Unlock()
	m.update()
}

func (m *Monitor) update() {
	m.mu.Lock()
	next := fromProperties(m.props)
	changed := next != m.state
	m.state = next
	m.mu.Unlock()
	if changed && m.onChange != nil {
		m.onChange(next)
	}
}

// parseChange picks the changed and invalidated properties out of a Properties.PropertiesChanged
// signal for the UPower device interface.
func parseChange(sig *dbus.Signal) (changed map[string]dbus.Variant, invalidated []string, ok bool) {
	if sig == nil || sig.Name != propsIface+".PropertiesChanged" || len(sig.Body) < 3 {
		return nil, nil, false
	}
	iface, _ := sig.Body[0].(string)
	changed, ok1 := sig.Body[1].(map[string]dbus.Variant)
	invalidated, ok2 := sig.Body[2].([]string)
	if iface != deviceIface || !ok1 || !ok2 {
		return nil, nil, false
	}
	return changed, invalidated, true
}

// Service is bound to the page: it returns the state when the page loads, and the page then follows
// the "power:changed" events.
type Service struct{ get func() State }

// NewService returns a Service that reads the state from get. (A constructor argument and not a
// method, because every exported method of a service is bound to the page.)
func NewService(get func() State) *Service { return &Service{get: get} }

// State returns the current power state.
func (s *Service) State() State { return s.get() }
