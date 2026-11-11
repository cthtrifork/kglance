# Dev container

A dev container with Go 1.27 and Node 26 (copied from the official `golang` and `node` images, pulled from `mirror.gcr.io`, no dev container features), `make`, the `wails3` CLI and the Go dev tools. Nothing GTK-related is installed: the build is CGO-free (Wails' purego backend), so `make check` and `make build` work as they do on a laptop. The GUI itself cannot run in a container, so shortcut, focus and layer-shell behaviour is checked on a Fedora KDE desktop.

## Start

1. Install Docker or Podman plus an editor that speaks dev containers (VS Code *Dev Containers*, GitHub Codespaces, or `devcontainer up`).
2. Open the repo and choose **Reopen in Container**. The first start runs `.devcontainer/post-create.sh` (`npm ci`, `go mod download`, installs `wails3`, `golangci-lint` and `goimports`).
3. `make dev` starts the Vite dev server with hot reload; open <http://localhost:9245/?demo=charging> (port 9245 is forwarded). `?demo=battery|charging|low|ac` shows a fixed power state without the Go side, which is how the popup's design is worked on.

## Checks

`CGO_ENABLED=0` and `GOFLAGS=-tags=purego` are set in the container, so plain `go build`, `go vet` and `go test` work. `make check`, `make lint`, `make bindings` and `make build` all work. `make test` runs without `-race` (the race detector needs cgo and a C compiler; CI runs it).

## Editor

The container installs the Go and Biome extensions. Format on save (`goimports` for Go, Biome for JS, CSS, HTML and JSON; Markdown is left alone) and `golangci-lint` on save use the repository's config (`.github/linters/golangci.yml`, `frontend/biome.json`). Settings live in `devcontainer.json`.

## What reloads

| You edit | What happens |
|---|---|
| `frontend/src/**`, `frontend/index.html` | Vite hot-reloads the browser |
| `*.go` | nothing runs the app here; run `make test` (or `make check`) |
| an exported service method or type | also run `make bindings` and commit `frontend/bindings` |

## File ownership

The container user is `vscode` with `updateRemoteUserUID` on, so on Linux it takes your host uid and your clone stays writable. `post-create.sh` also marks the workspace as a git `safe.directory` for setups where the owner still differs.

## Limits

- **CPUs without AVX2** (for example an Intel Celeron N4505): Vite 8's native bundler dies with "Illegal instruction", in a container too, because it uses the host CPU. `make dev` and `make frontend` fail there; use `make frontend-wasm` from the host (Node 26 in podman with Rolldown's WebAssembly build). CI and any CPU from the last decade are unaffected.
- **The real popup** (layer-shell, KGlobalAccel, UPower) needs a KDE session; see `AGENTS.md`.

## Troubleshooting

- *"The connection was reset" on localhost:9245*: the dev server must not bind only to the container's loopback. `devcontainer.json` sets `VITE_HOST=0.0.0.0` for that; rebuild the container if yours predates it, or run `VITE_HOST=0.0.0.0 make dev`.
- *Port in use*: `WAILS_VITE_PORT=9300 make dev`.
