SHELL := /bin/sh

MODULE  := github.com/cthtrifork/kglance
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
# OUT: where `make build` writes the binary.
OUT     ?= kglance

# BACKEND: how Wails talks to GTK4 and WebKitGTK 6.
#   purego (default): CGO-free. The libraries are loaded at run time, so building needs no GTK headers and
#                     no C compiler, and the same code builds in a sandbox.
#   cgo:              Wails' original backend. Needs gcc, gtk4-devel and webkitgtk6.0-devel.
BACKEND ?= purego
comma := ,
ifeq ($(BACKEND),cgo)
CGO      := 1
BACKEND_TAG :=
else
CGO      := 0
BACKEND_TAG := purego
endif
GOENV := CGO_ENABLED=$(CGO)
# Tags for a release build, and for `go run`, which must not get `production` (it turns off dev tools).
BUILD_TAGS := $(BACKEND_TAG)$(if $(BACKEND_TAG),$(comma))production
# Checks (vet, tests, lint) always use the CGO-free build, whatever BACKEND is: it is what CI runs.
CHECK_TAGS := purego

# -race needs cgo (and a C compiler); CI sets RACE=-race, the default keeps `make test` toolchain-free.
RACE ?=

NPM_CI := npm ci --ignore-scripts --no-audit --no-fund

# Tool versions. Renovate keeps them current through the `# renovate:` comments (.github/renovate.json5).
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v2.15.0
# renovate: datasource=go depName=golang.org/x/tools
GOIMPORTS_VERSION := v0.52.0

.DEFAULT_GOAL := help
.PHONY: help check test vet lint lint-go lint-frontend fmt format format-frontend tools \
	dev frontend frontend-wasm wails3-install bindings screenshots build rpm run clean

help: ## List the targets
	@awk -F':.*## ' '/^##@/{printf "\n%s\n", substr($$0, 5)} /^[a-zA-Z0-9_-]+:.*## /{printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Check

check: ## What CI checks without a C toolchain: gofmt, tidy, vet, tests
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "These files need gofmt:"; echo "$$unformatted"; exit 1; fi
	go mod tidy -diff
	$(MAKE) vet test

test: ## Go and frontend tests (RACE=-race adds the race detector and needs gcc)
	CGO_ENABLED=$(if $(RACE),1,0) go test $(RACE) -tags $(CHECK_TAGS) -count=1 ./...
	cd frontend && npm test

vet: ## go vet, with and without the production tag
	CGO_ENABLED=0 go vet -tags $(CHECK_TAGS) ./...
	CGO_ENABLED=0 go vet -tags $(CHECK_TAGS),production ./...

lint: lint-go lint-frontend ## golangci-lint, Biome and knip, as CI runs them

lint-go: ## golangci-lint
	CGO_ENABLED=0 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --config .github/linters/golangci.yml --build-tags $(CHECK_TAGS) ./...

lint-frontend: ## Biome lint, format check and knip (needs node_modules: `make dev` or `make frontend` once)
	cd frontend && npm run lint && npm run format:check && npm run knip

fmt: ## Format the Go code
	gofmt -w .

format-frontend: ## Format the frontend with Biome
	cd frontend && npm run format

format: fmt format-frontend ## Format the Go code and the frontend

##@ Frontend

dev: ## Vite dev server with hot reload on http://127.0.0.1:9245; add ?demo=battery|charging|low|ac
	cd frontend && ([ -d node_modules ] || $(NPM_CI)) && npm run dev

frontend: ## Build the frontend into frontend/dist (embedded by the Go binary)
	cd frontend && $(NPM_CI) && npm run build

# Vite 8's native bundler (Rolldown) needs AVX2, which some CPUs (Intel Jasper Lake, such as the Celeron
# N4505) lack: it dies with "Illegal instruction". Rolldown also ships a WebAssembly build; it only works
# on Node 26. Where `node` is already v26 (the dev container) the build runs right here; anywhere else it
# runs in the dev container's image, built with podman from .devcontainer/Dockerfile, so Node 26 is
# pinned in one place. Not needed in CI or on any CPU from the last decade.
DEVCONTAINER_IMAGE := localhost/kglance-dev
# Recursive (=) on purpose: the $$ must reach the shell as a single $.
WASM_BUILD = $(NPM_CI) && \
	npm install --no-save --ignore-scripts --no-audit --no-fund "@rolldown/binding-wasm32-wasi@$$(node -p "require(\"./node_modules/rolldown/package.json\").version")" && \
	npm run build

frontend-wasm: ## Build the frontend on a CPU without AVX2 (Rolldown's WebAssembly build on Node 26; podman if node is not v26)
ifeq ($(shell node --version 2>/dev/null | cut -d. -f1),v26)
	cd frontend && NAPI_RS_FORCE_WASI=error sh -c '$(WASM_BUILD)'
else
	podman build -q -t $(DEVCONTAINER_IMAGE) .devcontainer
	podman run --rm --user root -v "$(CURDIR)/frontend:/src" -w /src -e NAPI_RS_FORCE_WASI=error -e npm_config_cache=/tmp/npm \
		$(DEVCONTAINER_IMAGE) sh -c '$(WASM_BUILD)'
endif

screenshots: ## Render the popup in every demo state into screenshots/ (needs `npx playwright-core install chromium` once)
	cd frontend && $(NPM_CI) && node scripts/screenshots.mjs ../screenshots

##@ Generate

wails3-install: ## Install the wails3 CLI at the go.mod version into GOBIN
	CGO_ENABLED=0 go install -tags $(CHECK_TAGS) github.com/wailsapp/wails/v3/cmd/wails3@$$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)

bindings: ## Regenerate frontend/bindings from the Go services (needs the wails3 CLI)
	CGO_ENABLED=0 wails3 generate bindings -f "-tags $(CHECK_TAGS)" -silent

tools: ## Install goimports and golangci-lint into GOBIN (for editors)
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

##@ Build and run

build: frontend ## Release build of the single binary (see BACKEND above)
	$(GOENV) go build -tags $(BUILD_TAGS) -trimpath -buildvcs=false -ldflags "-w -s -X $(MODULE)/internal/version.Version=$(VERSION)" -o $(OUT) .

rpm: build ## Build the Fedora rpm into dist/ (needs the wails3 CLI, see wails3-install)
	scripts/package-linux.sh $(OUT) $(VERSION) dist

run: frontend ## Run the app from source (needs a Wayland or X11 session with GTK4 and WebKitGTK 6)
	$(GOENV) go run $(if $(BACKEND_TAG),-tags $(BACKEND_TAG)) .

clean: ## Remove the built binary
	rm -f $(OUT)
