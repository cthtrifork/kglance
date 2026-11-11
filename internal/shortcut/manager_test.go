package shortcut

import (
	"errors"
	"slices"
	"testing"

	"github.com/godbus/dbus/v5"
)

type call struct {
	method string
	keys   []int32
	flags  uint32
}

type fakeAccel struct {
	calls     []call
	saved     []int32 // what KGlobalAccel answers to the active setShortcut
	registerE error
}

func (f *fakeAccel) doRegister(id []string) error {
	f.calls = append(f.calls, call{method: "doRegister"})
	if !slices.Equal(id, []string{Component, Action, ComponentName, ActionName}) {
		return errors.New("wrong action id")
	}
	return f.registerE
}

func (f *fakeAccel) setShortcut(_ []string, keys []int32, flags uint32) ([]int32, error) {
	f.calls = append(f.calls, call{"setShortcut", keys, flags})
	if f.saved != nil && flags&flagIsDefault == 0 {
		return f.saved, nil
	}
	return keys, nil
}

func TestRegisterUsesDefaultWhenNothingSaved(t *testing.T) {
	f := &fakeAccel{}
	active, err := register(f, 0x18000047)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(active, []int32{0x18000047}) {
		t.Fatalf("active = %#x", active)
	}
	want := []call{
		{"doRegister", nil, 0},
		{"setShortcut", []int32{0x18000047}, flagSetPresent},
		{"setShortcut", []int32{0x18000047}, flagSetPresent | flagIsDefault},
	}
	if len(f.calls) != len(want) {
		t.Fatalf("calls = %+v", f.calls)
	}
	for i := range want {
		if f.calls[i].method != want[i].method || f.calls[i].flags != want[i].flags || !slices.Equal(f.calls[i].keys, want[i].keys) {
			t.Errorf("call %d = %+v, want %+v", i, f.calls[i], want[i])
		}
	}
}

func TestRegisterKeepsTheUsersBinding(t *testing.T) {
	m := New(func() {})
	f := &fakeAccel{saved: []int32{modCtrl | modAlt | 'K'}}
	if err := m.registerWith(f); err != nil {
		t.Fatal(err)
	}
	if got := m.Text(); got != "Ctrl+Alt+K" {
		t.Fatalf("Text() = %q, want the user's binding", got)
	}
}

func TestClearedShortcutHasNoText(t *testing.T) {
	m := New(func() {})
	f := &fakeAccel{saved: []int32{0}}
	if err := m.registerWith(f); err != nil {
		t.Fatal(err)
	}
	if got := m.Text(); got != "" {
		t.Fatalf("Text() = %q, want empty", got)
	}
}

func TestTextStartsAsTheDefault(t *testing.T) {
	if got := New(func() {}).Text(); got != DefaultShortcut {
		t.Fatalf("Text() = %q", got)
	}
}

func TestRegisterReportsErrors(t *testing.T) {
	f := &fakeAccel{registerE: errors.New("no such service")}
	if _, err := register(f, 1); err == nil {
		t.Fatal("register succeeded, want an error")
	}
}

func TestIsPressed(t *testing.T) {
	press := func(component, action string) *dbus.Signal {
		return &dbus.Signal{Name: componentInterface + "." + pressedMember, Body: []any{component, action, int64(1)}}
	}
	cases := []struct {
		name string
		sig  *dbus.Signal
		want bool
	}{
		{"ours", press(Component, Action), true},
		{"another app", press("org.kde.dolphin.desktop", "toggle"), false},
		{"another action", press(Component, "other"), false},
		{"released", &dbus.Signal{Name: componentInterface + ".globalShortcutReleased", Body: []any{Component, Action, int64(1)}}, false},
		{"short body", &dbus.Signal{Name: componentInterface + "." + pressedMember, Body: []any{Component}}, false},
		{"wrong types", &dbus.Signal{Name: componentInterface + "." + pressedMember, Body: []any{1, 2, 3}}, false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		if got := IsPressed(c.sig); got != c.want {
			t.Errorf("%s: IsPressed = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestIsAccelRestarted(t *testing.T) {
	owner := func(name, old, new string) *dbus.Signal {
		return &dbus.Signal{Name: "org.freedesktop.DBus.NameOwnerChanged", Body: []any{name, old, new}}
	}
	if !isAccelRestarted(owner(accelName, "", ":1.42")) {
		t.Error("a new owner of KGlobalAccel should count as a restart")
	}
	if isAccelRestarted(owner(accelName, ":1.42", "")) {
		t.Error("KGlobalAccel disappearing is not a restart yet")
	}
	if isAccelRestarted(owner("org.other", "", ":1.1")) {
		t.Error("another service is not a restart")
	}
}
