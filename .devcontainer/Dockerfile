# Go and Node are copied from the official images, so their versions are plain `FROM` tags that Renovate
# keeps in step with .tool-versions (the "go toolchain" and "node" groups in .github/renovate.json5).
FROM mirror.gcr.io/library/golang:1.27-trixie AS go
FROM mirror.gcr.io/library/node:26-trixie-slim AS node

FROM mcr.microsoft.com/devcontainers/base:trixie@sha256:38f9e6c291f054e3c868d95a5477cbfcf76c0fa7c218dda65dc6aa3fbfe06871

COPY --from=go /usr/local/go /usr/local/go
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
 && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

# make runs the Makefile; nothing GTK-related is installed: the build is CGO-free (Wails' purego backend).
RUN apt-get update \
 && apt-get install -y --no-install-recommends make \
 && rm -rf /var/lib/apt/lists/* \
 && mkdir -p /home/vscode/go \
 && chown -R vscode:vscode /home/vscode/go

ENV GOPATH=/home/vscode/go \
    PATH=/usr/local/go/bin:/home/vscode/go/bin:${PATH}
