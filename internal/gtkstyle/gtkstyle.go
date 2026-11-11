// Package gtkstyle makes GTK windows see-through, with a CSS rule loaded through purego (no cgo).
//
// Wails makes the web view transparent, but the GtkWindow behind it keeps the theme's opaque background,
// which would show as a grey rectangle (a grey screen, with a full-screen layer-shell surface).
package gtkstyle

import (
	"errors"
	"fmt"

	"github.com/ebitengine/purego"
)

// css clears the window background and the shadow some themes draw around client-side decorated windows.
const css = `window, window.background, window.csd { background: transparent; background-color: transparent; box-shadow: none; }`

// GTK_STYLE_PROVIDER_PRIORITY_APPLICATION: above the theme, below user overrides.
const priorityApplication = 600

type api struct {
	providerNew       func() uintptr
	loadFromString    func(provider uintptr, css string)
	displayGetDefault func() uintptr
	addForDisplay     func(display, provider uintptr, priority uint32)
}

// TransparentWindows makes every GtkWindow of the application transparent. It must run on the GTK main
// thread, after GTK is initialised (the application has started).
func TransparentWindows() error {
	var a api
	if err := load("libgtk-4.so.1", &a); err != nil {
		return err
	}
	display := a.displayGetDefault()
	if display == 0 {
		return errors.New("gtkstyle: no default display")
	}
	provider := a.providerNew()
	a.loadFromString(provider, css)
	a.addForDisplay(display, provider, priorityApplication)
	return nil
}

func load(name string, a *api) error {
	h, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("gtkstyle: %w", err)
	}
	for sym, fn := range map[string]any{
		"gtk_css_provider_new":                       &a.providerNew,
		"gtk_css_provider_load_from_string":          &a.loadFromString, // GTK 4.12 or newer; Fedora 44 has 4.20
		"gdk_display_get_default":                    &a.displayGetDefault,
		"gtk_style_context_add_provider_for_display": &a.addForDisplay,
	} {
		if err := register(fn, h, sym); err != nil {
			return err
		}
	}
	return nil
}

// register binds a C function to a Go func variable, turning purego's panic into an error.
func register(fptr any, h uintptr, sym string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("gtkstyle: %s: %v", sym, r)
		}
	}()
	purego.RegisterLibFunc(fptr, h, sym)
	return nil
}
