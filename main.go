// Command kglance is a small popup for KDE Plasma that shows the clock, the date and the power
// status. A global shortcut shows and hides it.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"slices"
	"sync/atomic"

	"github.com/cthtrifork/kglance/internal/layershell"
	"github.com/cthtrifork/kglance/internal/popup"
	"github.com/cthtrifork/kglance/internal/power"
	"github.com/cthtrifork/kglance/internal/shortcut"
	"github.com/cthtrifork/kglance/internal/version"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// appID is the GTK application id, the Wayland app_id and the single-instance key. It must equal the
// name of the .desktop file in build/linux (without the extension).
const appID = "io.github.cthtrifork.kglance"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load gtk4-layer-shell before anything touches GTK (it also does this from init).
	_ = layershell.Preload()

	showVersion := flag.Bool("version", false, "print the version and exit")
	// Not read here: a second launch with --toggle is handled in OnSecondInstanceLaunch, and a first
	// launch shows the popup anyway. It is declared so flag.Parse accepts it.
	// Not read here: a second launch with --toggle is handled in OnSecondInstanceLaunch, and a first
	// launch shows the popup anyway. It is declared so flag.Parse accepts it.
	_ = flag.Bool("toggle", false, "toggle the popup of the running instance (starts the app if none runs)")
	background := flag.Bool("background", false, "start without showing the popup (what the autostart entry uses)")
	flag.Parse()
	if *showVersion {
		fmt.Println("kglance", version.Version)
		return
	}

	frontend, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	var pop atomic.Pointer[popup.Popup]
	toggle := func() {
		if p := pop.Load(); p != nil {
			p.Toggle()
		}
	}
	keys := shortcut.New(toggle)
	svc := popup.NewService(keys.Text)

	// The power monitor sends its changes to the page as events; it is built first because the service
	// that returns its state is part of the app's options.
	var app *application.App
	battery := power.NewMonitor(func(s power.State) { app.Event.Emit(power.EventChanged, s) })

	app = application.New(application.Options{
		Name:        "KGlance",
		Description: "A glance at the time, the date and the battery (" + version.Version + ")",
		Services: []application.Service{
			application.NewService(svc),
			application.NewService(power.NewService(battery.State)),
		},
		Assets: application.AssetOptions{Handler: application.AssetFileServerFS(frontend)},
		Linux: application.LinuxOptions{
			// GTK reports the application id as the Wayland app_id, which KDE matches to the .desktop file.
			ApplicationID: appID,
			ProgramName:   appID,
			// The window is hidden, not closed, between uses; never quit because of that.
			DisableQuitOnLastWindowClosed: true,
		},
		// A second launch (the menu entry, or `kglance --toggle`) hands its arguments to the running
		// instance and exits.
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: appID,
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				if slices.Contains(data.Args, "--toggle") {
					toggle()
				} else if p := pop.Load(); p != nil {
					p.Show()
				}
			},
		},
	})

	p := popup.New(app, svc)
	pop.Store(p)
	// A first launch from the menu or with --toggle shows the popup; autostart does not.
	if !*background {
		p.ShowWhenReady()
	}

	// Register the global shortcut and start the power monitor once the app is running. Without a session bus or KGlobalAccel (not
	// KDE) the app still works through `kglance --toggle`.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		go func() {
			if err := keys.Start(ctx); err != nil {
				log.Printf("global shortcut unavailable (use `kglance --toggle`): %v", err)
			}
		}()
		go func() {
			if err := battery.Start(ctx); err != nil {
				log.Printf("power status unavailable, showing \"On AC\": %v", err)
			}
		}()
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
