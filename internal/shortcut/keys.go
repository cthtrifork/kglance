package shortcut

import (
	"fmt"
	"strings"
)

// Qt key codes, as KGlobalAccel sends them: a modifier mask ORed with a key.
const (
	modShift int32 = 0x02000000
	modCtrl  int32 = 0x04000000
	modAlt   int32 = 0x08000000
	modMeta  int32 = 0x10000000

	modMask = modShift | modCtrl | modAlt | modMeta

	keyF1 int32 = 0x01000030
)

// namedKeys are the Qt keys with a name rather than a character.
var namedKeys = map[string]int32{
	"Space":  0x20,
	"Escape": 0x01000000,
	"Tab":    0x01000001,
	"Return": 0x01000004,
	"Enter":  0x01000005,
	"Print":  0x01000009,
	"Pause":  0x01000008,
	"Home":   0x01000010,
	"End":    0x01000011,
	"Left":   0x01000012,
	"Up":     0x01000013,
	"Right":  0x01000014,
	"Down":   0x01000015,
	"PgUp":   0x01000016,
	"PgDown": 0x01000017,
}

// Modifier names in the order Qt prints them.
var modifierNames = []struct {
	name string
	bit  int32
}{{"Meta", modMeta}, {"Ctrl", modCtrl}, {"Alt", modAlt}, {"Shift", modShift}}

// ParseKeys turns text such as "Meta+Alt+G" into a Qt key code.
func ParseKeys(text string) (int32, error) {
	var code int32
	parts := strings.Split(text, "+")
	for i, part := range parts {
		last := i == len(parts)-1
		if !last {
			bit, ok := modifierBit(part)
			if !ok {
				return 0, fmt.Errorf("shortcut %q: unknown modifier %q", text, part)
			}
			code |= bit
			continue
		}
		key, ok := parseKey(part)
		if !ok {
			return 0, fmt.Errorf("shortcut %q: unknown key %q", text, part)
		}
		code |= key
	}
	return code, nil
}

func modifierBit(name string) (int32, bool) {
	for _, m := range modifierNames {
		if strings.EqualFold(m.name, name) {
			return m.bit, true
		}
	}
	return 0, false
}

func parseKey(name string) (int32, bool) {
	if len(name) == 1 {
		c := name[0]
		switch {
		case c >= 'a' && c <= 'z':
			return int32(c - 'a' + 'A'), true
		case c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '`':
			return int32(c), true
		}
		return 0, false
	}
	if n := len(name); n >= 2 && n <= 3 && (name[0] == 'F' || name[0] == 'f') {
		var num int
		if _, err := fmt.Sscanf(name[1:], "%d", &num); err == nil && num >= 1 && num <= 35 {
			return keyF1 + int32(num-1), true
		}
	}
	for n, code := range namedKeys {
		if strings.EqualFold(n, name) {
			return code, true
		}
	}
	return 0, false
}

// FormatKeys turns a Qt key code into text such as "Meta+Alt+G". Unknown keys come out as hex.
func FormatKeys(code int32) string {
	var b strings.Builder
	for _, m := range modifierNames {
		if code&m.bit != 0 {
			b.WriteString(m.name + "+")
		}
	}
	b.WriteString(keyName(code &^ modMask))
	return b.String()
}

func keyName(key int32) string {
	switch {
	case key >= 'A' && key <= 'Z', key >= '0' && key <= '9', key == '`':
		return string(rune(key))
	case key >= keyF1 && key < keyF1+35:
		return fmt.Sprintf("F%d", key-keyF1+1)
	}
	for n, code := range namedKeys {
		if code == key {
			return n
		}
	}
	return fmt.Sprintf("0x%x", key)
}
