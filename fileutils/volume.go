//go:build !windows

package fileutils

import (
	"path"
	"syscall"

	"github.com/spf13/afero"
)

// DeviceID returns the filesystem device id (st_dev) of path, with ok=false
// when it can't be determined (stat error, or a FileInfo without a
// *syscall.Stat_t — e.g. an in-memory test fs). Callers treat not-ok
// conservatively, the same philosophy as SameVolume below. Used by the trash
// package to find the top of the volume a deleted file lives on.
func DeviceID(afs afero.Fs, path string) (uint64, bool) {
	fi, err := afs.Stat(path)
	if err != nil {
		return 0, false
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return uint64(st.Dev), true //nolint:unconvert // Dev is int32 on darwin, uint64 on linux
}

// SameVolume reports whether paths a and b live on the same filesystem volume —
// i.e. whether a rename(a→b) would stay an instant metadata operation rather
// than fall back to a cross-device copy. It compares the underlying device id
// (st_dev) of each path.
//
// It exists purely to ROUTE a move onto the fast worker lane (see jobs.Registry)
// and is deliberately conservative: any uncertainty — a stat error, or a
// filesystem whose FileInfo carries no *syscall.Stat_t (e.g. an in-memory test
// fs) — yields (false, …) so the caller takes the safe, ordinary queued path.
//
// This is a hint, not a guarantee: two separate bind mounts backed by the same
// underlying filesystem share a device id yet still reject a cross-mount rename
// with EXDEV. MoveFileWithProgress copes by falling back to a copy, so a rare
// misclassification only costs that one item a copy — the move is never wrong.
//
// Compare b as the destination's PARENT directory (the destination itself does
// not exist yet), and a as the existing source.
func SameVolume(afs afero.Fs, a, b string) (bool, error) {
	fa, err := afs.Stat(a)
	if err != nil {
		return false, err
	}
	fb, err := afs.Stat(b)
	if err != nil {
		return false, err
	}
	sa, ok1 := fa.Sys().(*syscall.Stat_t)
	sb, ok2 := fb.Sys().(*syscall.Stat_t)
	if !ok1 || !ok2 {
		return false, nil // can't tell → treat as different (safe)
	}
	return sa.Dev == sb.Dev, nil
}

// IsMountpoint reports whether p is itself a filesystem mount point — a
// directory whose device id (st_dev) differs from that of its parent directory.
// In the documented vitrine deployment each served folder is an individual bind
// mount (`- /host/Movies:/srv/Movies`), which makes every top-level folder a
// mount point.
//
// A mount point can't be renamed or moved from inside the container: rename(2)
// fails with EXDEV/EBUSY, and MoveFile's copy-then-delete fallback then can't
// remove the source (unlinkat → EBUSY, "device or resource busy") — so it would
// strand a full duplicate at the destination. The HTTP layer calls this to
// refuse such a move up front, before any bytes are copied.
//
// The second return, ok, is false when it can't be determined (a stat error, or
// a FileInfo without a *syscall.Stat_t — e.g. an in-memory test fs). Callers
// treat not-ok as "not a mount point" so an ordinary move is never blocked by
// uncertainty — the same conservative stance as SameVolume above.
func IsMountpoint(afs afero.Fs, p string) (isMount, ok bool) {
	dev, ok := DeviceID(afs, p)
	if !ok {
		return false, false
	}
	parentDev, ok := DeviceID(afs, path.Dir(p))
	if !ok {
		return false, false
	}
	return dev != parentDev, true
}
