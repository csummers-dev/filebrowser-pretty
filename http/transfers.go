package fbhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/afero"

	fberrors "github.com/csummers-dev/vitrine/v3/errors"
	"github.com/csummers-dev/vitrine/v3/events"
	"github.com/csummers-dev/vitrine/v3/fileutils"
	"github.com/csummers-dev/vitrine/v3/jobs"
)

// transferPayload is the per-job execution context the worker reads back off the
// job. It captures everything a copy/move needs that varies per request — the
// user's scoped filesystem, the mode bits, and the audit event base — so the
// single shared executor can run any user's transfer.
type transferPayload struct {
	fs       afero.Fs
	fileMode fs.FileMode
	dirMode  fs.FileMode
	base     events.Base
	// prewalk counts bytes up front so the % has a denominator. Skipped for
	// same-volume (fast-lane) moves — an instant rename copies nothing, so
	// walking the (possibly huge) tree is pure wasted I/O (2.4.0 Stage 3 / G).
	prewalk bool
	// retry marks a re-run of a failed/interrupted job: each item's
	// destination is cleared first to drop any partial left by the original
	// crash before re-copying (2.4.0 Stage 3 / D).
	retry bool
	// verify turns on the post-copy integrity check (2.4.0 Stage 4 / F): a copy
	// is re-read + hashed against its source, and a cross-volume move verifies
	// before deleting the source. A mismatch fails the item, keeping the source.
	verify bool
}

// transferManager owns the background transfer (move/copy) job registry and the
// executor that runs each job with progress + cancellation. One per server; its
// registry runs jobs through a single sequential worker.
type transferManager struct {
	reg *jobs.Registry
}

func newTransferManager() *transferManager {
	tm := &transferManager{}
	tm.reg = jobs.New(tm.execute)
	return tm
}

// execute is the jobs.Executor. It copies/moves each item with byte-level
// progress, honoring cancellation. On any item error (including cancel) it rolls
// back that item's partial destination, leaves the source untouched, and stops
// the batch. Each successful item publishes the matching audit event — parity
// with the synchronous PATCH handler.
//
// Note: unlike the PATCH handler this does NOT proactively evict the moved
// file's thumbnail cache entries. Those are keyed by path, so a moved file just
// regenerates fresh thumbnails at its new path; the stale entries are reclaimed
// by the cache's LRU. Skipping it keeps the executor free of cache/file deps.
func (tm *transferManager) execute(ctx context.Context, j *jobs.Job) error {
	pl, ok := j.Payload().(*transferPayload)
	if !ok || pl == nil {
		return errors.New("transfer: missing execution context")
	}

	// Pre-walk so the percentage has a real denominator from the first poll —
	// unless this is a fast-lane (same-volume) move, where the rename copies
	// nothing and walking the tree would be wasted I/O (Stage 3 / G). Then the
	// file count is just the top-level item count; MarkComplete pins 100%.
	if pl.prewalk {
		var totalBytes int64
		var totalFiles int
		for _, it := range j.Items() {
			b, n, err := fileutils.CountBytes(pl.fs, it.From)
			if err != nil {
				continue // an unreadable source surfaces below when we try to copy it
			}
			totalBytes += b
			totalFiles += n
		}
		j.SetTotals(totalBytes, totalFiles)
	} else {
		j.SetTotals(0, len(j.Items()))
	}

	isCopy := j.Kind() == jobs.KindCopy
	for i, it := range j.Items() {
		if err := ctx.Err(); err != nil {
			return err
		}
		j.StartFile(path.Base(it.From), it.To)

		// On a retry, clear any partial destination the original run's crash
		// left behind before re-copying (Stage 3 / D). A fresh run's dest is
		// already conflict-resolved upstream, so this only runs for retries.
		if pl.retry {
			_ = pl.fs.RemoveAll(it.To)
		}

		err := tm.transferItem(ctx, pl, j, isCopy, it)
		if err != nil {
			// Roll back this item's partial destination. The source is never
			// touched until a copy fully succeeds (MoveFileWithProgress), so a
			// canceled or failed transfer loses nothing.
			_ = pl.fs.RemoveAll(it.To)
			return err
		}

		switch {
		case isCopy:
			events.Publish(events.FileCopied{Base: pl.base, From: it.From, To: it.To})
		case looksLikeMove(it.From, it.To):
			events.Publish(events.FileMoved{Base: pl.base, From: it.From, To: it.To})
		default:
			events.Publish(events.FileRenamed{Base: pl.base, From: it.From, To: it.To})
		}

		// Mark the item done + persist so a Retry after a crash skips it
		// (partial-batch retry, Stage 3 / D).
		j.MarkItemDone(i)
		tm.reg.PersistProgress(j)
	}

	// Pin counters to totals so an instant (same-volume) rename — which copies
	// nothing — still reads 100% / N-of-N.
	j.MarkComplete()
	return nil
}

// transferItem runs one item's copy/move, auto-retrying a TRANSIENT failure up
// to twice with a short backoff before giving up (Stage 3 / D). A retry first
// clears the partial destination so the next attempt starts clean. Cancellation
// and non-transient errors (missing source, permission, no space) fail
// immediately — retrying them just wastes time.
func (tm *transferManager) transferItem(ctx context.Context, pl *transferPayload, j *jobs.Job, isCopy bool, it jobs.Item) error {
	const maxAttempts = 3 // original + 2 retries
	backoff := []time.Duration{200 * time.Millisecond, 500 * time.Millisecond}
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			_ = pl.fs.RemoveAll(it.To) // clear the failed attempt's partial
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff[attempt-1]):
			}
		}
		switch {
		case isCopy:
			err = fileutils.CopyWithProgress(ctx, pl.fs, it.From, it.To, pl.fileMode, pl.dirMode, j)
			// Verify the new copy against its source; a mismatch fails the item
			// (the source is untouched by a copy, so nothing is lost). Skipped on
			// a copy error — there's nothing valid to verify.
			if err == nil && pl.verify {
				err = fileutils.Verify(pl.fs, it.From, it.To)
			}
		case pl.verify:
			// MoveFileVerified keeps the source if the cross-volume copy doesn't
			// hash-match (the same-volume rename path is a no-op verify).
			err = fileutils.MoveFileVerified(ctx, pl.fs, it.From, it.To, pl.fileMode, pl.dirMode, j)
		default:
			err = fileutils.MoveFileWithProgress(ctx, pl.fs, it.From, it.To, pl.fileMode, pl.dirMode, j)
		}
		// A verification mismatch is a permanent failure — retrying re-reads the
		// same bad bytes — so it falls through to the non-transient return below.
		if err == nil || ctx.Err() != nil || !isTransientTransferErr(err) {
			return err
		}
	}
	return err
}

// isTransientTransferErr reports whether a copy/move failure is worth retrying:
// transient OS conditions (interrupted/again/busy/timeout), NOT permanent ones
// (missing source, permission, exists, no space, cross-device). Conservative —
// an unknown error is treated as permanent so we don't loop on a real fault.
func isTransientTransferErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EINTR) || errors.Is(err, syscall.EAGAIN) ||
		errors.Is(err, syscall.EBUSY) || errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	return false
}

type transferItemRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Overwrite bool   `json:"overwrite"`
	Rename    bool   `json:"rename"`
}

type transferRequest struct {
	Kind  string                `json:"kind"` // "move" | "copy"
	Items []transferItemRequest `json:"items"`
}

// jobsPostHandler starts a background transfer and returns the new job's initial
// snapshot. Conflict resolution mirrors the PATCH handler (the frontend has
// already resolved conflicts, so per-item overwrite/rename flags arrive set).
func (tm *transferManager) jobsPostHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		var req transferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return http.StatusBadRequest, err
		}

		var kind jobs.Kind
		switch req.Kind {
		case "move":
			kind = jobs.KindMove
			if !d.user.Perm.Rename {
				return http.StatusForbidden, nil
			}
		case "copy":
			kind = jobs.KindCopy
			if !d.user.Perm.Create {
				return http.StatusForbidden, nil
			}
		default:
			return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
		}
		if len(req.Items) == 0 {
			return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
		}

		items := make([]jobs.Item, 0, len(req.Items))
		for _, ri := range req.Items {
			src := path.Clean("/" + ri.From)
			dst := path.Clean("/" + ri.To)
			if src == "/" || dst == "/" {
				return http.StatusForbidden, nil
			}
			if !d.Check(src) || !d.Check(dst) {
				return http.StatusForbidden, nil
			}
			if err := checkParent(src, dst); err != nil {
				return http.StatusBadRequest, err
			}
			// A move whose source is its own bind mount can't complete: the
			// rename fails and the copy-fallback can't delete the mount point
			// afterwards, leaving a stranded duplicate. Refuse it up front.
			// (Copy is fine — it never removes the source.)
			if kind == jobs.KindMove {
				if isMount, ok := fileutils.IsMountpoint(d.user.Fs, src); ok && isMount {
					return http.StatusBadRequest, errMountpointMove(src)
				}
			}
			if !ri.Overwrite && !ri.Rename {
				if _, err := d.user.Fs.Stat(dst); err == nil {
					verb := "move"
					if kind == jobs.KindCopy {
						verb = "copy"
					}
					return http.StatusConflict, fmt.Errorf(
						"can't %s %q: an item with that name already exists in the destination folder",
						verb, path.Base(dst))
				}
			}
			if ri.Rename {
				dst = addVersionSuffix(dst, d.user.Fs)
			}
			// AFTER the suffix: a keep-both self-copy is legitimate (its dst
			// was just renamed apart). What's left — src == dst exactly — is a
			// transfer onto itself: a copy would TRUNCATE the source while
			// reading it (creating the destination clobbers the file being
			// copied), and a move is a pointless self-rename. Reject both.
			// (Case-only renames never come through here — they use the
			// synchronous PATCH path, which allows them via os.SameFile.)
			if src == dst {
				verb := "move"
				if kind == jobs.KindCopy {
					verb = "copy"
				}
				return http.StatusBadRequest, fmt.Errorf(
					"can't %s %q onto itself", verb, path.Base(src))
			}
			if ri.Overwrite && !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}
			items = append(items, jobs.Item{From: src, To: dst, Overwrite: ri.Overwrite, Rename: ri.Rename})
		}

		_, view := tm.schedule(r, d, kind, items, false)
		return renderJSON(w, r, view)
	})
}

// schedule builds the per-request execution payload and routes the items onto
// the right worker lane, returning the enqueued job + its initial snapshot.
// Shared by the POST and retry handlers. `retry` marks a re-run (clears partial
// destinations before re-copying). A same-volume move takes the FAST lane (an
// instant rename that needn't wait behind a long copy) and skips the prewalk
// (Stage 3 / G); everything else takes the main lane with a byte-counting
// prewalk. SameVolume is conservative, so anything uncertain falls back.
func (tm *transferManager) schedule(r *http.Request, d *data, kind jobs.Kind, items []jobs.Item, retry bool) (*jobs.Job, jobs.JobView) {
	fast := kind == jobs.KindMove
	if fast {
		for _, it := range items {
			same, err := fileutils.SameVolume(d.user.Fs, it.From, path.Dir(it.To))
			if err != nil || !same {
				fast = false
				break
			}
		}
	}

	pl := &transferPayload{
		fs:       d.user.Fs,
		fileMode: d.settings.FileMode,
		dirMode:  d.settings.DirMode,
		base:     eventBase(r, d),
		prewalk:  !fast, // a fast-lane rename copies nothing → no prewalk
		retry:    retry,
		verify:   d.settings.VerifyTransfers,
	}

	if fast {
		return tm.reg.EnqueueFast(d.user.ID, kind, items, pl)
	}
	j := tm.reg.Enqueue(d.user.ID, kind, items, pl)
	return j, j.Snapshot()
}

// jobsRetryHandler re-runs a failed/canceled/interrupted transfer's not-yet-done
// items as a fresh job, then drops the old one (Stage 3 / D). The payload is
// rebuilt from THIS authenticated request, so a job interrupted by a restart is
// retryable by its owner without the server holding any stale filesystem handle.
func (tm *transferManager) jobsRetryHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		oldID := mux.Vars(r)["id"]
		kind, items, ok := tm.reg.RetrySource(oldID, d.user.ID)
		if !ok {
			return http.StatusNotFound, nil
		}
		// Same permission gate as a fresh transfer of this kind.
		if kind == jobs.KindMove && !d.user.Perm.Rename {
			return http.StatusForbidden, nil
		}
		if kind == jobs.KindCopy && !d.user.Perm.Create {
			return http.StatusForbidden, nil
		}
		// Re-validate every item against the CURRENT rules: a retry is a fresh
		// authorized action, and for a job interrupted by a restart the user's
		// scope or permissions may have changed since the original transfer. The
		// paths are already conflict-resolved (To carries any "(1)" suffix), so we
		// only re-check scope + the overwrite-needs-Modify gate — not the
		// conflict/self-copy logic the POST handler runs on un-resolved input.
		for _, it := range items {
			if !d.Check(it.From) || !d.Check(it.To) {
				return http.StatusForbidden, nil
			}
			if it.Overwrite && !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}
		}
		_, view := tm.schedule(r, d, kind, items, true)
		tm.reg.Dismiss(oldID, d.user.ID) // drop the now-superseded job + its record
		return renderJSON(w, r, view)
	})
}

// jobsListHandler returns the caller's active + recently-finished transfers, so
// the UI can rehydrate the progress dock after a reload.
func (tm *transferManager) jobsListHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		return renderJSON(w, r, tm.reg.List(d.user.ID))
	})
}

// jobsGetHandler returns one transfer's snapshot (the poll target).
func (tm *transferManager) jobsGetHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		view, ok := tm.reg.Get(mux.Vars(r)["id"], d.user.ID)
		if !ok {
			return http.StatusNotFound, nil
		}
		return renderJSON(w, r, view)
	})
}

// jobsDeleteHandler cancels an active transfer or dismisses a finished one.
func (tm *transferManager) jobsDeleteHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		id := mux.Vars(r)["id"]
		if tm.reg.Cancel(id, d.user.ID) || tm.reg.Dismiss(id, d.user.ID) {
			return renderJSON(w, r, map[string]bool{"ok": true})
		}
		return http.StatusNotFound, nil
	})
}
