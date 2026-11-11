package power

import (
	"testing"

	"github.com/godbus/dbus/v5"
)

func props(kv map[string]any) map[string]dbus.Variant {
	out := make(map[string]dbus.Variant, len(kv))
	for k, v := range kv {
		out[k] = dbus.MakeVariant(v)
	}
	return out
}

func TestOnBattery(t *testing.T) {
	got := fromProperties(props(map[string]any{
		"IsPresent": true, "Percentage": 72.0, "State": uint32(2), "TimeToEmpty": int64(11520), "TimeToFull": int64(0),
	}))
	want := State{Present: true, Percentage: 72, Status: "discharging", SecondsLeft: 11520}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestCharging(t *testing.T) {
	got := fromProperties(props(map[string]any{
		"IsPresent": true, "Percentage": 38.0, "State": uint32(1), "TimeToEmpty": int64(0), "TimeToFull": int64(3900),
	}))
	want := State{Present: true, Percentage: 38, Status: "charging", SecondsLeft: 3900}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTimeOfTheOtherDirectionIsIgnored(t *testing.T) {
	// UPower can leave a stale TimeToEmpty while charging; it must not be shown as "until full".
	got := fromProperties(props(map[string]any{
		"IsPresent": true, "Percentage": 50.0, "State": uint32(1), "TimeToEmpty": int64(999), "TimeToFull": int64(0),
	}))
	if got.SecondsLeft != 0 {
		t.Fatalf("SecondsLeft = %d, want 0", got.SecondsLeft)
	}
}

func TestNoBattery(t *testing.T) {
	got := fromProperties(props(map[string]any{
		"IsPresent": false, "Percentage": 0.0, "State": uint32(0),
	}))
	if got.Present || got.Percentage != 0 || got.Status != "unknown" {
		t.Fatalf("got %+v, want a state without a battery", got)
	}
	if empty := fromProperties(nil); empty != got {
		t.Fatalf("no properties = %+v, want the same as a missing battery", empty)
	}
}

func TestStatusNames(t *testing.T) {
	for code, want := range map[uint32]string{0: "unknown", 1: "charging", 2: "discharging", 3: "empty", 4: "fully-charged", 5: "pending-charge", 6: "pending-discharge", 99: "unknown"} {
		got := fromProperties(props(map[string]any{"IsPresent": true, "State": code})).Status
		if got != want {
			t.Errorf("State %d = %q, want %q", code, got, want)
		}
	}
}

func TestMergeAppliesPartialUpdates(t *testing.T) {
	var seen []State
	m := NewMonitor(func(s State) { seen = append(seen, s) })
	m.merge(props(map[string]any{"IsPresent": true, "Percentage": 80.0, "State": uint32(2), "TimeToEmpty": int64(3600)}))
	m.merge(props(map[string]any{"Percentage": 79.0}))
	m.merge(props(map[string]any{"Percentage": 79.0})) // nothing changed: no event
	m.merge(props(map[string]any{"State": uint32(1), "TimeToFull": int64(600)}))

	want := []State{
		{Present: true, Percentage: 80, Status: "discharging", SecondsLeft: 3600},
		{Present: true, Percentage: 79, Status: "discharging", SecondsLeft: 3600},
		{Present: true, Percentage: 79, Status: "charging", SecondsLeft: 600},
	}
	if len(seen) != len(want) {
		t.Fatalf("events = %+v, want %+v", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, seen[i], want[i])
		}
	}
	if m.State() != want[2] {
		t.Fatalf("State() = %+v", m.State())
	}
}

func TestParseChange(t *testing.T) {
	sig := &dbus.Signal{
		Name: propsIface + ".PropertiesChanged",
		Body: []any{deviceIface, props(map[string]any{"Percentage": 50.0}), []string{}},
	}
	changed, invalidated, ok := parseChange(sig)
	if !ok || len(changed) != 1 || len(invalidated) != 0 {
		t.Fatalf("parseChange = %v %v %v", changed, invalidated, ok)
	}

	other := &dbus.Signal{Name: propsIface + ".PropertiesChanged", Body: []any{"org.freedesktop.UPower", props(nil), []string{}}}
	if _, _, ok := parseChange(other); ok {
		t.Error("a change of another interface must be ignored")
	}
	if _, _, ok := parseChange(nil); ok {
		t.Error("nil signal must be ignored")
	}
	if _, _, ok := parseChange(&dbus.Signal{Name: propsIface + ".PropertiesChanged", Body: []any{1}}); ok {
		t.Error("short body must be ignored")
	}
}
