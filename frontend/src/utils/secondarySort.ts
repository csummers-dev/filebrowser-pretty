/**
 * Client-side listing sort (2.4.x).
 *
 * The listing's displayed order is decided HERE, on the client, so a sort
 * change is reflected instantly — the moment the user picks a field or flips
 * direction — instead of waiting on a server round-trip + silent reload. The
 * server still persists the primary sort (so it's the order on the next fresh
 * load) and sorts the initial payload, but the rendered order is this
 * function's. That means:
 *
 *   - Changing the primary sort re-orders the listing immediately.
 *   - A "None" secondary truly means none — no hidden tiebreaker survives.
 *
 * Called once per group (dirs and files render separately, dirs first), so
 * each group sorts in isolation and the dirs-first grouping is never
 * disturbed. Sorting copies (never `req.items` itself), so each item keeps its
 * original `index` — selection / keyboard-nav still map back correctly.
 *
 * Native Array.prototype.sort is stable (ES2019+), so two items equal on BOTH
 * criteria keep their incoming (server) order as the final tiebreaker.
 *
 * Comparator note: name / extension use localeCompare(numeric) — a close match
 * to the Go backend's natural.Less ("file_2" before "file_10"), and already
 * the comparator this codebase shipped for the secondary axis.
 */

/**
 * Comparator producing a negative / 0 / positive number for two items by the
 * given sort key (ascending). Returns 0 ("equal") rather than NaN/undefined for
 * any unparseable input or unexpected key — a comparator that returns NaN makes
 * Array.prototype.sort's order implementation-defined, so we never let one leak.
 */
/**
 * Resolve a folder's real recursive size (bytes) for size-sorting, or undefined
 * if it hasn't been fetched yet. Folders carry only their inode size in the
 * listing; the real size lives in the lazily-populated folder-size cache (the
 * same one the Size column reads), so the caller injects this lookup.
 */
export type FolderSizeResolver = (item: ResourceItem) => number | undefined;

function compareBy(
  a: ResourceItem,
  b: ResourceItem,
  by: SortKey,
  folderSize?: FolderSizeResolver
): number {
  switch (by) {
    case "name":
      return a.name.localeCompare(b.name, undefined, {
        numeric: true,
        sensitivity: "base",
      });
    case "size": {
      // Folders sort by their real recursive size (what the Size column shows),
      // NOT the listing's inode `size`. That size is fetched lazily and lives in
      // the folder-size cache, so the caller injects `folderSize`. A size that
      // hasn't resolved yet uses -1, so still-calculating folders cluster
      // together (name-broken) and snap into place as they load. With no
      // resolver — or when both are unknown/equal — fall back to name, so the
      // folder group stays predictable (matching the backend's bySize).
      // Dirs and files sort as separate groups, so a.isDir === b.isDir here.
      if (a.isDir && b.isDir) {
        if (folderSize) {
          const sa = folderSize(a) ?? -1;
          const sb = folderSize(b) ?? -1;
          if (sa !== sb) return sa - sb;
        }
        return a.name.localeCompare(b.name, undefined, {
          numeric: true,
          sensitivity: "base",
        });
      }
      const d = a.size - b.size;
      return Number.isNaN(d) ? 0 : d;
    }
    case "modified": {
      // Modified is an ISO 8601 string. Date.parse matches the user's
      // "time-based" mental model and is safe for ISO inputs; a malformed
      // value yields NaN, which we treat as a tie.
      const d = Date.parse(a.modified) - Date.parse(b.modified);
      return Number.isNaN(d) ? 0 : d;
    }
    case "extension":
      return a.extension.localeCompare(b.extension, undefined, {
        sensitivity: "base",
      });
    default:
      // Unknown / stale sort key (e.g. a pref left over from a removed
      // feature) — don't reorder.
      return 0;
  }
}

/**
 * Return a NEW array sorted by `primary`, breaking ties with `secondary`
 * (only when it's set and names a different key). The input array is not
 * mutated; arrays shorter than two items are returned as-is.
 */
export function sortListing(
  items: ResourceItem[],
  primary: SortCriterion,
  secondary: SortCriterion | null,
  folderSize?: FolderSizeResolver
): ResourceItem[] {
  if (items.length < 2) return items;

  // A secondary that repeats the primary key is a no-op tiebreaker (every
  // primary tie is also a secondary tie). The sort menu already disables that
  // combo; defend here too.
  const sec = secondary && secondary.by !== primary.by ? secondary : null;

  const sorted = [...items];
  sorted.sort((a, b) => {
    const p = compareBy(a, b, primary.by, folderSize);
    if (p !== 0) return primary.asc ? p : -p;
    if (sec) {
      const s = compareBy(a, b, sec.by, folderSize);
      if (s !== 0) return sec.asc ? s : -s;
    }
    // Equal on every active criterion — return 0 so the stable sort keeps
    // the incoming (server) order.
    return 0;
  });
  return sorted;
}
