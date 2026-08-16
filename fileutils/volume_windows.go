//go:build windows

package fileutils

import "github.com/spf13/afero"

// SameVolume always reports false on Windows: there is no portable st_dev to
// compare, so every move takes the ordinary queued path (correct, just not
// expedited onto the fast lane). The fast-lane optimization is a Unix-only
// nicety — the file servers this targets run Linux.
func SameVolume(_ afero.Fs, _, _ string) (bool, error) {
	return false, nil
}

// DeviceID is unavailable on Windows (no portable st_dev); ok=false makes the
// trash package fall back to a single .trash at the server root.
func DeviceID(_ afero.Fs, _ string) (uint64, bool) {
	return 0, false
}

// IsMountpoint is unavailable on Windows (no portable st_dev to compare); ok is
// always false so callers never block a move on this check. The bind-mount
// deployment this guards against is a Linux/container concern.
func IsMountpoint(_ afero.Fs, _ string) (isMount, ok bool) {
	return false, false
}
