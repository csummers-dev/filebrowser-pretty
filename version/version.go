//nolint:revive
package version

var (
	// Version is the current File Browser version. This is only a fallback —
	// release/CI builds stamp the real version from the git tag via ldflags
	// (see Taskfile `build:backend`). Keep it in sync with the latest tag.
	Version = "3.1.1"
	// CommitSHA is the commit sha.
	CommitSHA = "(unknown)"
)
