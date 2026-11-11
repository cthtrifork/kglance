// Package layershell turns a GTK4 window into a Wayland layer-shell surface with gtk4-layer-shell,
// loaded at run time through purego (no cgo, no headers).
//
// A layer surface can be anchored to the output edges and is placed by the compositor, which is the
// only way a Wayland client can end up centred on the active screen. The library must be loaded
// before GTK (and so before libwayland-client is bound), which is why Preload runs from init and
// again, explicitly, from main.
package layershell

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Soname of the Fedora gtk4-layer-shell package.
const soname = "libgtk4-layer-shell.so.0"

// Values of the GtkLayerShell enums (gtk4-layer-shell/gtk4-layer-shell.h).
const (
	edgeLeft   = 0
	edgeRight  = 1
	edgeTop    = 2
	edgeBottom = 3

	layerOverlay = 3

	keyboardModeExclusive = 1
)

// ErrUnavailable means the library is missing or the compositor has no layer-shell support.
var ErrUnavailable = errors.New("layer-shell is not available")

type api struct {
	isSupported     func() int32
	initForWindow   func(win unsafe.Pointer)
	setLayer        func(win unsafe.Pointer, layer int32)
	setKeyboardMode func(win unsafe.Pointer, mode int32)
	setAnchor       func(win unsafe.Pointer, edge int32, anchor int32)
	setNamespace    func(win unsafe.Pointer, name string)
	setExclusive    func(win unsafe.Pointer, zone int32)
}

var (
	once    sync.Once
	lib     api
	loadErr error
)

func init() { _ = Preload() }

// Preload opens the library once and returns why that failed, if it did. It is safe to call often.
func Preload() error {
	once.Do(func() { loadErr = load(soname, &lib) })
	return loadErr
}

func load(name string, a *api) error {
	h, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	for sym, fn := range map[string]any{
		"gtk_layer_is_supported":       &a.isSupported,
		"gtk_layer_init_for_window":    &a.initForWindow,
		"gtk_layer_set_layer":          &a.setLayer,
		"gtk_layer_set_keyboard_mode":  &a.setKeyboardMode,
		"gtk_layer_set_anchor":         &a.setAnchor,
		"gtk_layer_set_namespace":      &a.setNamespace,
		"gtk_layer_set_exclusive_zone": &a.setExclusive,
	} {
		if err := register(fn, h, sym); err != nil {
			*a = api{}
			return fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
	}
	return nil
}

// register binds a C function to a Go func variable, turning purego's panic into an error.
func register(fptr any, h uintptr, sym string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s: %v", sym, r)
		}
	}()
	purego.RegisterLibFunc(fptr, h, sym)
	return nil
}

// Supported reports whether layer-shell can be used. It asks GTK's current display, so call it after
// GTK is initialised (for example once the application has started).
func Supported() bool {
	return Preload() == nil && lib.isSupported() != 0
}

// Setup makes win (a GtkWindow*) a full-screen overlay surface that takes the keyboard while it is
// mapped. It must run on the GTK main thread, before the window is first shown.
//
// The surface is anchored to all four edges, so the compositor sizes it to the output it picks (the
// active one) and the page centres its card with CSS; a click on the transparent rest of the surface
// is how the popup knows the user clicked outside.
func Setup(win unsafe.Pointer, namespace string) error {
	if win == nil {
		return errors.New("layershell: no window")
	}
	if !Supported() {
		if err := Preload(); err != nil {
			return err
		}
		return ErrUnavailable
	}
	lib.initForWindow(win)
	lib.setNamespace(win, namespace)
	lib.setLayer(win, layerOverlay)
	for _, edge := range []int32{edgeLeft, edgeRight, edgeTop, edgeBottom} {
		lib.setAnchor(win, edge, 1)
	}
	// -1: ignore other surfaces' exclusive zones (panels), so the overlay covers the whole output and
	// the card sits in the middle of the screen, not of the area left by the panel.
	lib.setExclusive(win, -1)
	// Exclusive keyboard focus while mapped, so Escape works as soon as the popup appears.
	lib.setKeyboardMode(win, keyboardModeExclusive)
	return nil
}
