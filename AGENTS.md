# AGENTS.md

Start here. This file is for AI coding agents (and humans) picking up work on **KGlance**: a small popup for KDE Plasma 6 on Fedora that shows the clock, the date and the power status. A global shortcut toggles it; Escape or a click outside closes it. Go + Wails v3, vanilla HTML/CSS/JS frontend, one binary, one platform (Fedora 44, Wayland).

Why things are the way they are: [`docs/decisions.md`](docs/decisions.md). Do not re-litigate without new facts.

## Commands

`make help` lists them.

```sh
make check          # gofmt, go mod tidy -diff, vet (with and without -tags production), tests; no C toolchain needed
make lint           # golangci-lint plus Biome and knip for the frontend
make format         # gofmt plus Biome
make frontend       # frontend -> frontend/dist (embedded by the binary); needed before build/run, not before vet/test
make frontend-wasm  # same as `make frontend`, for a CPU without AVX2 (Vite 8's native bundler crashes there): Node 26 in podman with Rolldown's WebAssembly build
make build          # release binary, CGO-free (BACKEND=cgo for Wails' cgo backend)
make rpm            # the Fedora package (needs the wails3 CLI)
make screenshots    # popup in every demo state, light and dark, into screenshots/ (needs Playwright's Chromium; CI renders them too)
```

## Rules that are easy to get wrong

- **CGO-free build.** The default backend is Wails' `purego` one: every Go command needs `-tags purego` and `CGO_ENABLED=0` (the Makefile does this). A plain `go build` without the tag fails with a buildguard error.
- **The GUI cannot be exercised in a sandbox or in CI.** There is no KDE session, so shortcut, focus, layer-shell and centering behaviour needs a human on Fedora KDE. Test the logic around it with fakes, and say in the PR what was not run on a real desktop.
- **Layer-shell must load first.** `internal/layershell` opens `libgtk4-layer-shell.so.0` before GTK is loaded; do not import anything that initialises GTK ahead of it.
- **Never hand-edit `frontend/bindings`.** Regenerate with `make bindings` after changing an exported service method or type, and commit the result (CI checks it).
- **Always open a PR** for every change, from a new branch off `main`. Do not push to `main`.
- **PR titles are `type(scope)!: summary`** (`feat`, `fix`, `security`, `perf`, `refactor`, `docs`, `test`, `chore`, `ci`, `build`, `deps`; `!` for a breaking change). The title becomes the squash commit and the release-note line; `pr-title.yml` checks it. See [`docs/releasing.md`](docs/releasing.md).
- **Be honest about verification.** Say what you ran and list what you could not test.
- Pin GitHub Actions by commit SHA with a version comment.

## Layout

| Path | What |
|---|---|
| `main.go` | Entry point: flags (`--version`, `--toggle`), Wails app and window |
| `internal/version` | Build version, stamped with `-ldflags -X` |
| `internal/layershell` | `gtk4-layer-shell` through purego: centres a full-screen overlay on the active output |
| `internal/popup` | Window controller: show, hide, toggle, hide on Escape, backdrop click and focus loss |
| `internal/shortcut` | KGlobalAccel registration and the `globalShortcutPressed` listener (godbus) |
| `internal/power` | UPower reader and property-change subscription |
| `frontend/` | Vanilla JS + Vite; `bindings/` is generated (committed); `dist/` is build output |
| `build/linux/` | `.desktop` files, icons, nfpm config |
| `.github/workflows` | `ci`, `security`, `actions-lint`, `pr-title`, `release-drafter`, `screenshots`, `renovate`, `update-branches`, `tag-release`, `release`; see [`docs/ci.md`](docs/ci.md) |
