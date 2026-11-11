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
ifeq ($(BACKEND),cgo)
GOENV  := CGO_ENABLED=1
GOTAGS := production
RUNTAGS :=
else
GOENV  := CGO_ENABLED=0
GOTAGS := purego,production
RUNTAGS := purego
endif
# -race needs cgo (and a C compiler); CI sets RACE=-race, the default keeps `make test` toolchain-free.
RACE ?=

.DEFAULT_GOAL := help
.PHONY: help fmt format lint lint-go lint-frontend tools format-frontend vet test check frontend wails3-install bindings build run clean

help: ## List the targets
	@awk -F':.*## ' '/^[a-zA-Z_-]+:.*## /{printf "  %-17s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development -------------------------------------------------------------

test: ## Run the Go tests (set RACE=-race for the race detector; needs gcc) and the frontend's
	CGO_ENABLED=$(if $(RACE),1,0) go test $(RACE) -tags purego -count=1 ./...
	cd frontend && npm test

fmt: ## Format the Go code
	gofmt -w .

format: fmt format-frontend ## Format the Go code and the frontend

# Dev tool versions. Renovate keeps them current through the `# renovate:` comments (renovate.json5).
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v2.15.0
# renovate: datasource=go depName=golang.org/x/tools
GOIMPORTS_VERSION := v0.52.0

lint: lint-go lint-frontend ## golangci-lint plus the frontend's Biome lint and format check (as CI runs them)

lint-go: ## golangci-lint on the CGO-free build
	CGO_ENABLED=0 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --config .github/linters/golangci.yml --build-tags purego ./...

tools: ## Install goimports and golangci-lint into GOBIN (for editors)
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint-frontend: ## Biome lint, format check and knip (needs `make frontend` once for node_modules)
	cd frontend && npm run lint && npm run format:check && npm run knip

format-frontend: ## Reformat the frontend with Biome
	cd frontend && npm run format

vet: ## go vet, with and without the production tag
	CGO_ENABLED=0 go vet -tags purego ./...
	CGO_ENABLED=0 go vet -tags purego,production ./...

check: ## What CI checks without a C toolchain: gofmt, tidy, vet, tests
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "These files need gofmt:"; echo "$$unformatted"; exit 1; fi
	go mod tidy -diff
	$(MAKE) vet test

frontend: ## Build the frontend into frontend/dist (embedded by the Go binary)
	cd frontend && npm ci --ignore-scripts --no-audit --no-fund && npm run build

wails3-install: ## Install the wails3 CLI at the go.mod version into GOBIN
	CGO_ENABLED=0 go install -tags purego github.com/wailsapp/wails/v3/cmd/wails3@$$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)

bindings: ## Regenerate frontend/bindings from the Go services (needs the wails3 CLI)
	CGO_ENABLED=0 wails3 generate bindings -f "-tags purego" -silent

## Build and run ------------------------------------------------------------

build: frontend ## Release build of the single binary (see BACKEND above)
	$(GOENV) go build -tags $(GOTAGS) -trimpath -buildvcs=false -ldflags "-w -s -X $(MODULE)/internal/version.Version=$(VERSION)" -o $(OUT) .

run: frontend ## Run the app from source (needs a Wayland or X11 session with GTK4 and WebKitGTK 6)
	$(GOENV) go run $(if $(RUNTAGS),-tags $(RUNTAGS)) .

clean: ## Remove the built binary
	rm -f $(OUT)
