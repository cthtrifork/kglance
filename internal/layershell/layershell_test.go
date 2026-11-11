package layershell

import (
	"errors"
	"testing"
	"unsafe"
)

func TestLoadMissingLibrary(t *testing.T) {
	var a api
	err := load("libdoes-not-exist-kglance.so.0", &a)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("load of a missing library = %v, want ErrUnavailable", err)
	}
	if a.initForWindow != nil {
		t.Fatal("api must stay empty when loading fails")
	}
}

func TestSetupWithoutWindow(t *testing.T) {
	if err := Setup(nil, "kglance"); err == nil {
		t.Fatal("Setup(nil) = nil, want an error")
	}
}

func TestSetupUnavailableIsReported(t *testing.T) {
	// Where the library is installed (a Fedora desktop) there is no GTK display here, and poking
	// layer-shell without one is not what this test is about, so only run it where it is missing.
	if Preload() == nil {
		t.Skip("gtk4-layer-shell is installed")
	}
	var fake byte
	err := Setup(unsafe.Pointer(&fake), "kglance")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Setup without a display = %v, want ErrUnavailable", err)
	}
}
