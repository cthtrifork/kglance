package shortcut

import "testing"

func TestDefaultShortcutEncoding(t *testing.T) {
	got, err := ParseKeys(DefaultShortcut)
	if err != nil {
		t.Fatal(err)
	}
	// Meta (0x10000000) | Alt (0x08000000) | 'G' (0x47): what KGlobalAccel stores for Meta+Alt+G.
	if want := int32(0x18000047); got != want {
		t.Fatalf("ParseKeys(%q) = %#x, want %#x", DefaultShortcut, got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	for _, text := range []string{"Meta+Alt+G", "Ctrl+Shift+F5", "Meta+Space", "Alt+Tab", "Meta+`", "Ctrl+Alt+7", "Meta+F12", "Escape"} {
		code, err := ParseKeys(text)
		if err != nil {
			t.Fatalf("ParseKeys(%q): %v", text, err)
		}
		if got := FormatKeys(code); got != text {
			t.Errorf("FormatKeys(ParseKeys(%q)) = %q", text, got)
		}
	}
}

func TestParseKeysIsCaseInsensitive(t *testing.T) {
	a, _ := ParseKeys("meta+alt+g")
	b, _ := ParseKeys("Meta+Alt+G")
	if a != b || a == 0 {
		t.Fatalf("meta+alt+g = %#x, Meta+Alt+G = %#x", a, b)
	}
}

func TestParseKeysRejectsGarbage(t *testing.T) {
	for _, text := range []string{"", "Hyper+G", "Meta+", "Meta+Alt+Banana", "Meta+F99"} {
		if _, err := ParseKeys(text); err == nil {
			t.Errorf("ParseKeys(%q) succeeded, want an error", text)
		}
	}
}

func TestFormatUnknownKeyFallsBackToHex(t *testing.T) {
	if got := FormatKeys(modMeta | 0x1234567); got != "Meta+0x1234567" {
		t.Fatalf("FormatKeys = %q", got)
	}
}
