// Package version holds the build version, stamped at link time:
// -ldflags "-X github.com/cthtrifork/kglance/internal/version.Version=v1.2.3".
package version

// Version is the release version, or "dev" for a plain `go build`.
var Version = "dev"
