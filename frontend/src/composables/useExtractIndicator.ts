import { useToast } from "vue-toastification";
import { useRouter } from "vue-router";
import { useFileStore } from "@/stores/file";
import { files as api } from "@/api";
import {
  mapUnzipError,
  isArchivePasswordError,
  buildBatchDests,
} from "@/utils/unzipErrors";
import { useArchivePassword } from "@/composables/useArchivePassword";

/**
 * Background archive extraction with floating toast feedback (mirrors the
 * move/copy transfer indicator). Extraction can take minutes for a large
 * archive; the old flow pinned the user to a blocking ExtractPanel overlay
 * the whole time. Instead the panel closes immediately on submit and the
 * work runs here — a delayed "Extracting…" toast, then a success/error
 * toast — so the user can keep navigating the app while the server works.
 *
 * Returns a `runExtract` closure that survives the (now-closed) panel
 * unmounting: it holds only app-level singletons (toast / router / store /
 * api), so the in-flight promise + toasts complete regardless of component
 * lifecycle.
 */
const PROGRESS_DELAY_MS = 300;

export interface ExtractParams {
  /** Archive resource URL (the `item.url` of the .zip/.7z/…). */
  sourceUrl: string;
  /** Display name for the toast. */
  name: string;
  /** Final, URL-encoded destination path the backend receives. */
  dest: string;
  overwrite: boolean;
  /** Delete the source archive after a successful extraction. */
  deleteOriginal: boolean;
  /** Navigate into the destination folder when done. */
  openFolder: boolean;
}

export function useExtractIndicator() {
  const toast = useToast();
  const router = useRouter();
  const fileStore = useFileStore();
  const { requestPassword } = useArchivePassword();

  const runExtract = async (params: ExtractParams): Promise<void> => {
    const { sourceUrl, name, dest, overwrite, deleteOriginal, openFolder } =
      params;

    // Detect-&-prompt: try with no password first; if the server replies 422
    // (password required / incorrect) prompt for one and retry. `password`
    // carries the last entered value; `triedPassword` flips the prompt into its
    // "Incorrect password" state on the second+ attempt.
    let password: string | undefined;
    let triedPassword = false;

    for (;;) {
      let progressId: string | number | undefined;
      const timer = window.setTimeout(() => {
        progressId = toast.info(`Extracting “${name}”…`, {
          timeout: false,
          closeButton: false,
          draggable: false,
        });
      }, PROGRESS_DELAY_MS);

      try {
        await api.unzip(sourceUrl, dest, overwrite, password);
      } catch (err) {
        // Clear the progress toast before we either prompt or report.
        window.clearTimeout(timer);
        if (progressId !== undefined) toast.dismiss(progressId);

        if (isArchivePasswordError(err)) {
          const entered = await requestPassword({ incorrect: triedPassword });
          if (entered == null) return; // user cancelled — quietly stop
          password = entered;
          triedPassword = true;
          continue; // retry with the supplied password
        }

        toast.error(mapUnzipError(err));
        return;
      }

      // Success.
      window.clearTimeout(timer);
      if (progressId !== undefined) toast.dismiss(progressId);

      // Delete the source only on extract success — a failed remove is a
      // warning, not a failure (the extraction itself worked).
      if (deleteOriginal) {
        try {
          await api.remove(sourceUrl);
        } catch (delErr) {
          toast.warning(
            `Extracted, but couldn't delete the original: ${
              delErr instanceof Error ? delErr.message : "unknown error"
            }`
          );
        }
      }
      toast.success(`Extracted to ${decodeURIComponent(dest)}`);
      if (openFolder) {
        void router.push({ path: dest.replace(/\/?$/, "/") });
      } else {
        // Refresh the current listing so a same-folder extraction's new
        // folder animates in (harmless if the user navigated elsewhere).
        fileStore.reload = true;
      }
      return;
    }
  };

  /**
   * Extract SEVERAL archives in one action (multi-select). Each archive
   * extracts into its OWN subfolder (derived from its name) under `base`, so
   * they never collide. Runs sequentially — the backend does one archive per
   * request — with a single progress toast and one summary at the end.
   * Password-protected archives prompt individually (a wrong/cancelled one is
   * skipped, the rest still run). `openFolder` navigates to `base` (which now
   * holds all the new subfolders) once, on success.
   */
  const runExtractBatch = async (
    items: Array<{ url: string; name: string }>,
    opts: {
      base: string;
      overwrite: boolean;
      deleteOriginal: boolean;
      openFolder: boolean;
    }
  ): Promise<void> => {
    const { base, overwrite, deleteOriginal, openFolder } = opts;
    const baseSlash = base.replace(/\/?$/, "/");
    const total = items.length;
    let ok = 0;

    // Per-archive destinations, de-duplicated so two archives that derive the
    // same subfolder name don't extract into each other.
    const dests = buildBatchDests(
      base,
      items.map((i) => i.name)
    );

    const progressId = toast.info(`Extracting ${total} archives…`, {
      timeout: false,
      closeButton: false,
      draggable: false,
    });

    for (let idx = 0; idx < items.length; idx++) {
      const it = items[idx];
      const dest = dests[idx];
      let password: string | undefined;
      let triedPassword = false;
      let extracted = false;

      for (;;) {
        try {
          await api.unzip(it.url, dest, overwrite, password);
          extracted = true;
          ok++;
          break;
        } catch (err) {
          if (isArchivePasswordError(err)) {
            const entered = await requestPassword({ incorrect: triedPassword });
            if (entered == null) break; // skip this one, keep going
            password = entered;
            triedPassword = true;
            continue;
          }
          toast.error(`${it.name}: ${mapUnzipError(err)}`);
          break;
        }
      }

      // Delete the source only when its own extraction succeeded.
      if (extracted && deleteOriginal) {
        try {
          await api.remove(it.url);
        } catch {
          /* non-fatal — the extraction itself worked */
        }
      }
    }

    toast.dismiss(progressId);
    if (ok > 0) {
      toast.success(
        ok === total
          ? `Extracted ${ok} archives to ${decodeURIComponent(base)}`
          : `Extracted ${ok} of ${total} archives to ${decodeURIComponent(base)}`
      );
    }
    if (openFolder && ok > 0) {
      void router.push({ path: baseSlash });
    } else {
      fileStore.reload = true;
    }
  };

  return { runExtract, runExtractBatch };
}
