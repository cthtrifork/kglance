# kglance

KGlance is a tiny popup for KDE Plasma. Press a shortcut, it appears. Click somewhere else, or hit Escape, and it vanishes like it was never there.

It shows the time, the date and the power status (battery percentage, charging state and time left, or "On AC" on a machine without a battery). Battery changes show up live.

Inspired by [Lusan-sapkota/KGlance](https://github.com/Lusan-sapkota/KGlance), and much simpler: one Go binary, no world clock, no calendar, no notification history.

**Status:** under construction. Fedora 44 with KDE Plasma 6 on Wayland is the only supported platform.

## Install

Each [release](https://github.com/cthtrifork/kglance/releases) has two files:

- `kglance-<version>.x86_64.rpm`: recommended. `sudo dnf install ./kglance-*.rpm` pulls in GTK4, WebKitGTK 6 and gtk4-layer-shell and adds the launcher and autostart entries.
- `kglance_linux_amd64`: the bare binary. Install `gtk4`, `webkitgtk6.0` and `gtk4-layer-shell` yourself.

## Use

The default shortcut is `Meta+Alt+G`. Change it in System Settings → Keyboard → Shortcuts → KGlance. `kglance --toggle` toggles the running popup from a terminal or a custom shortcut on another desktop.

## Develop

```sh
make check      # gofmt, tidy, vet, tests
make lint       # golangci-lint, Biome, knip
make build      # CGO-free binary; needs no GTK headers
```

See [`AGENTS.md`](AGENTS.md) for the layout and rules, and [`docs/decisions.md`](docs/decisions.md) for why it is built this way.

## License

MIT
