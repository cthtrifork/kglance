#!/usr/bin/env bash
# Builds the Fedora rpm from a release binary. Signed when PKG_SIGNING_KEY_FILE is set.
set -euo pipefail

bin="${1:?usage: package-linux.sh <binary> <version> <outdir>}"
version="${2:?usage: package-linux.sh <binary> <version> <outdir>}"
out="${3:?usage: package-linux.sh <binary> <version> <outdir>}"
app=kglance

bin="$(realpath "$bin")"
out="$(realpath -m "$out")"
cd "$(dirname "$0")/.."

[ -x "$bin" ] || { echo "package-linux: $bin is not an executable file" >&2; exit 1; }
command -v wails3 >/dev/null || { echo "package-linux: wails3 not on PATH (make wails3-install)" >&2; exit 1; }
if [ -n "${PKG_SIGNING_KEY_FILE:-}" ]; then
  PKG_SIGNING_KEY_FILE="$(realpath "$PKG_SIGNING_KEY_FILE")"
  export PKG_SIGNING_KEY_FILE
else
  export PKG_SIGNING_KEY_FILE=""
fi

# nfpm wants a bare version: 1.2.3 or 1.2.3-rc.1 (packaged as 1.2.3~rc.1), no leading "v".
# A plain `git describe` dev version (15bc931-dirty) is made valid by prefixing it.
PKG_VERSION="${version#v}"
if ! [[ "$PKG_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$ ]]; then
  PKG_VERSION="0.0.0+${PKG_VERSION//-/.}"
fi
export PKG_VERSION

stage="$(mktemp -d)"
# nfpm does not expand variables in contents paths, so the binary goes where nfpm.yaml expects it
# (bin/ is gitignored).
trap 'rm -rf "${stage:?}" bin/kglance' EXIT
mkdir -p "$out" bin
cp "$bin" bin/kglance

wails3 tool package -name "$app" -format rpm -config ./build/linux/nfpm/nfpm.yaml -out "$stage"

rpm=("$stage"/*.rpm)
[ "${#rpm[@]}" -eq 1 ] || { echo "package-linux: expected one .rpm in $stage" >&2; ls "$stage" >&2; exit 1; }
# nfpm names it after the package only; use the usual <name>-<version>-<release>.<arch>.rpm.
cp "${rpm[0]}" "$out/${app}-${PKG_VERSION}-1.x86_64.rpm"
ls -l "$out"
