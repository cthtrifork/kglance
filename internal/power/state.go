// Package power reads the battery state from UPower over the system bus and reports changes as they
// happen (no polling).
package power

import "github.com/godbus/dbus/v5"

// State is what the popup shows about the power supply. Its JSON form is the Wails event payload and
// the return value of Service.State.
type State struct {
	// Present is false on a machine without a battery; the page then says "On AC".
	Present bool `json:"present"`
	// Percentage is 0 to 100.
	Percentage float64 `json:"percentage"`
	// Status is one of charging, discharging, fully-charged, empty, pending-charge, pending-discharge
	// or unknown.
	Status string `json:"status"`
	// SecondsLeft is the time to empty while discharging and to full while charging; 0 when unknown.
	SecondsLeft int64 `json:"secondsLeft"`
}

// UPower's Device.State values.
var statusNames = map[uint32]string{
	1: "charging",
	2: "discharging",
	3: "empty",
	4: "fully-charged",
	5: "pending-charge",
	6: "pending-discharge",
}

// fromProperties builds a State from the properties of UPower's display device.
func fromProperties(props map[string]dbus.Variant) State {
	var s State
	if v, ok := props["IsPresent"]; ok {
		s.Present, _ = v.Value().(bool)
	}
	if !s.Present {
		return State{Status: "unknown"}
	}
	if v, ok := props["Percentage"]; ok {
		s.Percentage, _ = v.Value().(float64)
	}
	s.Status = "unknown"
	var code uint32
	if v, ok := props["State"]; ok {
		code, _ = v.Value().(uint32)
	}
	if name, ok := statusNames[code]; ok {
		s.Status = name
	}
	var seconds int64
	switch s.Status {
	case "discharging":
		seconds = int64Prop(props, "TimeToEmpty")
	case "charging":
		seconds = int64Prop(props, "TimeToFull")
	}
	if seconds > 0 {
		s.SecondsLeft = seconds
	}
	return s
}

func int64Prop(props map[string]dbus.Variant, name string) int64 {
	v, ok := props[name]
	if !ok {
		return 0
	}
	n, _ := v.Value().(int64)
	return n
}
