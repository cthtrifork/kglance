// Command kglance is a small popup for KDE Plasma that shows the clock, the date and the power
// status. A global shortcut shows and hides it.
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"slices"
	"sync/atomic"

	"github.com/cthtrifork/kglance/internal/layershell"
	"github.com/cthtrifork/kglance/internal/popup"
	"github.com/cthtrifork/kglance/internal/version"
	"github.com/wailsapp/wails/v3/pkg/application"
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
	toggle := flag.Bool("toggle", false, "toggle the popup of the running instance (starts the app if none runs)")
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
	svc := popup.NewService(nil)

	app := application.New(application.Options{
		Name:        "KGlance",
		Description: "A glance at the time, the date and the battery (" + version.Version + ")",
		Services:    []application.Service{application.NewService(svc)},
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(frontend)},
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
				p := pop.Load()
				if p == nil {
					return
				}
				if slices.Contains(data.Args, "--toggle") {
					p.Toggle()
				} else {
					p.Show()
				}
			},
		},
	})

	p := popup.New(app, svc)
	pop.Store(p)
	if *toggle {
		p.ShowWhenReady()
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
