# Decisions

Short records of why things are the way they are. Add one when you make a choice that is not obvious from the code.

## 1. Fedora 44 and KDE Plasma 6 only

One platform, one build target. It keeps the packaging to a single rpm and lets the app rely on KDE's own services (KGlobalAccel for the shortcut, UPower for the battery) instead of abstractions. `--toggle` is the escape hatch for other desktops.

## 2. Wails v3 with the purego backend

The popup needs a webview, a frameless transparent window and a small Go backend. Wails v3 gives that. Its `purego` Linux backend loads GTK4 and WebKitGTK 6 at run time with `CGO_ENABLED=0`, so:

- building needs no GTK headers or C compiler; CI, a sandbox and a laptop build the same way;
- `gtk4-layer-shell` is a few `dlopen` calls too, with no extra `pkg-config` or link flags;
- the Fedora packages are runtime dependencies only.

Cost: the backend is marked experimental upstream. `make build BACKEND=cgo` builds with Wails' cgo backend (needs `gcc`, `gtk4-devel`, `webkitgtk6.0-devel`) if purego misbehaves on a real desktop. CI checks that every library the app opens exists on Fedora 44 with the expected soname.

## 3. Positioning with gtk4-layer-shell, not window coordinates

Wayland clients cannot place or centre their own windows. The popup is a full-screen, transparent `overlay` layer surface anchored to all four edges: the card is centred by CSS, the compositor picks the active output, and a click on the transparent backdrop closes the popup (focus loss alone does not fire while the surface holds exclusive keyboard focus). Without layer-shell support (X11, GNOME) the app falls back to a normal frameless window that is centred by Wails and hides on focus loss.

The library is `gtk4-layer-shell` (GTK4), not `gtk-layer-shell` (GTK3). It must be loaded before GTK, so `internal/layershell` opens it from `init()`.

## 4. Global shortcut through KGlobalAccel over D-Bus

The app registers a "Toggle KGlance" action under the component `io.github.cthtrifork.kglance.desktop` with default `Meta+Alt+G`, then listens for `globalShortcutPressed`. The user's own binding from System Settings wins over the default and survives restarts. Wails' xdg-desktop-portal shortcut backend was not used because KGlobalAccel gives a stable action id and a visible entry under the app's name.

## 5. `--toggle` through Wails' single-instance support

A second `kglance --toggle` hands its arguments to the running instance and exits. No extra D-Bus service or socket.

## 6. Vanilla frontend

The UI is a clock, a date and a battery row. No framework. Vite bundles `@wailsio/runtime`. The toolchain follows trivy-dashboard-wail3: Go 1.27, Node 26 and Vite 8. (Vite 8's native bundler crashes on CPUs without AVX2, so the frontend cannot be built on every development VM; CI is the reference build.)

## 7. Releases are a binary and an rpm

No Copr, no rpm repository, no tarball, no deb. The raw binary needs GTK4, WebKitGTK 6 and gtk4-layer-shell installed; the rpm pulls them in, so the README recommends it.

## 8. The battery icon is a three-state traffic light

A 24 px icon cannot show many levels apart, and the bar under the icon already shows the exact percentage. So the icon lights 1, 2 or 3 bars and each bar count has exactly one colour: 3 bars green from 70%, 2 bars orange from 30%, 1 bar red below 30% (an earlier idea let two colours share a bar count, which made the colour and the bars disagree). Charging is blue at every level and adds a bolt, so green means a healthy battery and blue means power is connected; a charging battery below 30% is red, because it is still nearly empty. Fully charged is green, and the no-battery plug is blue. `empty` is always red. The bolt has a halo in the card colour so it stays visible on lit bars. The thresholds live in `frontend/src/format.js`.
