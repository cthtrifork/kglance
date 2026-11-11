# kglance

KGlance is a tiny popup/widget for KDE Plasma. Press a shortcut, it appears. Click somewhere else, or hit Escape, and it vanishes like it was never there.

It shows the time, the date, the power status (battery percentage, charging state and time left, or "On AC" on a machine without a battery) and memory use (used of total, swap when it is in use, red when nearly full). Battery changes show up live; memory is read every two seconds while the popup is open.

Inspired by [Lusan-sapkota/KGlance](https://github.com/Lusan-sapkota/KGlance), and much simpler: one Go binary, no world clock, no calendar, no notification history.

![screenshot of widger](https://github.com/cthtrifork/kglance/blob/bot/update-screenshots/screenshots/popup-battery-light.png)

## Install

Each [release](https://github.com/cthtrifork/kglance/releases) has two files:

- `kglance-<version>-1.x86_64.rpm`: recommended. `sudo dnf install ./kglance-*.rpm` pulls in GTK4, WebKitGTK 6 and gtk4-layer-shell and adds the launcher entry and an autostart entry (`kglance --background`, so the shortcut works after login).
- `kglance_linux_amd64`: the bare binary. Install `gtk4`, `webkitgtk6.0` and `gtk4-layer-shell` yourself.

## Use

The default shortcut is `Meta+Alt+G`. Change it in System Settings → Keyboard → Shortcuts → KGlance. `kglance --toggle` toggles the running popup from a terminal or a custom shortcut on another desktop. Starting `kglance` from the menu shows the popup; `kglance --background` starts it hidden.

## Develop

```sh
make check      # gofmt, tidy, vet, tests
make lint       # golangci-lint, Biome, knip
make build      # CGO-free binary; needs no GTK headers
```

See [`AGENTS.md`](AGENTS.md) for the layout and rules, and [`docs/decisions.md`](docs/decisions.md) for why it is built this way.

## License

MIT
