//go:build !windows

package fileutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestSameVolumeSameDir(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	if err := os.WriteFile(a, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	same, err := SameVolume(afero.NewOsFs(), a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !same {
		t.Fatal("two files in the same tempdir should be on the same volume")
	}
}

func TestSameVolumeUnknownFsIsFalse(t *testing.T) {
	// A MemMapFs FileInfo carries no *syscall.Stat_t, so the device id is
	// unknowable → SameVolume must report false (the safe, ordinary-lane answer).
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/a", []byte("x"), 0o644)
	_ = afero.WriteFile(fs, "/b", []byte("y"), 0o644)

	same, err := SameVolume(fs, "/a", "/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if same {
		t.Fatal("an in-memory fs has no st_dev → must be reported not-same (safe)")
	}
}

func TestSameVolumeMissingPathErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := SameVolume(afero.NewOsFs(), filepath.Join(dir, "nope"), dir); err == nil {
		t.Fatal("a missing source path should surface a stat error")
	}
}

func TestIsMountpointPlainDirIsFalse(t *testing.T) {
	// A subdirectory of a tempdir shares its parent's device id, so it's not a
	// mount point — the common case a real move must be allowed to proceed.
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	isMount, ok := IsMountpoint(afero.NewOsFs(), sub)
	if !ok {
		t.Fatal("device ids of a tempdir subdir and its parent should be knowable")
	}
	if isMount {
		t.Fatal("a plain subdirectory shares its parent's device id → not a mount point")
	}
}

func TestIsMountpointUnknownFsIsNotOk(t *testing.T) {
	// A MemMapFs carries no *syscall.Stat_t, so a mount point can't be
	// determined → ok=false, and callers must treat that as "not a mount point"
	// so ordinary moves are never blocked by uncertainty.
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/parent/child", 0o755)

	isMount, ok := IsMountpoint(fs, "/parent/child")
	if ok {
		t.Fatal("an in-memory fs has no st_dev → ok must be false")
	}
	if isMount {
		t.Fatal("not-ok must report isMount=false so moves aren't blocked")
	}
}
