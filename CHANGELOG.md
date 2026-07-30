# Changelog

All notable changes to **vitrine**.

## v3.1.0 — Archive extraction improvements

- **Extract several archives at once.** Select multiple archives and extract
  them in one go — each unpacks into its own subfolder under the destination
  you pick. Colliding folder names (e.g. `data.zip` + `data.7z`) are
  automatically de-duplicated (`data`, `data (2)`, …), password-protected
  archives prompt individually, and a failed one is skipped without stopping
  the rest.
- **Extract is now in the header ⋯ menu and the multi-select bar.** Previously
  you could only reach it from a row's right-click menu or the details pane.
  The ⋯ "More" menu now lists every action valid for the current selection —
  Extract for archives, Edit tags for audio, plus Share / Download / Rename /
  Copy / Move / Delete — and hides the ones that don't apply.
- **Extract options are remembered.** Your choices in the Extract panel
  (into a new folder, overwrite, delete original after extraction, open the
  extracted folder) now persist as the defaults for your next extraction,
  until you change them.
- **Fixed: inline rename could stop working until a page refresh.** After using
  split view, the active-pane target could get stuck on the second pane, which
  silently prevented the single-pane rename box from opening (or closed it
  immediately). It now resets correctly when the split isn't showing.

## v3.0.2 — Settings copy fix

- **Two settings descriptions still said "filebrowser" instead of "vitrine"** —
  on the Global Settings and User Management pages. Now corrected. Cosmetic
  only; no behavior change.

## v3.0.1 — Login wordmark fix

- **The login page, browser tab, and PWA name now always show the "vitrine"
  wordmark.** On instances carried over from before the rename, these rendered
  the stored instance branding name — which still read "filebrowser pretty" and
  appeared without its accent coloring. The wordmark is now fixed and no longer
  reads the stored branding name, so it stays correct and consistently colored
  everywhere, matching the sidebar. (The sidebar already used the fixed
  wordmark, which is why only the login screen and tab title were affected.)

## v3.0.0 — Renamed to vitrine

**The project is now called vitrine.** Same app, new name — a cleaner identity
for the polished File Browser fork formerly known as "filebrowser pretty." This
is a name-only release: no features were added or removed, and your data,
settings, users, shares, and tags are all preserved.

- **New name and wordmark.** The app is "vitrine" everywhere — the login page,
  sidebar, browser tab, and PWA manifest. The per-letter rainbow accent now
  colors the whole wordmark.
- **New home and image.** The project lives at
  `github.com/csummers-dev/vitrine`, and the container image is now
  `ghcr.io/csummers-dev/vitrine`. The old `filebrowser-pretty` image stops at
  2.8.2 and won't receive further updates.
- **Environment variables are now prefixed `VITRINE_`** (was `FB_`) — for
  example, `FB_REDIS_CACHE_URL` becomes `VITRINE_REDIS_CACHE_URL`. `PUID`,
  `PGID`, and `TZ` are unchanged; they were never `FB_` variables.
- **Default on-disk names** follow the new brand: the database defaults to
  `vitrine.db` (audit log `vitrine-audit.db`), and the auto-discovered config
  file is `.vitrine.{json,toml,yaml,yml}` in `/etc/vitrine/`, `$HOME`, or the
  working directory.

> **Upgrading:** point your Compose file at
> `ghcr.io/csummers-dev/vitrine:latest` and pull. Your existing
> `/config/settings.json` keeps its `database` path, so the app keeps reading
> your current `filebrowser.db` with no data loss — renaming the file to
> `vitrine.db` is optional cleanup (stop the container, rename the file in your
> `/database` volume, then update the `database` value in `settings.json` to
> match). If you set any `FB_*` environment variables, rename them to
> `VITRINE_*`. `PUID` / `PGID` / `TZ` / `group_add` are unaffected.

## v2.8.2 — Security patch

- **Moved to the Go 1.25.12 toolchain**, closing an Encrypted Client Hello
  privacy leak in the standard library's `crypto/tls` (GO-2026-5856). No
  functional or visual changes.

## v2.8.1 — Rename selection fix

- **Fixed: renaming lost your text selection on release.** While renaming, if
  you dragged to select part of the name and let go with the cursor off the
  row, the selection (and the edit) were dropped. You can now click, drag to
  select, and release anywhere without losing your selection.

## v2.8.0 — Profile avatars, sortable move/copy, mobile polish

- **Upload your own profile photo.** Settings → Profile has a new Photo
  section: pick an image, reposition and zoom it under a circular crop, and
  save. Your avatar then shows everywhere your initials used to — the sidebar,
  the mobile drawer, and the settings rail. Removing it asks for confirmation,
  and reverts to your initials. The photo is cropped to a small square in your
  browser and stored on your account (it syncs like your theme and accent, and
  needs no server configuration).
- **Sort the destination list when moving or copying.** The folder picker in
  the Move / Copy panels was locked to one order — it now has a Sort control
  (Name or Modified, either direction), and remembers your choice.
- **Mobile drawer polish.** The dividers in the slide-out menu now match the
  desktop sidebar: no line under the logo, and the divider above your account
  is inset from the edges instead of running full-width.
- **Trash "Empty trash" button moved.** It now sits next to the title instead
  of the far-right edge, where a notification toast could cover it and block
  the click.

## v2.7.4 — A clearer trash error

- **"403 Forbidden" in the Trash now explains itself.** When the server can't
  delete a trashed item because something inside is owned by another system
  user (e.g. artwork or trickplay folders written by a different container),
  the Trash page now says exactly that — and that fixing the volume's
  ownership is the way out — instead of a bare status code. (The 2.7.3
  permission-lifting fix covers read-only modes the server's user owns;
  files owned by a *different* user are beyond what any unprivileged
  process can chmod.)

## v2.7.3 — Trash fix & housekeeping

- **Fixed: "Delete forever" failing with 403 Forbidden.** Emptying a trashed
  folder that contained a read-only subdirectory (or one owned with
  restrictive permissions by another service — easy to hit on a NAS where
  several containers write the same volume) failed, because deleting needs
  write access on every directory inside. Trash deletion now lifts the
  directory permissions inside the trash and pushes through — you already
  chose to delete it.
- Two dead-code removals that slipped past local checks and failed CI lint on
  the 2.7.2 push (an unused loop index in the sidebar's recents list and an
  unused store handle in the mobile drawer).
- Updated `golang.org/x/net` from 0.54.0 to 0.55.0 (indirect dependency,
  Dependabot).

## v2.7.2 — Branding, sidebar & header refinements

A refinement patch: your logo and favicon become theme-aware, the sidebar slims
down to pure navigation, and the header sheds its button crowd.

- **Theme-aware branding.** The logo in the sidebar, mobile drawer, and login
  card now switches between its light and dark variants with your theme, the
  favicon follows your system's light/dark appearance live (the browser tab
  it sits on is drawn by the OS theme, not the app's), and the installed-app
  icons use the new light set.
- **Truer favorites gold.** Favorite stars share one color everywhere and it's
  a luminous gold again (the dusty pass had muddied it to brown in light
  mode). Gold now means favorites only — the Rename action button and the
  conflict-dialog icon that borrowed it use their proper muted and warning
  tones.
- **A quieter sidebar.** The New folder / new file buttons left the rail (and
  the mobile drawer) — creation lives in the listing: the ⋯ menu, right-click
  on empty space, and the split-view header. A detail pass rides along: the
  active nav item's icon now tints with its label, the Favorites/Recent
  section labels match the app's eyebrow size, and the account row's divider
  is inset instead of running edge to edge.
- **A calmer header.** The toolbar's crowd of buttons thins to four: View,
  Sort, Upload, and ⋯. The new View button wears the current layout's icon
  and opens one popover holding List / Grid / Gallery and Split view —
  replacing the three-button switcher and the separate split toggle.
- **A real mobile pass.** Gallery view is a proper two-up grid on phones
  (tiles used to render one enormous column), search is a full-width pill
  instead of a stranded icon on its own row, Quick Look fits the screen
  (its panel used to overflow), and the drawer gained the Trash entry it
  never had — previously there was no way to reach the recycle bin on
  mobile at all.

## v2.7.1 — The panel shell

A visual patch on top of 2.7.0: the layout now actually shows the depth the new
palette was built for. (2.7.0's canvas and panel colors were right, but the
files view painted everything with the page background, so the whole app read
as one flat color.)

- **Content floats as panels.** The file listing, the second pane in split
  view, and the details rail are now rounded surface cards with a hairline
  border and a soft shadow, sitting on the deeper canvas with a small gutter —
  the layout the Dusty Minimal palette was designed around. Split view reads as
  two clean cards side by side.
- **The sidebar went chromeless.** Its right border is gone; the gutter alone
  separates it from the content card, and its accent button and storage card
  stand out as the rail's only chrome.
- **Trash matches.** The Trash page sits in the same floating panel as the
  files view instead of stretching edge to edge.
- **Ambient background still works.** With the mesh and translucent surfaces
  enabled, the panels thin slightly so the glow shows through — without
  flattening the canvas-and-panel depth again. On phones the gutters collapse
  and the page itself becomes the panel, as before.

## v2.7.0 — Dusty Minimal & a quality-of-life batch

A follow-up pass on the Calm Minimal redesign — the whole palette relaxes into
softer, evened-out tones — plus a batch of everyday quality-of-life features:
Quick Look previews, undo for moves and renames, faster ways to get around, and
comic covers. Your files, settings, and shares work exactly as before.

**A softer palette — "Dusty Minimal"**

- **The canvas and panels finally separate.** The page background drops a real
  step below the white panels (with hairlines you can actually see), so surfaces
  read as surfaces instead of one flat sheet.
- **Every neutral picks up a whisper of the accent's hue**, so the gray scale and
  the accent read as one family instead of clashing.
- **File-type colors stop shouting.** The tint palette (icons, tags, preview
  chips) has its saturation roughly halved and its lightness evened out — color
  still codes file type, it just no longer reads as confetti in a mixed folder.
- **New default accent: Iris** — the same violet family, pulled back from the
  saturation ceiling. It joins the picker as a seventh preset; if you've already
  chosen an accent, your choice is kept.
- **Matching touches everywhere:** the wordmark letters, favorite-star gold,
  danger rose, and the undo toast all soften to the same palette, and the big
  folder cards in grid/gallery views trade their solid accent fill for a calm
  accent wash (the small list-row folder icon keeps the solid marker).
- **Status colors joined the family.** Every success, error, and warning
  surface — toasts, dialogs, error messages, form validation, banners — now
  draws from one dusty status palette instead of scattered loud hexes, in both
  themes. This also caught a few spots that had quietly kept the very old
  pre-redesign palette (the boot spinner, legacy buttons, the right-click
  Delete hover), which now follow your accent and the new tones.

**Quick Look**

- **Press Space to peek at the selected file** — images, video, audio, PDFs, and
  text preview instantly in a frosted overlay without leaving the folder. Arrow
  keys move through neighbors while the peek follows, Enter opens the file for
  real, and Space or Esc dismisses. Works in both split-view panes.

**Undo**

- **Moves can be undone.** After a move completes, a toast offers Undo — clicking
  it moves everything straight back (even items that were renamed to avoid a
  conflict return to their original names).
- **Renames can be undone** from the same kind of toast.
- Deletes already had Undo via the Trash; now all three destructive actions are
  reversible.

**Getting around**

- **Breadcrumb folder menus.** The chevrons between breadcrumbs are now buttons —
  click one to list that level's folders and jump sideways (Comics → Music) in
  one click.
- **Type filter chips.** Folders that mix file kinds get a quiet chip row
  (Images · Videos · Audio · Comics · Documents · Archives) that filters the
  listing client-side. Folders always stay visible; the filter resets when you
  navigate.
- **Paste to upload.** ⌘V with files (or a screenshot) on the clipboard uploads
  into the current folder — screenshots get tidy timestamped names instead of
  colliding as "image.png".
- **Share & download on hover.** Grid and gallery file tiles show quick Share and
  Download buttons on hover, next to the favorite star.

**Comic covers**

- **CBZ and CBR comics now show their cover** (the first page) as their thumbnail
  in the listing, gallery, and details pane — served by the same cached pipeline
  as the other media covers, and shared with the comic reader so nothing is
  extracted twice.

**Fixes & polish**

- Fixed the gallery tile caption, which rendered as "· 1h ago0 B" — it now reads
  "0 B · Jul 1" like the grid.
- List-view checkboxes appear on hover/selection instead of sitting on every row
  (always visible on touch, where they're the multi-select entry point).
- The Quick Look backdrop is frosted glass, matching the app's other floating
  controls.
- Keyboard focus now shows a consistent accent ring on the new controls — and
  hover-revealed buttons become visible when focused, so tabbing never lands on
  an invisible control.

**Under the hood**

- The frontend suite grew from 383 to 412 tests (45 files): move-undo path math,
  the Quick Look state and overlay, breadcrumb sibling menus, type-filter
  buckets, and pasted-file naming, plus a new Go test locking in which archive
  extensions get comic covers.

## v2.6.2 — Security patch

A dependency and toolchain update that clears every known security advisory. No
functional or visual changes — your files, settings, and shares behave exactly as in
2.6.1.

- **Updated `golang.org/x/image` to v0.43.0**, fixing two TIFF image-decoding flaws (a
  decoder panic on a malformed strip offset, and an unbounded tile allocation) that
  were reachable when the server decodes uploaded images for thumbnails and previews.
- **Moved the build to the Go 1.25.11 toolchain**, picking up the current batch of Go
  standard-library security patches.
- Verified clean with `govulncheck` — zero advisories affecting the code.

## v2.6.1 — Details pane & rename polish

Two small follow-ups to the Calm Minimal redesign, plus a big expansion of the
automated test suite. Nothing about your files, settings, or shares changes.

- **The details pane stays where you put it.** Selecting a file or folder no longer
  forces the details pane open — it now follows only the collapse/expand toggle, so
  clicking around the listing never interrupts what you were doing. Expand it from the
  rail whenever you want it, and it remembers your choice.
- **Renaming keeps your text selection.** When you drag to select part of a name and
  let go of the mouse away from the row, the highlighted text stays selected instead
  of clearing — selecting a long stretch to overwrite is no longer fiddly.

**Under the hood**

- Substantially expanded automated tests — backend auth/IDOR, password, and command-
  parser units, plus frontend store/composable coverage and the first component-mount
  tests (383 frontend tests + new Go suites, all green). No runtime changes.

## v2.6.0 — Calm Minimal redesign

A top-to-bottom visual redesign — a calmer, more focused look — plus a round of
layout and drag-and-drop refinements. Your files, settings, and shares all work
exactly as before; everything just looks (and reads) cleaner.

**A calmer look — "Calm Minimal"**

- **One accent, everything else neutral.** The old multi-color chrome is gone. You
  pick a single accent color (six hues — violet by default) in Settings → Profile,
  and it's used only where it carries meaning: the active folder, primary buttons,
  the breadcrumb home, and the current selection. Every other icon is a quiet,
  uniform ink color.
- **Refined light & dark themes** on a deeper, flatter canvas, with the ambient
  background reduced to a single faint accent wash.
- **Muted file-type icons** — file types stay recognizable at a glance, but the
  colors are desaturated tint-chips instead of loud fills; folders are the one
  solid-accent tile.
- **Accessible contrast.** Text and chrome colors were tuned to meet WCAG AA.

**Header & navigation**

- **The breadcrumb leads the header.** The path now sits at the top as the primary
  location nav (the current folder is the last crumb), replacing the large folder
  title + "FOLDER" eyebrow and the separate bottom path bar. The header is
  noticeably shorter, with a compact item count beside the path.

**Listing & controls**

- **One Sort control.** The two sort buttons (field + direction) merged into a
  single Sort button whose popover holds the field, the Ascending/Descending
  direction, and the secondary "then by" sort. (List-view column headers still flip
  direction with a click.)
- **The details pane starts collapsed** and opens automatically when you select a
  file — so the listing gets the full width until you actually need details.
- **Uniform gallery cards.** Folder and file tiles in gallery view are now the same
  size, so rows line up cleanly instead of folders rendering smaller than files.

**Drag & drop**

- **Bigger breadcrumb drop targets.** While you drag a file, each breadcrumb folder
  grows into a larger, clearly-highlighted drop zone (filling the row height and
  bridging the gaps) so it's obvious where it will land.
- **Escape reliably cancels a drag.** Pressing Esc mid-drag now always clears the
  floating Copy/Move badge and the drag state — previously the badge could stay
  stranded on screen.

## v2.5.2 — Security hardening & code-quality pass

The result of a top-to-bottom security and code-quality audit of the whole app,
with the findings fixed. No new features, and nothing changes for day-to-day use.

**Security hardening**

- **Login brute-force protection.** Repeated failed logins from the same IP now
  trigger a short lockout (with a `Retry-After`), on top of the existing bcrypt
  password hashing.
- **Webhook SSRF guard.** Outbound webhook deliveries can no longer be aimed at
  loopback, private, link-local or cloud-metadata addresses — they're resolved
  and blocked at connect time, and re-checked across redirects (defeating
  DNS-rebinding).
- **Stronger share links.** Share hashes are now 128-bit, and share-password and
  token comparisons run in constant time.
- **Tighter sandboxing & confinement.** The EPUB reader's iframe explicitly
  disables scripting, the cache directory is created with private permissions,
  and custom-branding file overrides are confined to the branding folder.

**Tooling & dependencies**

- **CI security scanning** — `govulncheck` + `gosec` (Go) and `pnpm audit`
  (frontend) now run on every build.
- **Dependency cleanup** — removed unused frontend packages, patched
  build-time-only advisories (esbuild, rollup), updated DOMPurify, and tidied the
  Go module graph.

**Correctness, performance & tests**

- Fixed an O(n²) cache-eviction sort and hardened the internal events bus's
  subscribe / unsubscribe.
- The resumable (TUS) upload path now reports total bytes for progress, matching
  the direct-upload path.
- Added security regression tests (login lockout, path confinement, webhook SSRF)
  and tightened type-safety across the API layer.

## v2.5.1 — Faster cold start + simpler Docker permissions

**Smaller, faster first load.** Bundle optimizations — same app, just much less
to download and parse on first load:

- **Icons are tree-shaken.** The app bundled the *entire* icon library on every
  load; it now ships only the icons it actually uses — the icon chunk drops from
  ~195 kB to ~53 kB gzipped.
- **Settings, admin, login, share and trash pages load on demand** instead of in
  the initial bundle (route-level code splitting), trimming the boot chunk.
- **Dropped the legacy-browser build** — a full duplicate bundle set plus a
  loader, for pre-2018 browsers that this self-hosted app's users never run.

Net: roughly **~170 kB less gzipped JavaScript** fetched and parsed on a cold start.

**Simpler, safer Docker permissions.** The image now uses the standard
`PUID` / `PGID` pattern instead of a hardcoded `user:` line. Set them to the user
that owns your data (`id -u` / `id -g` on the host) and the container fixes the
ownership of its own `/config` + `/database` to match, then drops to that
unprivileged user to run the app — it never runs the app as root, and you never
have to `chown` your media or guess a `user:` value. Your files under `/srv`
aren't touched.

> **Upgrading:** remove any `user: "1000:1000"` from your `compose.yaml` and add
> `PUID` / `PGID` under `environment` (both default to `1000`). Mount media that's
> already owned by that uid (user-owned subdirectories, not root-owned NAS volume
> roots). Prefer strictly non-root? Set `user:` instead and the container skips
> the ownership step. Full details in the README's **Permissions** section.

## v2.5.0 — Dual-pane (split) file browsing

A second directory listing, side by side, so you can compare two folders at a
glance and move files between them without losing your place. Off by default.

**Two panes, side by side**

- A **Split** button in the listing toolbar (`columns-2`), or the **Split view**
  palette command, opens a second pane next to the first. Each pane navigates
  independently — folders, breadcrumb, parent button, and its own sort. The
  details rail steps aside while split; both panes are list view.
- **Per-user persistence.** Whether the split is open and which folder the second
  pane shows are saved to your prefs and restored on your next visit (a missing
  saved folder offers a one-click **Go Home**). The URL still encodes only the
  primary pane.
- **Active pane.** Exactly one pane is active (a soft accent edge marks it).
  Clicking a pane activates it, and the sidebar (Home / favorites / recents) and
  command-palette navigation target the active pane. **F6** switches panes (also
  a palette command).

**Both panes are full equals**

- The second pane has its **own toolbar** (parent · sort · new folder · upload ·
  ⋯ overflow · close) and the **full right-click menu** — Open, Copy path,
  Rename, Tag, Download, Share, Extract, Move/Copy, Delete, New folder, New file
  — all scoped to that pane and permission-gated.
- **Inline** new folder / file and rename (no dialog), a **selection pill**
  (move/copy · tag · download · delete · clear), **upload** straight into its
  folder with the usual conflict resolution, and a full **move/copy destination
  picker** to any folder.

**Moving files between folders**

- **Cross-pane drag** to move — drop on a folder row, or anywhere in the other
  pane (including an empty folder); hold the copy modifier to copy. Runs through
  the same background-transfer + conflict-resolution pipeline as every other
  move/copy; both panes refresh when it settles, and edits to a shared folder
  cross-refresh the other pane.

**Keyboard follows the active pane**

- Arrows · Home/End · Enter · ⌘A · Delete · Esc · **type-ahead** drive whichever
  pane is active; `/` refreshes it. The primary pane's keyboard is unchanged when
  it's active.

**Fast, and responsive to each pane**

- Folder sizes load **lazily per visible row** instead of pre-walking both panes
  up front, and the panes no longer wipe each other's tag chips on navigation.
- Each pane's list responds to **its own width** (not the window's): a narrow
  pane sheds the Size then Modified columns to keep the file name and icon
  visible — which also fixes narrow single-pane windows.
- Split needs room: below ~880px (and on mobile) it collapses to a single pane
  and the toggle hides.

**Listing header refresh** (single pane)

- Search moved to **its own row** beneath the toolbar; **Sort and Upload are
  icon-only** and the toolbar no longer wraps onto a second line. The breadcrumb
  shows the **full path and scrolls** when it's longer than the bar (long
  ancestor names ellipsize; the current folder stays in view).

## v2.4.0 — Cut/copy/paste, a recycle bin, and transfers that don't lose data

The largest release since the "pretty" milestone, landing the whole 2.4.0 arc:
a Finder-style cut/copy/paste, a real recycle bin, transfers that survive a
server restart and can verify what they wrote, recursive folder sizes, instant
indexed search, and bulk tagging — plus a long tail of UI polish and fixes.

**Moving & organizing files**

- **Cut / Copy / Paste, Finder-style** (⌘X / ⌘C / ⌘V, plus right-click *Paste into
  folder*) running through the same background transfer pipeline as drag-drop —
  same-volume paste is instant, conflicts get the full resolve dialog, and the
  conflict dialog now shows the exact "kept as …(1)" name up front.
- **Trash / recycle bin.** Deleting moves to a per-drive `.trash` (an instant
  rename, even for huge folders); a **Trash** page lets you Restore or Delete
  forever, Undo works even after navigating away, **Shift+Delete** skips the bin,
  and an admin **retention (days)** setting auto-purges. Tags follow a file in and
  back out; `.trash` never shows in listings or search.

**Transfers that don't lose data**

- **Survive a restart + Retry.** A big move/copy interrupted by a server restart
  reappears in the dock marked *Interrupted* with a **Retry** that re-runs only
  the unfinished items. Transient filesystem hiccups auto-retry; real errors fail
  fast.
- **Live speed + ETA** in the transfer dock; instant same-drive moves got faster
  (no up-front byte-count walk).
- **Opt-in verify-after-copy** (admin setting): re-reads and checksums every copy
  against its source — a mismatch fails the transfer and **keeps the original**,
  including before a cross-volume move deletes its source.

**Finding & understanding files**

- **Folder sizes** — a folder's recursive total now shows in the details pane (and
  the Size column once computed), measured on demand and kept fresh as contents
  change; the hidden Trash isn't counted.
- **Instant search** — the server keeps an in-memory index per user instead of
  walking the whole tree on every keystroke; identical results, far faster, self-
  refreshing, with a *Rebuild search index* command if ever needed.
- **Bulk tagging** — tag a whole multi-selection at once with a tri-state picker
  (on all ✓ / on some – / none).

**Polish & fixes**

- **Sorting is instant.** Choosing a sort field or flipping the direction now
  re-orders the listing immediately instead of only after the next navigation,
  and **Then by → None** truly clears the secondary sort.
- **Correct file-type icons everywhere.** Images, audio, video and PDFs now show
  their proper coloured glyph even where only the filename is known — the Trash
  page, search hits, recents — instead of a generic grey tile.
- **Trash page redesign.** The recycle bin now matches the rest of the app: a
  proper header, flat full-width rows with coloured icons, and Restore / Delete
  on hover. The sidebar's Trash icon is coloured to match.
- **Folder sizes fill in on their own.** A listing's Size column now populates
  every folder at once (bounded + cache-backed) instead of one click at a time.
- **"Last updated" never reads as the future.** A folder time slightly ahead of
  your clock no longer shows "in a few seconds" — it's always framed as past.
- **No false upload warning.** Browsing between folders or pages while an upload
  is running no longer claims it will be cancelled — it keeps going. (Closing the
  tab still warns, because that genuinely stops it.)
- The "moved to Trash" undo window is **5 seconds**, and the directory header's
  "N items · last updated …" line no longer wraps onto a second line.

## v2.4.0-alpha.6 — Indexed search + bulk tagging

- **Search is instant on big libraries.** Searching used to walk the whole folder tree on the server for *every keystroke* — fine for a handful of files, sluggish on a large collection. The server now keeps an in-memory **name + path index** per user that answers searches without touching the disk. It builds itself in the background the first time you search (falling back to the old live walk until it's ready) and keeps itself up to date automatically as you add, rename, move, or delete files. Results are identical to before — the speed-up is invisible except that it's fast. A **"Rebuild search index"** command (in the ⌘K palette) is there if results ever look stale.
- **Tag a whole selection at once.** Select multiple files/folders and a new **Tags** button appears in the selection bar. It opens a picker showing each tag as **on (✓)**, **off**, or **mixed (–)** — *mixed* meaning the tag is on some of the selected items but not all. Check a tag to add it to everything, uncheck to remove it from everything, or leave a mixed tag alone to keep each item as-is. One Save applies the whole change in a single request. (Previously you could only tag one file at a time.)

_Alpha builds are development checkpoints, not releases._

## v2.4.0-alpha.5 — Verify-after-copy + folder sizes

- **See how big a folder really is.** Select a folder and its **total size** (everything inside, recursively) now shows in the details pane — and once computed, in the listing's Size column too (folders used to just show "—"). It's measured on demand and cached, so it's quick and doesn't re-scan every time; the hidden Trash is never counted. The figure stays fresh automatically as you add, remove, or edit files anywhere inside.
- **Optional integrity check on every copy.** A new **admin setting — Global settings → "Verify copies"** — re-reads each freshly-copied file and checksums it (xxhash64) against the original. If they don't match, the transfer **fails and keeps your original** instead of trusting a bad copy. It also covers the copy phase of a cross-volume move, verifying *before* the source is removed — so a corrupted move can never delete the only good copy. Off by default (it roughly doubles copy time); turn it on when integrity matters more than speed.
- **Edits and extractions keep folder sizes honest.** Saving a file in the editor, or extracting an archive, now updates the affected folders' sizes immediately (previously a content edit could leave a stale total). These also flow through to the audit log and webhooks as a new **"file modified"** event.

_Alpha builds are development checkpoints, not releases._

## v2.4.0-alpha.4 — Transfer hardening

- **Big move/copy jobs survive a server restart.** Background transfers now record their progress to disk as they run, so if the server is restarted (or crashes) mid-transfer, the dock shows the job as **Interrupted** when it comes back — with a **Retry** button — instead of the job silently vanishing. Retry re-runs only the items that hadn't finished yet (a half-done batch picks up where it left off, not from the start) and drops the old entry. The same Retry button appears on any **failed** or **canceled** transfer.
- **A flaky transfer retries itself before giving up.** A transient filesystem hiccup mid-copy (an interrupted syscall, a momentarily-busy file, a timeout) is now retried automatically a couple of times with a short backoff, instead of failing the whole job on the first blip. Genuine errors (missing source, permission denied, out of space, cross-device) still fail fast — they're not the kind of thing retrying fixes.
- **Live speed and time-remaining in the transfer dock.** A running transfer now shows its current **throughput** (e.g. "12.4 MB/s") and an **estimated time remaining** (e.g. "3m left"), computed over a rolling window so the number tracks the *current* speed and recovers after a stall rather than being dragged down by a slow start.
- **Instant same-drive moves got even faster.** A same-volume move (the "fast lane" — a rename that copies no bytes) no longer walks the whole tree up front to total its bytes, since it isn't going to copy any. Moving a folder with thousands of files now starts immediately instead of pausing to count first.

_Alpha builds are development checkpoints, not releases._

## v2.4.0-alpha.3 — Trash / recycle bin

- **Deleting now moves to the Trash instead of being permanent.** Delete a file or folder (the Delete/⌫ key, the menu, the pill bar, anywhere) and it slides into a recycle bin you can get it back from. The move is instant — the item is renamed into a hidden `.trash` folder on the same drive, so even deleting a huge folder is immediate and copies nothing. A **new Trash page** (in the sidebar) lists everything you've deleted with where it came from and how long ago, and lets you **Restore** it to its original spot or **Delete forever**. There's an **Empty trash** button too.
- **Undo actually restores now.** The "Moved to Trash" toast's **Undo** puts the item right back — and because the delete already happened server-side, undo works even after you've navigated to another folder (the old optimistic-delete undo couldn't).
- **Skip the Trash when you mean it.** **Shift+Delete** (and the Trash page's own "Delete forever") removes an item permanently, with a clearly-worded confirmation and no recycle step.
- **Tags, search, and storage all stay sane.** A tag on a file follows it into the Trash and back out when you restore it. The `.trash` folder never shows up in listings or search results. Admins get a **Trash retention (days)** setting (Global settings; 0 = keep until you empty it) that auto-purges old items.

## v2.4.0-alpha.2 — Cut, Copy & Paste

- **Cut / Copy / Paste, Finder-style.** Select files or folders and press **⌘X** (cut) or **⌘C** (copy), then **⌘V** to paste — into the current folder, or use the new right-click **Paste into folder** on any folder to paste without navigating into it first. Cut rows render dimmed until you paste (or press **Esc** to call it off); a copy stays on the clipboard after pasting so you can paste it again elsewhere. Paste runs through the same background transfer pipeline as drag-drop and the move/copy tool, so a same-volume cut→paste completes instantly on the fast lane, conflicts get the full resolve dialog, and progress shows in the floating dock. The right-click menu gained **Cut** and **Copy** (the destination-picker entries were relabeled "Move/Copy … to…" so the two kinds of copy can't be confused), and the existing "Copy path" action picked up a keyboard shortcut: **⌘⇧C**.
- **Sane same-folder behavior (bug fix):** pasting a CUT back into the folder it came from used to *rename every item onto a "(1)" suffix of itself* — it's now a no-op that simply disarms the cut. Pasting a COPY in place duplicates the items with the usual "(N)" suffix directly, with no conflict popup — the old popup's "Override" choice there was a destructive self-copy trap.
- **"Keep both" now tells you the resulting name.** In the conflict dialog's per-item view, a row resolved to keep-both shows exactly what the incoming copy will be named (e.g. *will be kept as "report(1).pdf"*) — computed with the same naming scheme the server uses, including its quirky edge cases (".bashrc" → "(1).bashrc"). Works across every conflict surface: paste, the move/copy tool, and drag-drop. (Best-effort under concurrent writes; the server's final name wins.)
- **Fixed a data-loss edge:** copying an item onto itself with "Override" (reachable via the move/copy tool pointed at the item's own folder) would truncate the file while it was being read — the copy destroyed its own source. The server now rejects any move or copy of an item onto itself, while keep-both self-copies (which rename apart) and case-only renames still work.

## v2.4.0-alpha.1 — Drop-handler consolidation (internal)

- **All drop decisions now share one resolver.** Whether you drag a file onto a folder with the mouse, with touch, or drag a file in from your computer to upload, the "drop INTO this folder vs. drop alongside it" choice is now computed in exactly one place (`resolveRowDropMode`) — and recomputed the moment you release, never from a stale cached flag. This is the structural fix behind the 2.3.1 drop regression: the four drop paths can no longer drift out of sync, so the highlight you see and where the file actually lands always agree. No visible behavior change; it hardens the foundation for the rest of 2.4.0 (cut/copy/paste, recycle bin, transfer hardening). A new parity test asserts every path resolves identically.
- **Internal cleanup:** removed a legacy `__vue__` shim — a global hook that stamped a back-reference onto every component's element just so the drop code could read a folder's path from the DOM. The drop code now uses real component data instead, which also means slightly less work per rendered row.

_Alpha builds are development checkpoints, not releases._

## v2.3.1 — Fix dropping onto a folder row

- **Dropping a file onto a folder row is scoped to the icon + name again** — the "drop INTO this folder" target is meant to be just the folder's icon and name text (exactly where the highlight + spring-load countdown appear); the rest of the row drops "alongside" (into the current folder). Two problems combined to break this: (1) the hit-test measured the name off the `.item__name-text` element, which is `flex: 1` and spans the whole name column — so the rendered name now lives in an inner inline span whose box is exactly the glyph run, and the hit-test reads that (`Element.getBoundingClientRect()` stays correct mid-drag, unlike the text/Range measurement it replaced); and (2) more importantly, the **drop itself reused a cached "in zone" flag that could be left stale** (e.g. by spring-load navigation), so a release in the alongside area — with no highlight showing — still dropped INTO the folder. The drop now recomputes the same hit-test at the release point, so it can never disagree with the highlight. (This was newly exposed when 2.3.0 removed the bottom "drop into this folder" panel that had masked it.) All three drop paths now share that one hit-test: internal desktop drag, internal touch drag, and **dragging a file in from your computer to upload** — which previously treated the entire folder row as "upload into this folder" (it never checked the zone at all)
- **Clearer move/copy name-conflict message** — moving or copying an item into a folder that already has something with the same name now reports *“can't move "name": an item with that name already exists in the destination folder”* instead of a bare *“409 Conflict”*. (Conflict responses now carry their specific message through to the UI, the same way bad-request errors already did.)

## v2.3.0 — Quick same-volume moves jump the queue

- **A same-volume move no longer waits behind a long cross-volume copy** — background transfers used to run strictly one at a time, so if you kicked off a big move between drives (which copies every byte and can take a while) and then moved a file or folder *within* the same drive, that quick one had to wait for the big one to finish. Same-volume moves are instant `rename` operations, so they now run on a separate fast lane and complete immediately, even while a long copy is still going. Multi-file selections qualify too, as long as every item stays on its volume. Cross-volume moves and copies are unchanged — they still run one at a time on the main lane, because running large copies in parallel just thrashes the same disks. This is fully automatic; there's nothing new to click
- **Fixed a rare stuck state after many drag-moves** — dragging a file into a folder via spring-load (hovering a folder until it opens mid-drag) could unmount the dragged row before its "drag ended" cleanup ran, leaving an internal drag-snapshot behind. That stale snapshot made the app think a drag was still in progress, which quietly froze background listing refreshes and could make a rename appear to abort itself — previously only a full page reload cleared it. A document-level fallback now always clears the snapshot when a drag ends, so it can't leak
- **Removed the large "Drop to move into [folder]" panel** at the bottom of the listing — now that the row drop targets are tightened so you can't accidentally drop into the wrong folder, the big bottom drop zone was redundant. Dropping a dragged item on empty space or alongside a file still moves it into the current folder, exactly as before
- **Comic previews now open in "dynamic" fit by default** — instead of always fitting to width, a comic opens fitting each page to its own orientation (wide pages fit the width, tall pages fit the height), so it's sized sensibly from the first page without toggling. The width/height/original/dynamic toggle is unchanged

## v2.2.2 — Rename no longer jumps your scroll position

- **Renaming a file keeps your place in the list** — after the in-place refresh landed in 2.2.0, renaming an item in a long folder snapped the listing so the renamed (and now re-selected) row jumped to the top, losing your scroll position — disruptive in big folders. A same-folder refresh (a rename, and likewise a move/copy completing, an upload finishing, or a tag edit) now leaves your scroll exactly where it was; the listing updates underneath you without moving. Returning to a folder you'd previously scrolled still restores your position as before — only genuine folder changes touch the scroll now

## v2.2.1 — Multi-architecture Docker image

- **The Docker image is now published for both `linux/amd64` and `linux/arm64`** — only an amd64 image was being pushed, so `docker compose pull` failed with `no matching manifest for linux/arm64/v8` on Apple Silicon and other arm64 hosts (it ran only on amd64 machines like an x86 NAS). The image is now a multi-arch manifest, so it runs **natively** on arm64 and amd64 alike. The build is also fully self-contained: the Dockerfile builds the frontend and cross-compiles the Go binary for the target architecture, so the published image no longer depends on a pre-built host binary

## v2.2.0 — Transfer UX: live progress, in-place refresh & shimmer

- **The transfer progress dock moved to the bottom-RIGHT** — it used to sit bottom-left, where it covered the sidebar's favorites, storage meter, and account row. It's now anchored bottom-right, and when an upload is also in progress it floats just above the upload dock (measured to its exact height) so the two never overlap
- **The transfer progress dock now appears the moment a move/copy starts** — for a bulk drag-and-drop move or a cross-volume copy, the floating progress dock used to show up only when the transfer *finished*, instead of when it began. The dock now shows a placeholder row **synchronously, the instant you act** — before the request that enqueues the job has even returned — so it can never wait on a server round-trip or the once-a-second poll. The placeholder is swapped for the real job as soon as the server acknowledges it, and the first live progress is pulled promptly after. The reveal gate is also re-checked on a fast client-side ticker timed by the browser's own clock, so it's immune to server/browser clock skew — while an instant same-volume rename still doesn't flash
- **Smarter auto-select after a copy/move, so it never interrupts you** — when a transfer completes it selects the files it produced — for any move or copy, same- or cross-volume — but only when both of these hold: (1) you're **viewing the destination folder** the files landed in (if you've navigated elsewhere, none of the produced paths match the listing, so your current selection is left untouched), and (2) you **haven't selected different file(s) while the transfer was running** — so a long copy you kicked off never yanks your selection away from whatever you've moved on to. This replaces the earlier device-ID volume check, which has been removed in favor of these two simpler, more predictable guards
- **The listing now keeps itself current as a bulk move/copy progresses** — when you move or copy many (or large) files/folders at once, the affected folders used to update only when the *entire* batch finished, so you couldn't tell what had and hadn't completed without a manual refresh. The listing now refreshes itself each time the transfer finishes another file: completed items drop out of the **source** folder as they leave, and the new files appear in the **destination** folder as they land — whichever you're viewing (including a folder you've browsed into at either end). It only refreshes when the transfer actually touches the folder on screen (an unrelated folder isn't touched), it's throttled so a folder of many small files can't thrash, and your scroll position is kept so the list updates in place
- **Background refreshes never interrupt what you're doing** — a refresh now waits if you're mid-action: an inline rename (a file, or the current folder), creating a new folder/file, an open move/copy/share/conflict panel, or an active drag. The refresh is remembered and runs the instant you finish, so a long transfer ticking away in the background can never yank focus out of a rename or close a panel you're using
- **Background refreshes no longer flash the listing** — any in-place refresh (the bulk-move source updates above, an upload finishing, a tag edit, an extract, a delete) used to blank the whole listing to a loading skeleton for a moment before the new rows appeared — a visible flash, repeated for every tick of a long move. These same-folder refreshes now revalidate *in place*: the current rows stay on screen and the fresh data is swapped in when it arrives (changed rows just animate in/out), with no skeleton and no flash. Navigating to a different folder still shows the skeleton as before
- **Items being moved now shimmer in the listing** — when you start a move (the move tool, drag-and-drop onto a folder, or Cmd+V paste), the name of each item being moved away fills with an animated lilac→violet→blue gradient that sweeps across it for as long as the move is in flight, then stops the moment it settles. It's an always-present cue right where you're looking — handy for drag-and-drop, which doesn't always surface the floating transfer dock. A moved folder shimmers its own name, and if you **browse into it while it's still being moved** its whole subtree lights up too (every file and subfolder you've moved is marked, at any depth). Respects reduced-motion by falling back to a steady accent tint
- **The delete / undo toast was redesigned** — long file names used to run off the side of the screen because the toast never wrapped. The deleted-item toast now wraps the name (capped at two lines), sits in a polished card with the trash icon in a rounded chip, and is **dark orange** instead of the old neutral grey so a reversible delete reads as its own kind of action at a glance. The 5-second Undo button is unchanged

## v2.1.4 — Transfer dock visibility & selection persistence

- **The move/copy progress dock no longer stays hidden during long transfers** — a background move or copy (e.g. a cross-volume move between separate mounts) could run to completion with no floating progress indicator ever appearing, no matter how it was started — the move/copy tool, drag-and-drop onto a folder or breadcrumb, or Cmd+V paste all share the one dock. The dock's "don't flash a row for an instant rename" gate compared the browser's clock against the *server's* job timestamp; on a self-hosted setup where the server and the browser machine clocks differ by a few seconds, that mismeasured the elapsed time and suppressed the dock for the whole transfer. A running transfer's reveal is now timed by the client's own clock (skew-immune), while finished transfers still use the server-measured duration — so genuine transfers always show progress and instant same-volume renames still don't flash
- **Refreshing the listing keeps your selection** — pressing `/` to refresh, or any background reload (a settled transfer, an upload finishing, a tag change), used to clear every selected file, which was disruptive mid-task. A refresh that lands back in the same folder now restores the selection by re-matching the items; anything that was deleted or renamed simply drops out, and navigating to a different folder still clears it as before
- **A completed copy/move selects the new files, wherever you're viewing them** — when a transfer finishes, the items it produced are now selected automatically, as long as you're looking at the folder they landed in. This covers copy-in-place (the new `report(1).pdf` copies are selected and the originals deselected), the move/copy tool's **"open destination"** toggle (the copies are selected once you arrive and the copy actually finishes), and copying then navigating to another folder and pasting. The transfer reports its *resolved* destination names (version suffix and all), so exactly the right items are selected the moment the job settles — and if you've navigated to an unrelated folder, your current selection is left untouched. This is now handled in one place (the transfer dock) for every move/copy path instead of per-trigger
- **Corrected the example `compose.yaml`** — the repo's Compose file still pointed at upstream `filebrowser/filebrowser`, with a `/flux/vault` mount the image doesn't use and a Redis service that only applies to multi-instance setups. It now uses `ghcr.io/csummers-dev/filebrowser-pretty` with the image's real volumes (`/srv`, `/config`, `/database`), the non-root `1000:1000` user, the healthcheck, and a runnable named-volume default — so `docker compose up -d` just works (production data mounts + reverse proxy stay documented in the README)
- **Internal — dead-code cleanup + tests** — removed three orphaned components (`ProgressBar.vue`, `settings/Themes.vue`, `files/ExtendedImage.vue`) plus a handful of unused exports/functions left over from the rebuild (no behavior change). The transfer-dock reveal gate and the listing selection logic were pulled out into pure, unit-tested helpers, with added backend coverage for the job snapshot — closing test gaps around the fixes above

## v2.1.3 — Audio volume bar fill

- **The audio player's volume slider now fills lilac** — the track read as a flat grey bar with only the drag handle accented, out of step with the lilac scrubber and the rest of the player. The portion up to the handle now fills with the accent (lilac), matching the playback scrubber, across Chrome/Safari and Firefox

## v2.1.2 — Drag-drop into-folder precision

- **Dropping a file "into" a folder now requires the exact spot that highlights — on mouse *and* touch** — a file dropped anywhere over a folder's name could land *inside* that folder even when the row never highlighted, because the drop re-measured the target zone and the glyph-width measurement is unreliable during the drop event (it fell back to the full name column). The drop now reuses the same state that drives the highlight + spring-load countdown, so a folder only receives the drop where it actually lit up; releasing anywhere else drops alongside, into the current folder. Desktop and touch drag now share one drop-zone definition, so the tightening behaves identically across phone, tablet, and desktop

## v2.1.1 — Sidebar & move/paste fixes

- **Search bar no longer shoves the header around** — opening the search used to expand it inside the toolbar row, squeezing the folder title and meta into a tiny column. It now floats as a compact, centered overlay above the header while it's the active element instead of displacing anything; closing it restores the row untouched
- **Sidebar no longer overflows the page** — with enough pinned Favorites (or Recents), the sidebar used to grow past the bottom of the window, pushing the storage meter and the account/logout row off-screen. The header and the main folder link now stay fixed at the top, the storage card and the user/logout row stay fixed at the bottom, and only the Favorites + Recents lists scroll when they run long
- **Cmd+V paste now shows progress and refreshes** — pasting a cut/copied file ran silently through a different code path than the move/copy tool: no progress indicator, and the listing didn't refresh so the pasted file wasn't visible. Paste now uses the same background transfer as the move tool and drag-and-drop, so the floating transfer dock shows progress and the listing refreshes the moment the job finishes
- **Conflict dialog "Skip all" gives feedback** — when a move/copy/paste hit a name conflict and you chose **Skip all conflicting files**, the dialog just closed with no indication of what happened. It now shows a toast ("All conflicting items were skipped — nothing was moved/copied.") so it's clear the action completed with nothing to do — in both the paste flow and the move/copy tool

## v2.1.0 — Listing, drag & preview polish

- **Listing no longer hides behind the breadcrumb bar** — the bottom breadcrumb strip was sitting under an oversized scroll area, so the last rows/tiles could disappear behind it. The listing now fills its scroll section as a proper flex child (the section gained the missing `min-height: 0` so its overflow actually scrolls), leaving just the intended slim gap above the bar. The virtualized list view also shed a stray 1rem bottom padding that was lifting its scroll viewport off the breadcrumb, so list rows now run all the way down to the bar
- **Drag-highlight works from any direction** — hovering a file over a folder now lights it up fully no matter which side you approach from. Previously the full highlight only "took" when you came in from the right; from any other direction it fell back to the dimmed state. The un-dim is now re-asserted on every drag-over frame so the document-wide drag dim can't win on a timing race
- **Countdown ring stands out** — the spring-load "open this folder" countdown ring is now **vivid blue** instead of white, so it no longer blends into the white folder glyph on the amber tile
- **Text preview controls moved to the details panel** — the Rendered/Raw, Soft-wrap, and Edit controls left the floating bottom pill and now live as colorful, full-width buttons in the details rail, keeping the reading area clear
- **Exit button** — slightly smaller (the arrow sits closer to "Exit"), and in the text preview the reading card now drops below it so the two never overlap
- **Album art for `.opus`** — the audio player now shows cover art for Opus files. The client-side metadata reader (`music-metadata`) throws on Opus's embedded picture and aborts, so the player falls back to the server-extracted cover (which the backend reads reliably for every audio format)
- **Album art is a 10px squircle** — the square cover in the audio player gets softly rounded corners so it doesn't read as a flat slab (still square)
- **Kobo `.kepub.epub` books render** — Kobo kepub chapters showed blank pages. The cause: Kobo injects a self-closing `<script src="kobo.js"/>` into every chapter's `<head>`, and since epub.js parses spine documents (whose filenames end in `.html`) as HTML — where a self-closing `<script/>` isn't valid — the parser swallowed the rest of the chapter as unterminated script text, leaving an empty page. The reader now closes self-closing raw-text tags in spine documents before parsing (scoped to content documents, never the package/OPF XML), so kepub files page through normally like any other epub. This also fixes a kepub that's simply been renamed to `.epub`
- **Comic & PDF chrome auto-hides** — in comic and PDF previews the Exit and details buttons now fade out together with the controls pill after a spell of inactivity (and return on movement), for a cleaner full-screen read. Other previews keep them always visible
- **Conflict dialog "Decide" button** — recoloured from amber (which darkened to an unpleasant brown on the light surface) to a positive-action **green**, tuned per theme for AA contrast (a deepened green with a white label on light; the bright green token with a near-black label on dark), with its icon matching the label. Fixed a hover bug where the button washed out to an unreadable near-white on the light theme (the generic button-hover rule was overriding the accent fill; the accent hover now declares its own green)

## v2.0.0 — The "pretty" milestone

filebrowser pretty 2.0 is a top-to-bottom reimagining of upstream File Browser: a redesigned interface, in-browser viewers for every common format, real keyboard-driven navigation, tags, an audio-tag editor, admin tooling, and a deep performance pass. This single entry consolidates everything that defines the 2.0 product.

### Breaking / upgrade notes

- **English-only.** The legacy multi-language system — inaccurate translations carried over from upstream and incompatible with the reworked UI — was removed: all non-English locale bundles, the in-app language picker, the RTL machinery, and the per-user `Locale` field (model, settings, `--locale` flag, auth payloads) are gone. **Existing databases keep working** — a stored locale is simply ignored and dropped on next save.
- Otherwise drop-in: **no schema migration is required**, and existing users, shares, tags, and settings are preserved.

### Interface & design

- **A cohesive, modern shell** — rebuilt **List**, **Grid** (media cards with a thumbnail + name/meta footer), and **Gallery** views; a redesigned sidebar with Recent / Favorites / Tags + a storage meter; slide-over panels for move / copy / share; and a settings area with a full-height left rail where each page carries its own tinted icon + title + description. Full **light & dark themes**, with a six-preset **accent color** picker synced per user.
- **Header-less, unified layout** — the top toolbars are gone. The listing runs edge-to-edge into a slim floating control cluster (search · view · sort · direction · upload · More), with the directory breadcrumb in a quiet bar pinned to the bottom. Search is a compact "⌘K" pill that expands into a real input on focus. Sidebar → primary → details rails read as one continuous frosted surface.
- **Color, used deliberately** — file-type icons, per-action buttons (share / download / rename / move / copy / extract / delete), settings nav chips, the command palette, and the storage gradient are all color-coded for fast scanning, consistent across both themes.
- **Brand** — uniformly **filebrowser pretty** (lowercase) with a lilac (`#5e6ad2`) accent applied to the theme-color, favicon, and PWA manifest (an admin `branding.color` override still wins). Rename the storage root ("My files") to a custom label that syncs across devices, and give favorited folders friendly display names.

### File previews

- **Seven format viewers** — **image** (zoom / fit, basic rotate · flip · crop saved as a copy), **video** (themed video.js skin, subtitle upload, on-the-fly transcode with a live progress label), **audio**, **PDF** (full PDF.js chrome with a thumbnail rail and lazy, resolution-capped page rendering so huge documents stay sharp), **EPUB** (chapter list, remembered reading position, an independent dark-reading toggle), **CSV**, and **code / text**. Markdown renders, with a toggle back to raw.
- **Comic reader** — open a `.cbz` / `.cbr` and page through it (edge taps, chevrons, keyboard, fullscreen, right-to-left for manga); pages stream and pre-cache, and it resumes on the page you left off.
- **Floating preview shell** — previews are a docked region beside the sidebar (not a take-over modal): **Exit** floats top-left, a single **details** toggle floats top-right, and per-format controls (zoom / page / fit) sit in a bottom-center pill that auto-hides on inactivity like a media player. Arrow keys step between files; swipe does the same on touch (swipe-down closes a fit image).
- **Rich metadata + cover art** — EXIF for photos, ID3 / Vorbis for audio (with embedded album art shown full + square), and track info for video. The details-rail file icon matches the listing row exactly, and listing rows show real cover thumbnails for **audio** (album art), **EPUB**, **PDF** (first page), and **`.cbr`** (first page) — extracted server-side and disk-cached.

### Browsing & navigation

- **Finder-style keyboard navigation** — arrows / Home / End / Enter to move and open, Shift to range-select, and **type-ahead** to jump to a name as you type (a space extends a multi-word prefix). Plus `Cmd/Ctrl+A` select-all, `/` to refresh, PDF page keys, and audio prev/next-track keys. A shortcuts cheat-sheet overlay lists them.
- **Command palette + search** — a ⌘K palette with recents and quick actions, and a fast backend-backed file search with live, debounced results.
- **Power-user sidebar** — Recent and Favorites (drag-reorder, custom display names, graceful dead-pin handling), per-folder view-mode memory, multi-column sort (including by extension) with a persistent ascending/descending toggle, breadcrumb depth-ellipsis, and right-click context menus on rows and empty space.
- **Tags** — per-file, color-coded tags shown as inline dots on each row (alphabetical, expanding on hover) with a picker; tags follow files through renames and moves.

### File operations

- Inline **new folder / file**, inline **rename**, and **delete** with an optimistic **Undo** toast.
- **Move / copy** run in the background via a floating transfer indicator, with an optional "open destination" toggle; **bulk rename** offers pattern and find-and-replace modes with live preview and conflict highlighting.
- **Drag & drop** — spring-loaded folders open on hover (with a countdown ring), breadcrumb segments are drop targets (drag to any ancestor), a drag ghost shows what's moving, and grid / gallery support marquee selection. The drop-into-folder zone hugs the icon + name so dropping *alongside* into the current folder is easy.
- **Archive extraction** — extract `.zip` / `.7z` / `.rar` server-side, including **password-protected** archives (detect-and-prompt — the password travels in a request header, never the URL or audit log, and a wrong password leaves no half-extracted files behind), with an optional "delete the archive after success" toggle and a destination picker.
- **Uploads** — a redesigned upload dock with per-file rows, an aggregate progress bar, per-file cancel, and **resumable** transfers that survive a dropped connection; each row names its destination folder.

### Audio tag editing

- Edit **title, artist, album, album artist, year, track / disc, genre, composer, and comment**, plus embedded **cover art** (add / replace / remove), for **MP3, FLAC, M4A** (MP4 / AAC / ALAC), and the **Ogg** family (Vorbis / Opus) — single-file or **batch** (the editor applies only the fields you actually change across a selection). Writes are atomic (temp file → rename); the newer formats use TagLib compiled to WebAssembly via the pure-Go `wazero` runtime (no cgo). Requires the **modify** permission.

### Admin & operations

- **Audit log** recording file operations, shares, sign-ins, and settings changes, filterable by action / user / date / path.
- **Webhooks** that POST a JSON payload on file events, with per-endpoint event filters, a test button, last-delivery status, and retry-with-backoff.
- **Session management** with "sign out everywhere."

### Performance & platform

- **Faster large folders** — file type is taken from the extension whenever that's definitive, and the few remaining header reads run in parallel; a ~2,000-file NAS folder drops from ~15 s to a second or two (`--disable-type-detection-by-header` skips header reads entirely).
- **Virtual scrolling** keeps folders with tens of thousands of files smooth; **format viewers lazy-load** (~1.7 MB of pdfjs / video.js / ace / epub.js / music-metadata pulled out of first load); and an **offline app shell** loads without a connection, with named error states (server unreachable / permission denied / not found) and one-click retry.
- Server-side **video thumbnails** (ffmpeg) and the audio / EPUB / PDF / CBR cover pipeline are bundled in the Docker image and fall back to colored icons when a tool isn't present. Robust cover extraction — including a lenient ID3 fallback for MP3s whose malformed tag frames defeat the strict parser.

### Mobile

- Pull-to-refresh, camera-roll / photo upload, swipe between files (and swipe-down to close) in previews, a navigation drawer (Recent + collapsible Favorites), tap-to-select with an opt-in **Details** sheet, and a pinned header so the app chrome can't scroll off-screen. Login fields carry proper `name` attributes so password managers (Bitwarden / 1Password / iOS Keychain) detect and fill them.

### Build & internals

- **SemVer** releases published to GHCR (`X.Y.Z` plus rolling `X.Y` / `X` / `latest`, and `nightly` + dated snapshots from `main`), with the in-app version stamped from the git tag and OCI provenance labels.
- A design-token system (color / motion / radius); pure-Go media libraries (no cgo) so the static cross-platform builds are unaffected; and lint + type-check + unit tests on every push. Third-party licenses are tracked in [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md).
