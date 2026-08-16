package fbhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/spf13/afero"

	fberrors "github.com/csummers-dev/vitrine/v3/errors"
	"github.com/csummers-dev/vitrine/v3/events"
	"github.com/csummers-dev/vitrine/v3/files"
	"github.com/csummers-dev/vitrine/v3/fileutils"
	"github.com/csummers-dev/vitrine/v3/trash"
)

var resourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    d.user.Perm.Download,
	})
	if err != nil {
		return errToStatus(err), err
	}

	encoding := r.Header.Get("X-Encoding")
	if file.IsDir {
		file.Sorting = d.user.Sorting
		file.ApplySort()
		return renderJSON(w, r, file)
	} else if encoding == "true" {
		if !d.user.Perm.Download {
			return http.StatusAccepted, nil
		}
		if file.Type != "text" {
			return renderJSON(w, r, file)
		}

		f, err := d.user.Fs.Open(r.URL.Path)
		if err != nil {
			return errToStatus(err), err
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		return 0, err
	}

	if checksum := r.URL.Query().Get("checksum"); checksum != "" {
		err := file.Checksum(checksum)
		if errors.Is(err, fberrors.ErrInvalidOption) {
			return http.StatusBadRequest, nil
		} else if err != nil {
			return http.StatusInternalServerError, err
		}

		// do not waste bandwidth if we just want the checksum
		file.Content = ""
	}

	return renderJSON(w, r, file)
})

func resourceDeleteHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.URL.Path == "/" || !d.user.Perm.Delete {
			return http.StatusForbidden, nil
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		err = d.store.Share.DeleteWithPathPrefix(file.Path)
		if err != nil {
			log.Printf("WARNING: Error(s) occurred while deleting associated shares with file: %s", err)
		}

		// 2.4.0 Stage 2: deletes default to MOVE-TO-TRASH (an instant
		// same-volume rename into the volume's .trash, recorded in the trash
		// index). Permanent deletion happens only when explicitly requested
		// (?permanent=true — Shift+Delete / the Trash view), when the target
		// already lives INSIDE a trash dir (trash is never trashed), or when
		// the trash store is unavailable.
		absPath := d.user.FullPath(r.URL.Path)
		permanent := r.URL.Query().Get("permanent") == "true" ||
			trash.IsTrashPath(absPath) ||
			d.trashStore == nil

		if !permanent {
			var entry trash.Entry
			err = d.RunHook(func() error {
				var trashErr error
				entry, trashErr = d.trashStore.MoveToTrash(
					trashOsFs, absPath, trashScope(d), d.user.Username)
				return trashErr
			}, "delete", r.URL.Path, "", d.user)
			if err != nil {
				return errToStatus(err), err
			}
			// Published as a MOVE (into the trash dir), not a delete — tags
			// follow the file in and follow it back out on restore.
			events.Publish(events.FileMoved{
				Base: eventBase(r, d),
				From: r.URL.Path,
				To:   trashRel(trashScope(d), entry.TrashPath),
			})
			// The undo toast needs the entry id to restore.
			return renderJSON(w, r, map[string]string{"trashId": entry.ID})
		}

		// Permanent delete — thumbnails go too (a trashed file keeps them so a
		// restore doesn't regenerate; LRU reclaims either way).
		err = delThumbs(r.Context(), fileCache, file)
		if err != nil {
			return errToStatus(err), err
		}

		err = d.RunHook(func() error {
			return d.user.Fs.RemoveAll(r.URL.Path)
		}, "delete", r.URL.Path, "", d.user)

		if err != nil {
			return errToStatus(err), err
		}

		events.Publish(events.FileDeleted{
			Base:  eventBase(r, d),
			Path:  r.URL.Path,
			IsDir: file.IsDir,
		})

		return http.StatusNoContent, nil
	})
}

func resourcePostHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}

		// Directories creation on POST.
		if strings.HasSuffix(r.URL.Path, "/") {
			err := d.user.Fs.MkdirAll(r.URL.Path, d.settings.DirMode)
			if err == nil {
				events.Publish(events.FileCreated{
					Base:  eventBase(r, d),
					Path:  r.URL.Path,
					IsDir: true,
				})
			} else if errors.Is(err, syscall.ENOTDIR) {
				// os.MkdirAll reports ENOTDIR ("not a directory") when a
				// component of the path — or the target name itself — already
				// exists as a non-directory (a stray file, a dangling symlink),
				// or when a parent isn't a real directory (e.g. a misconfigured
				// bind-mount). That's a client-visible name conflict, not a
				// server fault, so answer 409 with an actionable message rather
				// than a bare 500 that hides the cause.
				return http.StatusConflict, fmt.Errorf(
					"cannot create directory %q: a file with that name already exists, or its parent is not a directory",
					path.Clean(r.URL.Path),
				)
			}
			return errToStatus(err), err
		}

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err == nil {
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			err = delThumbs(r.Context(), fileCache, file)
			if err != nil {
				return errToStatus(err), err
			}
		}

		// Capture the written size for FileUploaded so the audit log
		// + webhook consumers don't have to re-stat the file.
		var uploadedSize int64
		err = d.RunHook(func() error {
			info, writeErr := writeFile(d.user.Fs, r.URL.Path, r.Body, d.settings.FileMode, d.settings.DirMode)
			if writeErr != nil {
				return writeErr
			}
			uploadedSize = info.Size()

			etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
			w.Header().Set("ETag", etag)
			return nil
		}, "upload", r.URL.Path, "", d.user)

		if err != nil {
			_ = d.user.Fs.RemoveAll(r.URL.Path)
		} else {
			events.Publish(events.FileUploaded{
				Base: eventBase(r, d),
				Path: r.URL.Path,
				Size: uploadedSize,
			})
		}

		return errToStatus(err), err
	})
}

var resourcePutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Modify || !d.Check(r.URL.Path) {
		return http.StatusForbidden, nil
	}

	// Only allow PUT for files.
	if strings.HasSuffix(r.URL.Path, "/") {
		return http.StatusMethodNotAllowed, nil
	}

	exists, err := afero.Exists(d.user.Fs, r.URL.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !exists {
		return http.StatusNotFound, nil
	}

	var savedSize int64
	err = d.RunHook(func() error {
		info, writeErr := writeFile(d.user.Fs, r.URL.Path, r.Body, d.settings.FileMode, d.settings.DirMode)
		if writeErr != nil {
			return writeErr
		}

		savedSize = info.Size()
		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "save", r.URL.Path, "", d.user)

	// A content edit changes the file's size without touching its parent dir's
	// mtime, so emit FileModified so size-tracking subscribers (the folder-size
	// cache, 2.4.0 Stage 4) can invalidate this file's ancestor folders.
	if err == nil {
		events.Publish(events.FileModified{
			Base: eventBase(r, d),
			Path: path.Clean("/" + r.URL.Path),
			Size: savedSize,
		})
	}

	return errToStatus(err), err
})

func resourcePatchHandler(fileCache FileCache) handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		src := r.URL.Path
		dst := r.URL.Query().Get("destination")
		action := r.URL.Query().Get("action")
		dst, err := url.QueryUnescape(dst)
		dst = path.Clean("/" + dst)
		src = path.Clean("/" + src)
		if !d.Check(src) || !d.Check(dst) {
			return http.StatusForbidden, nil
		}
		if err != nil {
			return errToStatus(err), err
		}
		if dst == "/" || src == "/" {
			return http.StatusForbidden, nil
		}

		err = checkParent(src, dst)
		if err != nil {
			return http.StatusBadRequest, err
		}

		// A top-level served folder is typically its own bind mount. Renaming or
		// moving a mount point can't work from inside the container: the rename
		// syscall fails and MoveFile's copy-fallback then can't delete the source
		// (unlinkat → EBUSY), stranding a full duplicate at the destination.
		// Refuse it up front — before any copy runs — with a message the UI
		// surfaces (400 is passed through to the client body).
		if action == "rename" {
			if isMount, ok := fileutils.IsMountpoint(d.user.Fs, src); ok && isMount {
				return http.StatusBadRequest, errMountpointMove(src)
			}
		}

		srcInfo, _ := d.user.Fs.Stat(src)
		dstInfo, _ := d.user.Fs.Stat(dst)
		same := os.SameFile(srcInfo, dstInfo)

		if action != "rename" || !same {
			override := r.URL.Query().Get("override") == "true"
			rename := r.URL.Query().Get("rename") == "true"
			if !override && !rename {
				if _, err = d.user.Fs.Stat(dst); err == nil {
					return http.StatusConflict, nil
				}
			}
			if rename {
				dst = addVersionSuffix(dst, d.user.Fs)
			}

			if override && !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}
		}

		// A copy onto itself (override path — a keep-both copy was renamed
		// apart by addVersionSuffix above) would TRUNCATE the source while
		// reading it: creating the destination clobbers the very file the
		// copy is reading from. Renames are exempt — a case-only rename
		// (a.txt → A.txt on a case-insensitive fs) legitimately has
		// src == dst by os.SameFile and is handled above.
		if action == "copy" && src == dst {
			return http.StatusBadRequest, fmt.Errorf(
				"can't copy %q onto itself", path.Base(src))
		}

		err = d.RunHook(func() error {
			return patchAction(r.Context(), action, src, dst, d, fileCache)
		}, action, src, dst, d.user)

		if err == nil {
			publishPatch(r, d, action, src, dst)
		}

		return errToStatus(err), err
	})
}

// publishPatch fans the PATCH action out onto the right event type. The
// HTTP "rename" action covers both rename-in-place and move-across-
// folders; we pick FileRenamed vs FileMoved based on whether the parent
// directory changed. "copy" always maps to FileCopied.
func publishPatch(r *http.Request, d *data, action, src, dst string) {
	base := eventBase(r, d)
	switch action {
	case "rename":
		if looksLikeMove(src, dst) {
			events.Publish(events.FileMoved{Base: base, From: src, To: dst})
		} else {
			events.Publish(events.FileRenamed{Base: base, From: src, To: dst})
		}
	case "copy":
		events.Publish(events.FileCopied{Base: base, From: src, To: dst})
	}
}

// errMountpointMove is the user-facing error returned when a move/rename is
// refused because the source is itself a filesystem mount point (see
// fileutils.IsMountpoint). It's mapped to 400 so data.go surfaces the message
// to the UI instead of a bare status line.
func errMountpointMove(src string) error {
	return fmt.Errorf(
		"can't move or rename %q — it's a mounted folder (its own mount point), so it "+
			"can't be moved from inside the app. Change its mount path in your Docker/compose "+
			"config instead", path.Base(src))
}

func checkParent(src, dst string) error {
	rel, err := filepath.Rel(src, dst)
	if err != nil {
		return err
	}

	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "../") && rel != ".." && rel != "." {
		return fberrors.ErrSourceIsParent
	}

	return nil
}

func addVersionSuffix(source string, afs afero.Fs) string {
	counter := 1
	dir, name := path.Split(source)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for {
		if _, err := afs.Stat(source); err != nil {
			break
		}
		renamed := fmt.Sprintf("%s(%d)%s", base, counter, ext)
		source = path.Join(dir, renamed)
		counter++
	}

	return source
}

func writeFile(afs afero.Fs, dst string, in io.Reader, fileMode, dirMode fs.FileMode) (os.FileInfo, error) {
	dir, _ := path.Split(dst)
	err := afs.MkdirAll(dir, dirMode)
	if err != nil {
		return nil, err
	}

	file, err := afs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, in)
	if err != nil {
		return nil, err
	}

	// Sync the file to ensure all data is written to storage.
	// to prevent file corruption.
	if err := file.Sync(); err != nil {
		return nil, err
	}

	// Gets the info about the file.
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return info, nil
}

func delThumbs(ctx context.Context, fileCache FileCache, file *files.FileInfo) error {
	for _, previewSizeName := range PreviewSizeNames() {
		size, _ := ParsePreviewSize(previewSizeName)
		if err := fileCache.Delete(ctx, previewCacheKey(file, size)); err != nil {
			return err
		}
	}

	return nil
}

func patchAction(ctx context.Context, action, src, dst string, d *data, fileCache FileCache) error {
	switch action {
	case "copy":
		if !d.user.Perm.Create {
			return fberrors.ErrPermissionDenied
		}

		return fileutils.Copy(d.user.Fs, src, dst, d.settings.FileMode, d.settings.DirMode)
	case "rename":
		if !d.user.Perm.Rename {
			return fberrors.ErrPermissionDenied
		}
		src = path.Clean("/" + src)
		dst = path.Clean("/" + dst)

		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       src,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
		})
		if err != nil {
			return err
		}

		// delete thumbnails
		err = delThumbs(ctx, fileCache, file)
		if err != nil {
			return err
		}

		return fileutils.MoveFile(d.user.Fs, src, dst, d.settings.FileMode, d.settings.DirMode)
	default:
		return fmt.Errorf("unsupported action %s: %w", action, fberrors.ErrInvalidRequestParams)
	}
}

// RecursiveEntry is a single file/directory entry returned by the recursive listing endpoint.
type RecursiveEntry struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modified"`
	IsDir   bool      `json:"isDir"`
}

// resourceGetRecursiveHandler returns a flat list of every file and directory
// under the requested path, walking the tree recursively on the server side
// so the client only needs a single HTTP call.
var resourceGetRecursiveHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	rootPath := r.URL.Path
	if rootPath == "" {
		rootPath = "/"
	}

	// Make sure the root itself exists and is a directory.
	info, err := d.user.Fs.Stat(rootPath)
	if err != nil {
		return errToStatus(err), err
	}
	if !info.IsDir() {
		return http.StatusBadRequest, fmt.Errorf("path is not a directory")
	}

	entries := make([]RecursiveEntry, 0)

	err = afero.Walk(d.user.Fs, rootPath, func(fPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip entries we cannot read
		}

		// Skip the root directory itself.
		if fPath == rootPath {
			return nil
		}

		// Respect user rules.
		if !d.Check(fPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		entries = append(entries, RecursiveEntry{
			Path:    fPath,
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, entries)
})

type DiskUsageResponse struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

var diskUsage = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: false,
		Checker:    d,
		Content:    false,
	})
	if err != nil {
		return errToStatus(err), err
	}
	fPath := file.RealPath()
	if !file.IsDir {
		// V3-G #19: disk usage is volume-level, so a file path (e.g. while a
		// file preview is open) used to return 0/0 — collapsing the sidebar
		// storage bar to "0 B, 0%". Report the usage of the file's containing
		// directory instead so the figure stays correct on preview routes.
		fPath = filepath.Dir(fPath)
	}

	usage, err := disk.UsageWithContext(r.Context(), fPath)
	if err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, &DiskUsageResponse{
		Total: usage.Total,
		Used:  usage.Used,
	})
})
