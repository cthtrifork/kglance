# kglance

KGlance is a tiny popup/widget for KDE Plasma. Press a shortcut, it appears. Click somewhere else, or hit Escape, and it vanishes like it was never there.

It shows the time, the date, the power status and memory use. Power is the battery percentage, charging state and time left, or "On AC" on a machine without a battery; memory is used of total, with swap when it is in use, and turns red when nearly full. Battery changes show up live, memory is read every two seconds, and nothing ticks while the popup is hidden.

A small pill above the footer shows the OS update state. When a new image is staged for the next reboot it is highlighted and says so with a second line: `OS update staged · latest.20261111` and `Built 2h ago · reboot to apply`. Otherwise it is quiet: `No OS update staged · built 3d ago` (when the running image was built), or `OS update status unavailable` when it cannot be read. This pill is made for one kind of system, an image-based Fedora (bootc/ostree), see below.

Inspired by [Lusan-sapkota/KGlance](https://github.com/Lusan-sapkota/KGlance), and much simpler: one Go binary, no world clock, no calendar, no notification history.

![screenshot of the widget](https://github.com/cthtrifork/kglance/blob/bot/update-screenshots/screenshots/popup-battery-light.png)

## Requirements

Fedora 44 with KDE Plasma 6 on Wayland is the only supported platform. KGlance is built for one specific target, the author's own machines: a Fedora 44 KDE desktop installed as a bootc (image-based, ostree) system, from [homeserver-bootc](https://github.com/cthtrifork/homeserver-bootc). It also works on a regular package-based Fedora 44 KDE, except for the OS update pill, which needs `rpm-ostree` and shows `OS update status unavailable` without it.

- **KDE Plasma** provides the global shortcut (KGlobalAccel). On another desktop the popup still works through `kglance --toggle`, but there is no built-in shortcut.
- **gtk4-layer-shell** lets the compositor centre the popup on the active screen, which Wayland does not allow an ordinary window to do. Without it (or on X11) the popup falls back to a plain window that is centred by the toolkit and hides when it loses focus.
- **rpm-ostree** (present on Fedora's image-based/bootc systems) provides the OS update pill. KGlance runs `rpm-ostree status --json` as your user each time the popup is shown; it needs no root and only reads what the host already knows. It never checks a registry or downloads anything, so an update shows as "staged" only after the system's own `bootc upgrade` has fetched it, and the pill clears after the reboot that applies it.
- **UPower** (normally installed on a Plasma desktop; the rpm recommends it) provides the battery state. Without it the popup shows "On AC".
- **GTK 4 and WebKitGTK 6** draw the window. They are loaded at run time, so the binary has no build-time dependency on them.

## Install

Each [release](https://github.com/cthtrifork/kglance/releases) has two files plus `SHA256SUMS` and an SBOM:

- `kglance-<version>-1.x86_64.rpm` (for example `kglance-0.0.1-1.x86_64.rpm`): recommended. `sudo dnf install ./kglance-*.rpm` pulls in GTK4, WebKitGTK 6 and gtk4-layer-shell, and adds the launcher entry and an autostart entry (`kglance --background`, so the shortcut works after login).
- `kglance_linux_amd64`: the bare binary. Install `gtk4`, `webkitgtk6.0`, `libsoup3` and `gtk4-layer-shell` yourself.

To build it yourself instead, run `make rpm` (see below).

Both files carry signed build-provenance attestations; verify one with `gh attestation verify <file> --repo cthtrifork/kglance`.

## Use

- The default shortcut is `Meta+Alt+G`. Change it in System Settings → Keyboard → Shortcuts → KGlance; the popup's footer shows whatever it is set to.
- `Esc` or a click outside the card closes the popup.
- Starting `kglance` from the menu shows the popup. `kglance --toggle` toggles the running one (handy for a custom shortcut on another desktop), and `kglance --background` starts it hidden, which is what autostart uses.
- `kglance --version` prints the version.

There is no configuration file.

### If something is off

Run `kglance` from a terminal; it logs why a part is unavailable.

- **No shortcut under "KGlance" in System Settings:** the launcher entry (`io.github.cthtrifork.kglance.desktop`) must be installed, which the rpm does. The log line is `global shortcut unavailable`.
- **The popup is not centred:** layer-shell is missing or unsupported (`no layer-shell, using a plain window`). Install `gtk4-layer-shell`.
- **"On AC" on a laptop:** UPower is not reachable (`power status unavailable`).

## Screenshots

Screenshots are never committed. Every pull request that touches the frontend gets a comment with the popup rendered in each demo state (battery, charging, low battery, no battery; light and dark), and the draft pull request [Screenshots preview](https://github.com/cthtrifork/kglance/pulls?q=is%3Apr+Screenshots+preview) shows the ones from `main` in its Files changed tab.

## Develop

```sh
make help       # every target, grouped
make check      # gofmt, tidy, vet, tests (Go and frontend)
make lint       # golangci-lint, Biome, knip
make dev        # Vite dev server with hot reload; add ?demo=charging (battery|charging|low|ac) to the URL, and &update=staged|none|unknown for the OS update pill
make build      # the single binary, CGO-free; needs no GTK headers
make rpm        # the Fedora package (needs the wails3 CLI: make wails3-install)
make screenshots
```

The build needs only Go and Node (versions in `.tool-versions`). A [dev container](docs/devcontainer.md) has both, and `make frontend-wasm` builds the frontend on a CPU without AVX2.

The shortcut, centring and focus behaviour needs a real KDE session to check; everything around it is unit tested with fakes.

More: [`AGENTS.md`](AGENTS.md) (layout and rules), [`docs/decisions.md`](docs/decisions.md) (why it is built this way), [`docs/ci.md`](docs/ci.md) (workflows, the variable and secret, Renovate, screenshots), [`docs/releasing.md`](docs/releasing.md) (cutting a release) and [`docs/devcontainer.md`](docs/devcontainer.md).

## License

MIT
