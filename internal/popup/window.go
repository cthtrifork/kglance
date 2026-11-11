package popup

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/cthtrifork/kglance/internal/gtkstyle"
	"github.com/cthtrifork/kglance/internal/layershell"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Events sent to the page.
const (
	EventShown  = "popup:shown"
	EventHidden = "popup:hidden"
)

// Popup is the Wails window together with its controller.
type Popup struct {
	*Controller
	window *application.WebviewWindow

	prepareOnce sync.Once
	layer       atomic.Bool
}

// New creates the (hidden) popup window and wires its events. svc, which the page calls through the
// Wails bindings, is connected to the new popup.
func New(app *application.App, svc *Service) *Popup {
	p := &Popup{}
	p.window = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:   "popup",
		Title:  "KGlance",
		URL:    "/",
		Hidden: true,
		// Used when layer-shell is not available; with layer-shell the compositor sizes the surface.
		Width:     340,
		Height:    400,
		Frameless: true,
		// Not DisableResize: a layer-shell surface is sized by the compositor (full screen), and a
		// non-resizable GtkWindow ignores that and stays at its default size in the corner.
		AlwaysOnTop: true,
		// Transparent, so the rounded card floats on the desktop (and, with layer-shell, the rest of the
		// full-screen surface is invisible).
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),
		Linux:            application.LinuxWindow{WindowIsTranslucent: true},
	})
	p.Controller = NewController(&surface{p: p}, func(visible bool) {
		if visible {
			app.Event.Emit(EventShown)
		} else {
			app.Event.Emit(EventHidden)
		}
	})
	p.window.OnWindowEvent(events.Common.WindowLostFocus, func(*application.WindowEvent) { p.FocusLost() })
	svc.attach(p)
	return p
}

// ShowWhenReady shows the popup once the page has loaded. It is for the first launch with --toggle,
// when the window may not exist yet.
func (p *Popup) ShowWhenReady() {
	p.window.OnWindowEvent(events.Common.WindowRuntimeReady, func(*application.WindowEvent) { p.Show() })
}

// LayerShell reports whether the popup runs as a layer-shell surface.
func (p *Popup) LayerShell() bool { return p.layer.Load() }

// prepare runs once, on the GTK main thread, just before the first show: the window exists by then,
// and layer-shell has to be set up before a window is mapped.
func (p *Popup) prepare() {
	p.prepareOnce.Do(func() {
		application.InvokeSync(func() {
			// The GtkWindow itself must be see-through, or the card sits on a grey rectangle.
			if err := gtkstyle.TransparentWindows(); err != nil {
				log.Printf("popup: window background stays opaque: %v", err)
			}
			err := layershell.Setup(p.window.NativeWindow(), "kglance")
			if err != nil {
				log.Printf("popup: no layer-shell, using a plain window: %v", err)
				return
			}
			log.Printf("popup: layer-shell active")
			p.layer.Store(true)
		})
	})
}

// surface adapts the Wails window to the Surface the controller drives.
type surface struct{ p *Popup }

func (s *surface) Show() {
	s.p.prepare()
	if !s.p.LayerShell() {
		s.p.window.Center()
	}
	s.p.window.Show()
	s.p.window.Focus()
}

func (s *surface) Hide() { s.p.window.Hide() }
