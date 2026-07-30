import { StatusError } from "@/api/utils";
import { archiveBaseName } from "@/utils/archive";

/**
 * Map a zip-extraction failure into a single line of user-facing copy.
 *
 * The backend (PR #5746, `http/unzip.go`) returns plain-text error bodies
 * pulled from `errors/errors.go` (e.g. "the zip file is too large"). We
 * match on those strings so the UI can surface something more meaningful
 * than the raw backend text. Falls back to a generic catch-all for
 * unknown payloads so a bug in the mapper can't silence the error
 * entirely.
 *
 * If the backend ever changes its error wording these mappings will
 * silently regress to the catch-all — keep this file in lockstep with
 * the package-level error vars in `errors/errors.go`.
 */
export function mapUnzipError(err: unknown): string {
  // Network / transport-level failures land here as plain Error.
  if (err instanceof StatusError) {
    const msg = (err.message ?? "").toLowerCase();

    if (msg.includes("the archive is too large")) {
      return "Archive is too large to extract (server limit reached).";
    }
    if (msg.includes("too high a decompression rate")) {
      return "Possible zip bomb detected — extraction blocked.";
    }
    if (msg.includes("invalid path in some files")) {
      return "Archive contains unsafe file paths.";
    }
    if (msg.includes("too high a decompression size")) {
      return "Archive contents would exceed the server's size limit.";
    }
    if (msg.includes("some files are invalid in the archive")) {
      return "Archive is corrupt or malformed.";
    }
    if (msg.includes("this archive format isn't supported")) {
      return "This archive format can't be extracted.";
    }
    if (msg.includes("split or multi-volume archives")) {
      return "Split / multi-volume archives of this format aren't supported.";
    }
    if (msg.includes("password-protected archives")) {
      return "Password-protected archives aren't supported.";
    }
    // Extraction now accepts a password (gathered by the prompt-and-retry loop
    // in useExtractIndicator); these only surface on non-prompt callers.
    if (msg.includes("this archive is password-protected")) {
      return "This archive is password-protected.";
    }
    if (msg.includes("the archive password is incorrect")) {
      return "Incorrect password — try again.";
    }

    // Status-based fallbacks for non-text payloads.
    const status = err.status ?? 0;
    if (status === 403) {
      return "You don't have permission to extract here.";
    }
    if (status === 404) {
      return "Archive not found.";
    }
    if (status >= 500) {
      return "Extraction failed — check the server logs for details.";
    }
    if (status === 400) {
      return "The archive couldn't be extracted.";
    }
  }

  if (err instanceof Error && err.message) {
    return err.message;
  }

  return "Extraction failed.";
}

/**
 * Whether an extraction failure means "the server needs a (correct) archive
 * password" — the signal to prompt the user and retry. The backend returns
 * HTTP 422 for both "password required" and "password incorrect" (NOT 401,
 * which fetchURL treats as session-expiry and force-logs-out on), so the status
 * alone is an unambiguous trigger.
 */
export function isArchivePasswordError(err: unknown): boolean {
  return err instanceof StatusError && err.status === 422;
}

/**
 * Derive the auto-named subfolder for "extract into a new folder" by
 * stripping the archive suffix. Delegates to the shared `archiveBaseName`
 * helper, which understands every supported format (zip / 7z / rar incl.
 * multi-volume parts / tar family), e.g. `backup.tar.gz` → `backup`.
 */
export function deriveSubfolderName(filename: string): string {
  return archiveBaseName(filename);
}

/**
 * Destination paths for a MULTI-archive extraction: each archive extracts into
 * its own subfolder (derived from its name) under `base`. Subfolder names are
 * de-duplicated case-insensitively — if two selected archives derive the same
 * folder (e.g. `data.zip` + `data.7z`, or the same name twice) the later ones
 * get a ` (2)`, ` (3)`… suffix so they never extract into each other. Returns
 * one URL-encoded destination path per input name, in order. A name that
 * derives an empty base falls back to `archive`.
 */
export function buildBatchDests(base: string, names: string[]): string[] {
  const baseSlash = base.replace(/\/?$/, "/");
  const used = new Set<string>();
  return names.map((name) => {
    let sub = deriveSubfolderName(name) || "archive";
    if (used.has(sub.toLowerCase())) {
      let n = 2;
      while (used.has(`${sub} (${n})`.toLowerCase())) n++;
      sub = `${sub} (${n})`;
    }
    used.add(sub.toLowerCase());
    return baseSlash + encodeURIComponent(sub);
  });
}
