<template>
  <aside
    v-if="paneVisible"
    class="info-pane w-[320px] shrink-0 border-l border-line bg-canvas flex flex-col overflow-hidden max-lg:fixed max-lg:top-12 max-lg:right-0 max-lg:bottom-0 max-lg:h-[calc(100vh-48px)] max-lg:z-40 max-lg:shadow-2xl max-[540px]:left-16 max-[540px]:w-auto"
  >
    <!-- Pane header (no divider — keeps the panel calm + sidebar-like) -->
    <div class="h-11 px-4 flex items-center justify-between shrink-0">
      <div class="text-[12px] font-semibold text-ink-2">Details</div>
      <button
        v-if="isMobile || !collapsed"
        @click="isMobile ? close() : toggleCollapse()"
        class="w-6 h-6 rounded hover:bg-hover flex items-center justify-center text-[var(--color-ink-2)] transition"
        :title="isMobile ? 'Close' : 'Collapse panel'"
        :aria-label="isMobile ? 'Close info pane' : 'Collapse info pane'"
      >
        <Icon :name="isMobile ? 'x' : 'panel-right-close'" :size="14" />
      </button>
    </div>

    <!-- Scrollable body. The header and footer stay pinned (shrink-0); this
         middle region owns the scroll. `min-h-0` lets it shrink below its
         content's natural height so overflow-y engages here instead of the
         details spilling past the bottom of the pane uncut. It is NOT a flex
         container — a plain block so its children keep their natural height
         and never get squished by flex-shrink. -->
    <div class="ip-scroll flex-1 min-h-0 overflow-y-auto">
      <!-- ── State 1: a single item is selected (RC-4) ─────────────── -->
      <template v-if="item">
        <!-- Preview -->
        <div class="px-4 pt-2 pb-3">
          <div
            class="aspect-[4/3] rounded-lg border border-line overflow-hidden relative flex items-center justify-center preview-mesh"
          >
            <!-- Photo / video / PDF thumbnail — fill the whole frame. -->
            <img
              v-if="hasThumbnail"
              :src="thumbnailUrl"
              class="w-full h-full object-cover"
              :alt="item.name"
            />
            <!-- RC-42: embedded album art for a selected music file. -->
            <img
              v-else-if="albumArtUrl"
              :src="albumArtUrl"
              class="w-full h-full object-cover"
              alt="Album artwork"
            />
            <!-- batch #4: embedded EPUB cover for a selected book, so the
               details pane matches the preview pane. -->
            <img
              v-else-if="epubCoverUrl"
              :src="epubCoverUrl"
              class="max-w-full max-h-full w-auto h-auto object-contain rounded-md shadow-sm"
              alt="Book cover"
            />
            <!-- No image preview (folders, archives, code, blobs, …): fill the
               whole frame with the file's vivid color tile + matching icon,
               mirroring the grid / gallery tile styling. -->
            <div v-else class="info-tile" :class="iconColorClass">
              <Icon
                :name="iconName"
                :size="46"
                :stroke-width="1.5"
                :fill="item.isDir ? 'currentColor' : 'none'"
              />
            </div>
          </div>
        </div>

        <!-- Title -->
        <div class="px-4 pb-3">
          <div
            class="info-name text-[15px] font-semibold text-ink-1 leading-snug"
          >
            <template v-for="(seg, i) in nameSegments" :key="i"
              >{{ seg }}<wbr
            /></template>
          </div>
          <div class="text-[12px] text-ink-3 mt-0.5 tabular">
            {{ typeLabel }}<span v-if="!item.isDir"> · {{ sizeLabel }}</span>
          </div>
        </div>

        <!-- Primary action grid: Share / Download / Rename / Delete — fixed positions -->
        <div class="px-4 pb-2 grid grid-cols-4 gap-1.5">
          <button
            v-if="canShare"
            @click="action('share')"
            class="info-action info-action--share"
            title="Share"
          >
            <Icon name="share" :size="14" />
            <span>Share</span>
          </button>
          <button
            v-if="canDownload"
            @click="download"
            class="info-action info-action--download"
            title="Download"
          >
            <Icon name="download" :size="14" />
            <span>Download</span>
          </button>
          <button
            v-if="canRename"
            @click="action('rename')"
            class="info-action info-action--rename"
            title="Rename"
          >
            <Icon name="pencil" :size="14" />
            <span>Rename</span>
          </button>
          <button
            v-if="canDelete"
            @click="action('delete')"
            class="info-action info-action--danger"
            title="Delete"
          >
            <Icon name="trash-2" :size="14" />
            <span>Delete</span>
          </button>
        </div>

        <!-- Secondary action row: Move / Copy / (Preview or Extract).
         For files this is a clean 3-up grid; for folders Preview/Extract
         are omitted and the row falls back to 2-col so Move + Copy still
         fill the width. The third tile swaps to "Extract" when the file
         is a supported archive (zip / 7z / rar / tar family) and the user
         can create — an archive's "preview" really is its contents, which
         is what Extract opens. -->
        <div
          v-if="canMove || canCopy || !item.isDir"
          class="pb-4 grid gap-1.5"
          :class="
            !item.isDir
              ? canEditTags
                ? 'px-4 grid-cols-4'
                : 'px-4 grid-cols-3'
              : 'grid-cols-2 w-2/3 mx-auto'
          "
        >
          <button
            v-if="canMove"
            @click="action('move')"
            class="info-action info-action--move"
            title="Move"
          >
            <Icon name="forward" :size="14" />
            <span>Move</span>
          </button>
          <button
            v-if="canCopy"
            @click="action('copy')"
            class="info-action info-action--copy"
            title="Copy"
          >
            <Icon name="copy" :size="14" />
            <span>Copy</span>
          </button>
          <!-- V2 #5: edit-tags joins the action row (between Copy and Preview),
             mirroring the preview details panel, instead of a separate button. -->
          <button
            v-if="canEditTags"
            @click="action('audio-tags')"
            class="info-action info-action--tags"
            title="Edit tags"
          >
            <Icon name="music" :size="14" />
            <span>Edit tags</span>
          </button>
          <button
            v-if="!item.isDir && canExtract"
            @click="action('extract')"
            class="info-action info-action--extract"
            title="Extract"
          >
            <Icon name="package-open" :size="14" />
            <span>Extract</span>
          </button>
          <button
            v-else-if="!item.isDir"
            @click="open"
            class="info-action info-action--preview"
            title="Preview"
          >
            <Icon name="eye" :size="14" />
            <span>Preview</span>
          </button>
        </div>
        <div v-else class="pb-4"></div>

        <!-- Properties -->
        <div class="px-4 py-3">
          <div
            class="text-[11px] font-semibold text-ink-3 uppercase tracking-[0.06em] mb-2"
          >
            Properties
          </div>
          <dl class="text-[12px] space-y-1.5">
            <div class="flex justify-between gap-3">
              <dt class="text-ink-3">Type</dt>
              <dd class="text-ink-1">{{ typeLabel }}</dd>
            </div>
            <div class="flex justify-between gap-3">
              <dt class="text-ink-3">Size</dt>
              <dd class="text-ink-1 tabular">
                {{ item.isDir ? folderSizeLabel : sizeLabel }}
              </dd>
            </div>
            <div class="flex justify-between gap-3">
              <dt class="text-ink-3">Modified</dt>
              <dd class="text-ink-1 tabular">{{ modifiedLabel }}</dd>
            </div>
            <div v-if="item.extension" class="flex justify-between gap-3">
              <dt class="text-ink-3">Extension</dt>
              <dd class="text-ink-1 font-mono text-[11px]">
                {{ item.extension }}
              </dd>
            </div>
          </dl>
        </div>

        <!-- WS7 #8: read-only audio tags for a single selected music file.
           Non-empty fields only; refreshes after an edit (watch keys on
           modified). Distinct from the v1.3 user "Tags" section below. -->
        <div v-if="audioTagRows.length > 0" class="px-4 py-3">
          <div
            class="text-[11px] font-semibold text-ink-3 uppercase tracking-[0.06em] mb-2"
          >
            Audio tags
          </div>
          <dl class="text-[12px] space-y-1.5">
            <div
              v-for="row in audioTagRows"
              :key="row.label"
              class="flex justify-between gap-3"
            >
              <dt class="text-ink-3 shrink-0">{{ row.label }}</dt>
              <dd class="text-ink-1 text-right break-words min-w-0">
                {{ row.value }}
              </dd>
            </div>
          </dl>
        </div>

        <!-- Tags (v1.3 S2-5). Full chip list + "Manage tags" CTA that
         opens the picker SlideOver. Empty state shows the CTA on its
         own. Reads tags from the cache populated by FileListing's
         batch fetch; falls back to forFile when stale. -->
        <div class="px-4 py-3">
          <div class="mb-1.5 flex items-center justify-between gap-2">
            <div
              class="text-[11px] font-semibold text-ink-3 uppercase tracking-[0.06em]"
            >
              Tags
            </div>
            <button
              type="button"
              class="info-pane__manage-tags"
              :title="itemTags.length === 0 ? 'Add tags' : 'Manage tags'"
              @click="tagPicker.open"
            >
              <Icon
                :name="itemTags.length === 0 ? 'plus' : 'pencil'"
                :size="11"
              />
              <span>{{ itemTags.length === 0 ? "Add" : "Manage" }}</span>
            </button>
          </div>
          <div v-if="itemTags.length > 0" class="info-pane__tag-list">
            <TagChip
              v-for="t in itemTags"
              :key="t.id"
              :tag="t"
              size="md"
              :focusable="false"
            />
          </div>
          <div v-else class="info-pane__tag-empty">No tags</div>
        </div>

        <!-- Location -->
        <div class="px-4 py-3">
          <div class="mb-1.5 flex items-center justify-between gap-2">
            <div
              class="text-[11px] font-semibold text-ink-3 uppercase tracking-[0.06em]"
            >
              Location
            </div>
            <!-- G6: Copy path action. Lives next to the Location label so
             the affordance is right where the user is looking when they
             want it. Uses the navigator.clipboard API (modern, no fallback
             needed for our target browsers). Toast confirms because the
             OS clipboard is invisible — the user needs feedback to know
             something happened. -->
            <button
              type="button"
              class="info-pane__copy-path"
              :title="copiedPath ? 'Copied!' : 'Copy path'"
              :aria-label="copiedPath ? 'Path copied' : 'Copy path'"
              @click="copyPath"
            >
              <Icon :name="copiedPath ? 'check' : 'copy'" :size="12" />
              <span>{{ copiedPath ? "Copied" : "Copy" }}</span>
            </button>
          </div>
          <div
            class="font-mono text-[11px] text-ink-2 break-all bg-elevated rounded-md px-2 py-1.5 border border-line"
          >
            {{ item.path || item.url }}
          </div>
        </div>
      </template>

      <!-- ── State 2: 2+ items selected ─────────────────────────────── -->
      <template v-else-if="selectedCount >= 2">
        <div class="ip-summary">
          <div class="ip-summary__icon ip-summary__icon--accent">
            <Icon name="layers" :size="26" :stroke-width="1.4" />
          </div>
          <div class="ip-summary__title">
            {{ selectedCount }} items selected
          </div>
          <div class="ip-summary__meta">{{ selectionSizeLabel }} total</div>
          <p class="ip-summary__hint">
            Use the action bar at the bottom of the list to move, copy, rename
            or delete the selection.
          </p>
        </div>
      </template>

      <!-- ── State 3: nothing selected → current folder summary ─────── -->
      <template v-else>
        <div class="ip-summary">
          <div class="ip-summary__icon ip-summary__icon--folder">
            <Icon
              name="folder"
              :size="26"
              :stroke-width="1.4"
              fill="currentColor"
            />
          </div>
          <div class="ip-summary__title">{{ folderName }}</div>
          <div class="ip-summary__meta">
            {{ folderItemCount }} {{ folderItemCount === 1 ? "item" : "items" }}
          </div>
          <dl class="ip-summary__props">
            <div class="ip-summary__prop">
              <dt>Folders</dt>
              <dd class="tabular">{{ folderDirCount }}</dd>
            </div>
            <div class="ip-summary__prop">
              <dt>Files</dt>
              <dd class="tabular">{{ folderFileCount }}</dd>
            </div>
            <div class="ip-summary__prop ip-summary__prop--path">
              <dt>Path</dt>
              <dd class="ip-summary__path" :title="folderPath">
                {{ folderPath }}
              </dd>
            </div>
          </dl>
          <p class="ip-summary__hint">Select an item to see its details.</p>
        </div>
      </template>
    </div>
    <!-- /ip-scroll -->

    <!-- Keyboard hint footer — only meaningful with a selection -->
    <div
      v-if="item"
      class="px-4 py-2.5 text-[11px] text-ink-3 flex items-center justify-between shrink-0"
    >
      <span class="flex items-center gap-1.5">
        <kbd
          class="font-mono text-[10px] px-1 py-px rounded border border-line bg-elevated"
          >Esc</kbd
        >
        <span>to clear</span>
      </span>
    </div>

    <!-- Tag picker sheet (v1.3 S2-5). Renders as a SlideOver via
         Teleport so it sits above the InfoPane in stacking order. -->
    <TagPickerSheet
      v-if="item"
      :open="tagPicker.isOpen.value"
      :path="item.url"
      @cancel="tagPicker.close"
      @saved="tagPicker.close"
    />
  </aside>

  <!-- Collapsed rail (desktop): a thin strip with an expand affordance
       so the pane can be reclaimed without losing the entry point. -->
  <button
    v-else-if="showRail"
    class="info-pane-rail shrink-0 border-l border-line bg-canvas flex items-start justify-center pt-3 text-[var(--color-ink-2)] hover:bg-hover transition"
    title="Show details"
    aria-label="Show details panel"
    @click="toggleCollapse"
  >
    <Icon name="panel-left-open" :size="16" />
  </button>
</template>

<script setup lang="ts">
import Icon from "@/components/Icon.vue";
import TagChip from "@/components/TagChip.vue";
import TagPickerSheet from "@/components/files/TagPickerSheet.vue";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useTagsStore } from "@/stores/tags";
import { fileIcon, fileIconColor } from "@/utils/fileIcon";
import { filesize } from "@/utils";
import {
  enableThumbs,
  enableVideoThumbs,
  enablePdfThumbs,
  unzipEnabled,
} from "@/utils/constants";
import { isExtractable } from "@/utils/archive";
import { isAudioTaggable } from "@/utils/audio";
import { files as api, audiotags as audioTagsApi } from "@/api";
import type { Book } from "epubjs";
import dayjs from "dayjs";
import { computed, inject, onMounted, onUnmounted, ref, watch } from "vue";
import { useTagPicker } from "@/composables/useTagPicker";
import { usePreferences } from "@/composables/usePreferences";
import { useFolderSizes } from "@/composables/useFolderSizes";

const fileStore = useFileStore();
const tagsStore = useTagsStore();

// Tag picker open-state (v1.3 S2-5; hoisted to a singleton composable
// in S4-1 so the row context menu can also open it). The actual sheet
// component still lives in this file — only the flag moved out.
const tagPicker = useTagPicker();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

// G6: Copy-path local state. We flash the button label to "Copied" for
// ~1.5 s after a successful copy as immediate inline feedback, and ALSO
// show a toast. Double-feedback is intentional: the inline change is
// instant and right where the user clicked, the toast is durable so it
// survives them looking away.
const copiedPath = ref<boolean>(false);
let copiedResetTimer: number | null = null;

const copyPath = async () => {
  const path = item.value?.path || item.value?.url;
  if (!path) return;
  try {
    await navigator.clipboard.writeText(path);
    copiedPath.value = true;
    if (copiedResetTimer !== null) window.clearTimeout(copiedResetTimer);
    copiedResetTimer = window.setTimeout(() => {
      copiedPath.value = false;
      copiedResetTimer = null;
    }, 1500);
    $showSuccess(`Path copied: ${path}`);
  } catch {
    // navigator.clipboard fails in insecure contexts (http://) or when
    // the user denied permission. Fall back to a manual selection prompt
    // via a one-off textarea + execCommand("copy") so the action still
    // succeeds on LAN-only HTTP setups (homelab without TLS).
    try {
      const ta = document.createElement("textarea");
      ta.value = path;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
      copiedPath.value = true;
      if (copiedResetTimer !== null) window.clearTimeout(copiedResetTimer);
      copiedResetTimer = window.setTimeout(() => {
        copiedPath.value = false;
        copiedResetTimer = null;
      }, 1500);
      $showSuccess(`Path copied: ${path}`);
    } catch {
      /* both methods failed — silently no-op */
    }
  }
};

const item = computed<ResourceItem | null>(() => {
  if (!fileStore.req || fileStore.selectedCount !== 1) return null;
  const idx = fileStore.selected[0];
  return fileStore.req.items[idx] ?? null;
});

// Folder size (2.4.0 Stage 4 / E): folders have no intrinsic size, so fetch the
// recursive total on demand when a single folder is selected and show it in the
// Size row ("Calculating…" until it resolves). Keyed by path + mtime so a folder
// whose contents changed re-fetches. The fetch + cache live in the composable.
const folderSizes = useFolderSizes();
const folderSizeKey = (it: ResourceItem) => String(it.modified ?? "");
watch(
  item,
  (it) => {
    if (it?.isDir && it.url) void folderSizes.ensure(it.url, folderSizeKey(it));
  },
  { immediate: true }
);
const folderSizeLabel = computed(() => {
  const it = item.value;
  if (!it?.isDir || !it.url) return "";
  const s = folderSizes.cached(it.url, folderSizeKey(it));
  return s === undefined ? "Calculating…" : filesize(s);
});

// ── RC-4: always-present details pane ───────────────────────────────
// The pane stays docked so the listing never reflows when an item is
// selected (which used to shift tiles mid-double-click and made
// double-click-to-open impossible in grid/gallery). It shows three
// states: a single item, a multi-select summary, or — when nothing is
// selected — a summary of the current folder.
const prefs = usePreferences();

const selectedCount = computed(() => fileStore.selectedCount);

// Below lg the pane reverts to an on-select overlay (an always-docked
// 320px panel would dominate a phone). Tracked via matchMedia.
const isMobile = ref(false);
const updateIsMobile = () => {
  isMobile.value =
    typeof window !== "undefined" &&
    window.matchMedia("(max-width: 1023px)").matches;
};

// Desktop collapse-to-rail, persisted across sessions. Calm Minimal: defaults to
// collapsed (a slim rail) so an empty 320px column isn't parked.
//
// The pane's visibility follows ONLY this collapsed/expanded toggle — selecting a
// file or folder no longer force-opens it, so clicking around the listing never
// interrupts the user's flow. Selection drives the pane's CONTENT (which details
// to show) when it IS open; it no longer decides whether it's open.
const collapsed = ref<boolean>(
  prefs.get<boolean>("detailsPaneCollapsed", true)
);
const toggleCollapse = () => {
  // A plain pin toggle: the rail's expand affordance opens it, the open pane's
  // chevron collapses it. The state sticks regardless of what's selected.
  collapsed.value = !collapsed.value;
  void prefs.set("detailsPaneCollapsed", collapsed.value);
};

// Full pane (desktop): shown only when the user has expanded it (!collapsed),
// independent of selection. Mobile: opt-in overlay. Collapsed rail: desktop,
// whenever collapsed.
const paneVisible = computed(
  () =>
    (!isMobile.value && !collapsed.value) ||
    // Mobile: opt-in only. The sheet used to auto-open on any selection,
    // which covered the row and hijacked the user's next tap; now it opens
    // exclusively from the selection toolbar's Details button.
    (isMobile.value && layoutStore.mobileDetailsOpen)
);
const showRail = computed(() => !isMobile.value && collapsed.value);

// Folder summary (nothing selected).
const folderName = computed(() => fileStore.req?.name || "Home");
const folderDirCount = computed(() => fileStore.req?.numDirs ?? 0);
const folderFileCount = computed(() => fileStore.req?.numFiles ?? 0);
const folderItemCount = computed(
  () => folderDirCount.value + folderFileCount.value
);
const folderPath = computed(() => fileStore.req?.path ?? "/");

// Multi-select summary (2+ selected).
const selectionSizeLabel = computed(() => {
  const req = fileStore.req;
  if (!req) return "";
  const total = fileStore.selected.reduce(
    (sum, idx) => sum + (req.items[idx]?.size ?? 0),
    0
  );
  return filesize(total);
});

// Tag list for the currently-selected item. Reads from the same
// byPath cache the listing rows use → no extra HTTP, and the chip
// set stays consistent between the row + the info pane. V2 #23: ordered
// alphabetically by name, matching the left-to-right order of the row dots.
const itemTags = computed<Tag[]>(() => {
  if (!item.value) return [];
  return [...tagsStore.forPath(item.value.url)].sort((a, b) =>
    a.name.localeCompare(b.name)
  );
});

const iconName = computed(() =>
  item.value
    ? fileIcon({
        isDir: item.value.isDir,
        type: item.value.type,
        name: item.value.name,
      })
    : "file"
);

const iconColorClass = computed(() =>
  item.value
    ? fileIconColor({
        isDir: item.value.isDir,
        type: item.value.type,
        name: item.value.name,
      })
    : "bg-zinc-100 text-zinc-600"
);

const hasThumbnail = computed(() => {
  if (!enableThumbs || !item.value || item.value.isDir) return false;
  // Mirror the listing's thumbnail support so the details preview shows the
  // same backend-rendered thumbnails the grid does — images always, video +
  // PDF when the server supports them. Audio falls through to album art, then
  // to the color tile.
  const t = item.value.type;
  if (t === "image") return true;
  if (t === "video") return enableVideoThumbs;
  if (t === "pdf") return enablePdfThumbs;
  // V3-E #20: comics get a server-extracted cover (first page) — the same
  // thumbnail the row icon shows. Mirror the listing's showThumbnail gate so
  // the details frame shows the cover instead of the generic color tile.
  // (.cbz joined .cbr in v2.7 — the original exclusion was reversed.)
  const n = (item.value.name || "").toLowerCase();
  if (n.endsWith(".cbr") || n.endsWith(".cbz")) return true;
  return false;
});

// Break long file names at separators (., -, _, /, space) instead of mid-word:
// each segment is rendered with a trailing <wbr> so wrapping prefers those
// natural points and only breaks a long run as a last resort.
const nameSegments = computed(() =>
  (item.value?.name ?? "").split(/(?<=[._\-\s/])/)
);

const thumbnailUrl = computed(() => {
  if (!item.value) return "";
  return api.getPreviewURL(item.value as Resource, "thumb");
});

// ── Album art for a selected music file (RC-42) ─────────────────────
// Mirrors AudioViewer's approach (E2): range-fetch the file header
// (ID3v2 + APIC live in the first few hundred KB) and parse it client-
// side. `music-metadata` is dynamically imported so it never enters the
// always-loaded listing bundle — only when an audio file is selected.
const albumArtUrl = ref<string>("");
let albumArtRevocable: string | null = null;
let albumArtToken = 0;
let albumArtAbort: AbortController | null = null;
// V3-E #17: match AudioViewer — 512 KB sliced large OGG covers in half (the
// Vorbis-comment picture is base64-inflated and overruns the header window).
// 2 MB captures the full cover; the fetch is aborted the moment the selection
// moves on, so browsing isn't slowed by the larger window.
const AUDIO_META_RANGE = 2 * 1024 * 1024; // 2 MB

const isAudio = computed(
  () => !!item.value && !item.value.isDir && item.value.type === "audio"
);

const clearAlbumArt = () => {
  // Cancel any in-flight range fetch so a rapid selection sweep doesn't leave
  // a stack of pending 512 KB requests (RC review #4).
  if (albumArtAbort) {
    albumArtAbort.abort();
    albumArtAbort = null;
  }
  if (albumArtRevocable) {
    URL.revokeObjectURL(albumArtRevocable);
    albumArtRevocable = null;
  }
  albumArtUrl.value = "";
};

const loadAlbumArt = async (resource: ResourceItem) => {
  const token = ++albumArtToken;
  clearAlbumArt(); // also aborts any previous in-flight fetch
  const ctrl = new AbortController();
  albumArtAbort = ctrl;
  try {
    const url = api.getDownloadURL(resource, true);
    const res = await fetch(url, {
      headers: { Range: `bytes=0-${AUDIO_META_RANGE - 1}` },
      signal: ctrl.signal,
    });
    if (token !== albumArtToken) return; // selection moved on
    if (!res.ok && res.status !== 206) return;
    const blob = await res.blob();
    if (token !== albumArtToken) return;
    const { parseBlob } = await import("music-metadata");
    const parsed = await parseBlob(blob);
    if (token !== albumArtToken) return;
    const pic = parsed.common.picture?.[0];
    if (!pic) return; // no embedded art → keep the generic music icon
    const artBlob = new Blob([new Uint8Array(pic.data)], { type: pic.format });
    const objUrl = URL.createObjectURL(artBlob);
    albumArtRevocable = objUrl;
    albumArtUrl.value = objUrl;
  } catch {
    /* parse failure / aborted — fall back to the generic icon */
  }
};

// Re-extract whenever the selected audio file changes; clear for anything
// that isn't a single audio file.
watch(
  () => (isAudio.value ? (item.value?.url ?? null) : null),
  (url) => {
    if (url) void loadAlbumArt(item.value as ResourceItem);
    else clearAlbumArt();
  },
  { immediate: true }
);

onUnmounted(clearAlbumArt);

// ── WS7 #6 + #8: audio tag entry point + read-only tag display ──────
// A single selected MP3/FLAC/… file. #6 surfaces an "Edit tags" button
// (gated on perm.modify); #8 lists its non-empty tags read-only (no perm
// needed — reading is harmless and useful even without edit rights).
const isTaggableAudioFile = computed(
  () => !!item.value && !item.value.isDir && isAudioTaggable(item.value.name)
);
const canEditTags = computed(
  () => isTaggableAudioFile.value && !!authStore.user?.perm.modify
);

const audioTagRows = ref<Array<{ label: string; value: string }>>([]);
let audioTagsToken = 0;
const AUDIO_TAG_FIELDS: Array<{ key: string; label: string }> = [
  { key: "title", label: "Title" },
  { key: "artist", label: "Artist" },
  { key: "album", label: "Album" },
  { key: "albumArtist", label: "Album artist" },
  { key: "year", label: "Year" },
  { key: "track", label: "Track" },
  { key: "disc", label: "Disc" },
  { key: "genres", label: "Genre" },
  { key: "composer", label: "Composer" },
];

const clearAudioTags = () => {
  audioTagsToken++;
  audioTagRows.value = [];
};

const loadAudioTags = async (resource: ResourceItem) => {
  const token = ++audioTagsToken;
  audioTagRows.value = [];
  const path = resource.path || resource.url;
  if (!path) return;
  try {
    const results = await audioTagsApi.read([path]);
    if (token !== audioTagsToken) return;
    const tags = results[0]?.tags;
    if (!tags) return;
    const rows: Array<{ label: string; value: string }> = [];
    for (const f of AUDIO_TAG_FIELDS) {
      const raw =
        f.key === "genres"
          ? (tags.genres ?? []).join(", ")
          : ((tags as unknown as Record<string, string>)[f.key] ?? "");
      if (raw && raw.trim()) rows.push({ label: f.label, value: raw });
    }
    audioTagRows.value = rows;
  } catch {
    /* read failed (corrupt / unsupported) — show nothing */
  }
};

// Re-read when the selected taggable audio file changes OR its modified time
// bumps (a tag save), so the read-only list never goes stale after an edit.
watch(
  () =>
    isTaggableAudioFile.value
      ? `${item.value?.url ?? ""}#${item.value?.modified ?? ""}`
      : null,
  (key) => {
    if (key && item.value) void loadAudioTags(item.value as ResourceItem);
    else clearAudioTags();
  },
  { immediate: true }
);
onUnmounted(clearAudioTags);

// ── Embedded cover art for a selected EPUB (batch #4) ───────────────
// The Preview pane shows the cover because it mounts EpubViewer, which
// loads epub.js and calls book.coverUrl(). The details pane never mounts
// the reader, so it had no cover. Mirror the album-art approach: lazily
// import epub.js (kept OUT of the always-loaded listing bundle via the
// dynamic import — it stays in the shared epubjs async chunk), open the
// book just long enough to resolve its cover into our OWN blob URL, then
// destroy the book so the archive doesn't linger in memory.
const epubCoverUrl = ref<string>("");
let epubCoverRevocable: string | null = null;
let epubCoverToken = 0;

const isEpub = computed(
  () =>
    !!item.value &&
    !item.value.isDir &&
    item.value.extension?.toLowerCase() === ".epub"
);

const clearEpubCover = () => {
  // Bump the token so any in-flight load drops its result on return.
  epubCoverToken++;
  if (epubCoverRevocable) {
    URL.revokeObjectURL(epubCoverRevocable);
    epubCoverRevocable = null;
  }
  epubCoverUrl.value = "";
};

const loadEpubCover = async (resource: ResourceItem) => {
  const token = ++epubCoverToken;
  if (epubCoverRevocable) {
    URL.revokeObjectURL(epubCoverRevocable);
    epubCoverRevocable = null;
  }
  epubCoverUrl.value = "";
  let book: Book | null = null;
  try {
    // The ?auth= token in the inline download URL is the real
    // authenticator (same as the working preview path, RC-44), so we
    // don't need requestCredentials. openAs:'epub' skips epub.js's
    // extension sniff, which the auth JWT's dots otherwise confuse.
    const url = api.getDownloadURL(resource, true);
    const ePub = (await import("epubjs")).default;
    if (token !== epubCoverToken) return; // selection moved on
    book = ePub(url, { openAs: "epub" });
    await book.ready;
    if (token !== epubCoverToken) return;
    const coverUrl = await book.coverUrl();
    if (token !== epubCoverToken || !coverUrl) return;
    // Copy the cover into our OWN blob URL so it survives book.destroy()
    // (which revokes epub.js's internally-created archive URLs).
    const blob = await (await fetch(coverUrl)).blob();
    if (token !== epubCoverToken) return;
    const objUrl = URL.createObjectURL(blob);
    epubCoverRevocable = objUrl;
    epubCoverUrl.value = objUrl;
  } catch {
    /* no cover / open failed — keep the generic book glyph */
  } finally {
    try {
      book?.destroy();
    } catch {
      /* ignore */
    }
  }
};

// Re-extract whenever the selected EPUB changes; clear for anything else.
watch(
  () => (isEpub.value ? (item.value?.url ?? null) : null),
  (url) => {
    if (url) void loadEpubCover(item.value as ResourceItem);
    else clearEpubCover();
  },
  { immediate: true }
);

onUnmounted(clearEpubCover);

const typeLabel = computed(() => {
  if (!item.value) return "";
  if (item.value.isDir) return "Folder";
  if (item.value.extension) {
    return `${item.value.extension.replace(".", "").toUpperCase()} file`;
  }
  return "File";
});

const sizeLabel = computed(() => (item.value ? filesize(item.value.size) : ""));

const modifiedLabel = computed(() => {
  if (!item.value) return "";
  const m = dayjs(item.value.modified);
  const now = dayjs();

  // Same calendar day → "Nh ago" (rounded to nearest hour, min 1)
  if (m.isSame(now, "day")) {
    const hours = Math.max(
      1,
      Math.round((now.valueOf() - m.valueOf()) / 3600000)
    );
    return `${hours}h ago`;
  }

  // Same year → "MMM D"   (e.g. "May 24")
  if (m.isSame(now, "year")) {
    return m.format("MMM D");
  }

  // Older → "MMM D, YYYY" (e.g. "May 24, 2025")
  return m.format("MMM D, YYYY");
});

const canShare = computed(
  () => !!authStore.user?.perm.share && !!authStore.user?.perm.download
);
const canDownload = computed(() => !!authStore.user?.perm.download);
const canRename = computed(() => !!authStore.user?.perm.rename);
const canDelete = computed(() => !!authStore.user?.perm.delete);
// Move requires rename perm (same backend operation); Copy requires create.
const canMove = computed(() => !!authStore.user?.perm.rename);
const canCopy = computed(() => !!authStore.user?.perm.create);
// Extract (PR #5746, generalized) — show in place of Preview when the
// selected file is a supported archive (zip / 7z / rar / tar family) and the
// user can create + the operator hasn't disabled the feature. Preview is
// useless for archives, so the slot is more valuable as the entry point.
const canExtract = computed(
  () =>
    unzipEnabled &&
    !!authStore.user?.perm.create &&
    !!item.value &&
    !item.value.isDir &&
    isExtractable(item.value.name)
);

const close = () => {
  // Mobile-only (desktop's header button calls toggleCollapse instead). Just
  // retract the sheet; the selection is kept so the top selection toolbar
  // stays available for actions. Deselect via the row checkboxes.
  layoutStore.mobileDetailsOpen = false;
};

const action = (name: string) => {
  layoutStore.showHover(name);
};

const open = () => {
  if (item.value) {
    window.location.href = item.value.url;
  }
};

const download = () => {
  if (item.value) {
    const url = api.getDownloadURL(item.value as ResourceItem, false);
    window.open(url);
  }
};

const onKey = (event: KeyboardEvent) => {
  if (event.key === "Escape" && item.value) {
    close();
  }
};

onMounted(() => {
  window.addEventListener("keydown", onKey);
  updateIsMobile();
  window.addEventListener("resize", updateIsMobile);
});
onUnmounted(() => {
  window.removeEventListener("keydown", onKey);
  window.removeEventListener("resize", updateIsMobile);
});
</script>

<style scoped>
/* ── RC-4: folder-summary + multi-select states + collapsed rail ──── */
.ip-summary {
  padding: 22px 16px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.ip-summary__icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}
/* Folder summary icon mirrors the GRID folder tile exactly: the accent fill +
   a soft diagonal sheen + the theme-correct on-accent folder glyph (was a
   hardcoded amber-500 base + white glyph — Calm Minimal). */
.ip-summary__icon--folder {
  background-color: var(--color-accent);
  background-image: linear-gradient(
    140deg,
    rgba(255, 255, 255, 0.22),
    rgba(255, 255, 255, 0) 60%
  );
  color: var(--color-on-accent);
}
.ip-summary__icon--accent {
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.12));
  color: var(--color-accent, #6e72d9);
}
.ip-summary__title {
  font-size: 14.5px;
  font-weight: 600;
  color: var(--color-ink-1, #18181b);
  word-break: break-word;
  line-height: 1.3;
}
.ip-summary__meta {
  font-size: 12px;
  color: var(--color-ink-3, #a1a1aa);
  margin-top: 2px;
  font-variant-numeric: tabular-nums;
}
.ip-summary__props {
  width: 100%;
  margin: 16px 0 0;
  padding-top: 14px;
  border-top: 1px solid var(--color-line, #ececec);
  display: flex;
  flex-direction: column;
  gap: 7px;
  text-align: left;
}
.ip-summary__prop {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 12px;
}
.ip-summary__prop dt {
  color: var(--color-ink-3, #a1a1aa);
}
.ip-summary__prop dd {
  color: var(--color-ink-1, #18181b);
  margin: 0;
  min-width: 0;
}
/* The Path row stacks (label above, value below) and the value wraps fully
   so the whole path is visible instead of being clipped with an ellipsis. */
.ip-summary__prop--path {
  flex-direction: column;
  gap: 3px;
}
.ip-summary__path {
  font-family: var(--font-mono, monospace);
  font-size: 11px;
  line-height: 1.4;
  word-break: break-all;
  max-width: 100%;
}
.ip-summary__hint {
  font-size: 11.5px;
  color: var(--color-ink-3, #a1a1aa);
  line-height: 1.45;
  margin: 16px 0 0;
}
.info-pane-rail {
  width: 36px;
  cursor: pointer;
}

/* Match the details panel background to the main sidebar (both --color-canvas)
   so the two rails read as the same surface. */
.info-pane {
  background: var(--color-canvas);
}

/* Full-frame color tile for files with no image preview (folders, archives,
   code, blobs, …). The vivid background-color comes from the file's
   iconColorClass (bg-{hue}); we layer the same diagonal white sheen the grid /
   gallery tiles use, and the icon inherits the class's white color. */
.info-tile {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-image: linear-gradient(
    140deg,
    rgba(255, 255, 255, 0.3),
    rgba(255, 255, 255, 0) 60%
  );
}

/* File name wraps at the <wbr> separators first, only breaking a long
   unbroken run as a last resort — no ugly mid-word splits. */
.info-name {
  overflow-wrap: anywhere;
  word-break: normal;
}

.preview-mesh {
  /* Calm Minimal: a single faint accent wash over the surface (was a
     pink/amber/indigo three-color mesh). Follows the accent picker. */
  background:
    radial-gradient(
      at 25% 18%,
      rgb(var(--accent-rgb) / 0.1) 0%,
      transparent 55%
    ),
    radial-gradient(
      at 80% 78%,
      rgb(var(--accent-rgb) / 0.06) 0%,
      transparent 55%
    ),
    var(--color-surface, #ffffff);
}

/* Dark-mode: bump alpha so the radial blobs still read against the dark
   surface, and let the surface token drive the base color. */
html.dark .preview-mesh {
  background:
    radial-gradient(
      at 25% 18%,
      rgb(var(--accent-rgb) / 0.16) 0%,
      transparent 58%
    ),
    radial-gradient(
      at 80% 78%,
      rgb(var(--accent-rgb) / 0.1) 0%,
      transparent 58%
    ),
    var(--color-surface);
}

.info-action {
  --action-hue: var(--c-blue);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 4px;
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--color-line, #ececec);
  background: var(--color-surface, #fff);
  color: var(--color-ink-2, #52525b);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition:
    background-color var(--dur-base) ease,
    border-color var(--dur-base) ease,
    color var(--dur-base) ease;
}

/* Per-action accent hue (matches the command-palette icon colors) so each
   action is recognizable at a glance. */
.info-action--share {
  --action-hue: var(--c-teal);
}
.info-action--download {
  --action-hue: var(--c-blue);
}
.info-action--rename {
  /* Muted like its siblings (was --c-amber, a pre-collapse leftover —
     gold is the FAVORITES color, v2.7.x). */
  --action-hue: var(--c-lilac);
}
.info-action--move {
  --action-hue: var(--c-blue);
}
.info-action--copy {
  --action-hue: var(--c-teal);
}
.info-action--extract {
  --action-hue: var(--c-green);
}
.info-action--preview {
  --action-hue: var(--c-lilac);
}
.info-action--tags {
  --action-hue: var(--c-lilac);
}
.info-action--danger {
  --action-hue: var(--c-rose);
}

/* Calm Minimal: resting action icons are muted ink (chrome), not per-action
   hues. Lucide svg draws with currentColor. */
.info-action :deep(svg) {
  color: var(--color-ink-2, #52525b);
}

/* Hover tints the tile with the single accent (was the per-action hue).
   color-mix degrades gracefully on engines that lack it. */
.info-action:hover {
  background: color-mix(in srgb, var(--color-accent) 12%, var(--color-surface));
  border-color: color-mix(in srgb, var(--color-accent) 40%, var(--color-line));
  color: var(--color-ink-1, #18181b);
}
/* Delete: turn the label red on hover too, for an unmistakable danger cue. */
.info-action--danger:hover {
  color: var(--c-rose);
}

/* ── Copy-path button (G6) ────────────────────────────────────────
   Small ghost button sitting next to the "Location" label. Goes
   accent-tinted in the "Copied" state to reinforce the success. */
.info-pane__copy-path {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 22px;
  padding: 0 7px;
  border: 1px solid transparent;
  border-radius: 4px;
  background: transparent;
  color: var(--color-ink-3, #a1a1aa);
  font-size: 10.5px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  cursor: pointer;
  transition:
    background-color var(--dur-base) ease,
    color var(--dur-base) ease,
    border-color var(--dur-base) ease;
}
.info-pane__copy-path:hover {
  background: var(--color-elevated, #f4f4f5);
  color: var(--color-ink-1, #18181b);
  border-color: var(--color-line, #ececec);
}
.info-pane__copy-path:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 1px;
}

/* Tag section (v1.3 S2-5) — chip list + "Manage" button. Same visual
   shape as the Copy-path affordance so the section header pattern
   stays consistent. */
.info-pane__manage-tags {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 22px;
  padding: 0 8px;
  border-radius: 6px;
  font: inherit;
  font-size: 11px;
  font-weight: 500;
  color: var(--color-ink-2, #52525b);
  background: var(--color-surface, #fff);
  border: 1px solid var(--color-line, #ececec);
  cursor: pointer;
  transition:
    background-color 0.1s ease,
    color 0.1s ease,
    border-color 0.1s ease;
}
.info-pane__manage-tags:hover {
  background: var(--color-elevated, #f4f4f5);
  color: var(--color-accent, #6e72d9);
  border-color: var(--color-line-strong, #d4d4d8);
}
.info-pane__manage-tags:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 1px;
}

.info-pane__tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.info-pane__tag-empty {
  font-size: 12px;
  color: var(--color-ink-3, #a1a1aa);
  font-style: italic;
}
/* Activate when the path is freshly in the clipboard. Subtle accent
   wash + matching ink so the flash reads as "yes, it happened". */
.info-pane__copy-path:has(svg[data-name="check"]),
.info-pane__copy-path:hover:has(svg[data-name="check"]) {
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.12));
  color: var(--color-accent, #6e72d9);
  border-color: transparent;
}
</style>
