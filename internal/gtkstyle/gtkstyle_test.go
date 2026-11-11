package gtkstyle

import (
	"strings"
	"testing"
)

func TestLoadMissingLibrary(t *testing.T) {
	var a api
	if err := load("libdoes-not-exist-kglance.so.0", &a); err == nil {
		t.Fatal("load of a missing library succeeded, want an error")
	}
	if a.providerNew != nil {
		t.Fatal("api must stay empty when loading fails")
	}
}

func TestCSSClearsTheWindowBackground(t *testing.T) {
	for _, want := range []string{"window", "background: transparent", "box-shadow: none"} {
		if !strings.Contains(css, want) {
			t.Errorf("css %q lacks %q", css, want)
		}
	}
}
