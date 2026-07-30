<div align="center">

<img src="branding/banner.svg" width="900" alt="vitrine" />

# vitrine

Self-hosted file management.

</div>

- **List view** — Dense, scannable, color-coded.
- **Grid + Gallery** — Comfortable file listings with previews.
- **Split pane view** — Navigate multiple folders at once.
- **Image + EXIF** — Pixel-perfect previews with EXIF details.
- **Image** editor — Crop, rotate, and flip.
- **Video** — Streams with tracks and subtitles.
- **Audio** — Album art, embedded tags and artwork.
- **ID3 tag editor** — Edit id3tag metadata, one track or a whole album at once.
- **PDF** — Page through PDFs with thumbnail previews and sharp rendering.
- **EPUB** — Read EPUBs with a live chapter list and location remembered.
- **Comic** reader — Read CBZ and CBR, optionally read manga right-to-left.
- **Text** - Read and edit text files.
- **Markdown** — Markdown, renderedor raw.
- **Command palette** — ⌘K to open command palette, navigate files, folders, and actions.
- **Search-as-you-type** — Results the moment you start typing.
- **Breadcrumb navigation** — Move between folders without backtracking.
- **Context menu** — Context-aware ight-click options throughout.
- **Favorites** + Recents — Pin favorites, rename them, track recent files.
- **Drag & drop** — Drag to move. Folders spring open, quick access to parent directory.
- **Multi-select + lasso** — Select in bulk, click through or lasso a whole region.
- **Move/Copy + transfer dock** — Moves and copies run in the background.
- **Bulk rename** — Rename multiple files at once with a live preview.
- **Conflict resolution** — File and folder conflicts handled. Keep, replace, or skip.
- **Tagging** — Tag anything in color. Persistent across moves.
- **Extract** — Extract ZIP, 7z, and RAR files in-app.
- **Password-protected archive** — Support for extracting password-protected archives.
- **Upload dock** — Displays per-file progress, and resumes if the connection drops.
- **Themes** — Hand-crafted light and dark themes.
- **User Management** — Manage users and permissions.
- **Settings** — Global and per-user settings.
- **Audit log** — Actions taken in-app recorded to an audit log available to administrators.
- **Webhooks** — Fire a webhook on any file event.
- **Sessions** — View active sessions and sign out everywhere.
- **Shares** — Share a link in seconds.
- **Mobile** — Mobile design prioritized.
- **Mobile upload / pull-to-refresh** — Upload from your camera roll; pull to refresh.

[![Version](https://img.shields.io/badge/version-3.1.0-6e72d9?style=flat-square)](#)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](#)
[![Vue](https://img.shields.io/badge/Vue-3.5-42b883?style=flat-square&logo=vue.js&logoColor=white)](#)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white)](#)
[![License](https://img.shields.io/badge/license-Apache_2.0-blue?style=flat-square)](LICENSE)

</div>

## Screenshots

<p align="center">
  <img src="docs/screenshots/hero.png" width="100%" alt="vitrine — gallery view" />
</p>

<p align="center">
  <img src="docs/screenshots/listing-list-light.png" width="49%" alt="List view (light)" />
  <img src="docs/screenshots/listing-list-dark.png" width="49%" alt="List view (dark)" />
</p>

<p align="center">
  <img src="docs/screenshots/listing-grid-light.png" width="49%" alt="Grid view (light)" />
  <img src="docs/screenshots/listing-grid-dark.png" width="49%" alt="Grid view (dark)" />
</p>

<p align="center">
  <img src="docs/screenshots/listing-gallery-light.png" width="49%" alt="Gallery view (light)" />
  <img src="docs/screenshots/listing-gallery-dark.png" width="49%" alt="Gallery view (dark)" />
</p>

<p align="center">
  <img src="docs/screenshots/split-view.png" width="80%" alt="Split (dual-pane) view" />
</p>

<p align="center">
  <img src="docs/screenshots/command-palette.png" width="80%" alt="Command palette (⌘K)" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-image.png" width="80%" alt="Image preview with EXIF" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-video.png" width="80%" alt="Video preview" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-audio.png" width="80%" alt="Audio preview with album art" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-pdf.png" width="80%" alt="PDF preview" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-epub.png" width="80%" alt="EPUB reader" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-comic.png" width="80%" alt="Comic (CBZ / CBR) reader" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-text.png" width="80%" alt="Text preview" />
</p>

<p align="center">
  <img src="docs/screenshots/preview-markdown.png" width="80%" alt="Markdown preview" />
</p>

<p align="center">
  <img src="docs/screenshots/edit-image.png" width="80%" alt="Image editor" />
</p>

<p align="center">
  <img src="docs/screenshots/edit-markdown.png" width="80%" alt="Markdown editor" />
</p>

<p align="center">
  <img src="docs/screenshots/edit-tags.png" width="80%" alt="Audio tag editor" />
</p>

<p align="center">
  <img src="docs/screenshots/mobile-list-view.png" width="32%" alt="Mobile list view" />
  <img src="docs/screenshots/mobile-grid-view.png" width="32%" alt="Mobile grid view" />
  <img src="docs/screenshots/mobile-gallery-view.png" width="32%" alt="Mobile gallery view" />
</p>

<p align="center">
  <img src="docs/screenshots/user-settings.png" width="80%" alt="Profile settings" />
</p>

<p align="center">
  <img src="docs/screenshots/global-settings.png" width="80%" alt="Global settings" />
</p>

<p align="center">
  <img src="docs/screenshots/login.png" width="80%" alt="Login" />
</p>

---

The published image lives at **`ghcr.io/csummers-dev/vitrine:latest`** and works on any Linux x86_64 host (NAS, mini-PC, VPS, homelab).

### Docker Compose

```yaml
services:
  filebrowser:
    image: ghcr.io/csummers-dev/vitrine:latest
    container_name: vitrine
    restart: unless-stopped
    ports:
      - "8080:80"
    volumes:
      # Your real data — mount user-owned SUBDIRECTORIES,
      # not whole NAS volume roots (those are owned by root and you'll hit
      # 403s on uploads).
      - /path/to/your/movies:/srv/Movies
      - /path/to/your/music:/srv/Music
      - /path/to/your/downloads:/srv/Downloads
      # Filebrowser's own state — keep on a fast disk if you have one
      - ./filebrowser/database:/database
      - ./filebrowser/config:/config
    environment:
      # The host user that owns your mounted data — run `id -u` / `id -g`.
      PUID: 1000
      PGID: 1000
      TZ: America/Los_Angeles
    healthcheck:
      test: ["CMD", "/healthcheck.sh"]
      interval: 30s
      timeout: 5s
      retries: 3
```

#### Permissions

On startup the container reads `PUID`/`PGID`, fixes the ownership of its own data to match, then **drops to that unprivileged user** to run the app — so filebrowser never runs as root, and you never have to `chown` your media or juggle a `user:` line. Just point it at whoever owns your files:

```yaml
    environment:
      PUID: 1000   # host user ID that owns your mounted data  (run `id -u`)
      PGID: 1000   # its group ID                              (run `id -g`)
```

- The app's own `/config` + `/database` are re-owned to `PUID:PGID` automatically on every start, so the database is always writable — even if Docker first created those folders as `root`.
- Your media under `/srv` is **not** touched: mount folders already owned by `PUID` (use user-owned **subdirectories**, not root-owned NAS volume roots, or uploads 403).
- **Need strictly non-root** (no root at all, e.g. a hardened host)? Set `user: "<uid>:<gid>"` on the service instead — the container detects it, skips the root/ownership step, and runs only as that user (you manage folder ownership yourself).

Behind a reverse proxy (Traefik shown here — adapt for Caddy / nginx / your stack):

```yaml
    networks: [web]
    labels:
      - traefik.enable=true
      - traefik.docker.network=web
      - traefik.http.routers.vitrine.rule=Host(`files.yourdomain.com`)
      - traefik.http.routers.vitrine.entrypoints=websecure
      - traefik.http.routers.vitrine.tls.certresolver=letsencrypt
      - traefik.http.services.vitrine.loadbalancer.server.port=80
```

Then:

```bash
docker compose up -d filebrowser
docker compose logs filebrowser | grep "password for"
```

Or skip Docker entirely and run the binary directly: `./filebrowser` — opens on <http://localhost:8080>.

## Architecture

```
┌────────────────────────────────────────────────────────────────┐
│  Browser                                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Vue 3 + TypeScript + Pinia                              │  │
│  │  • Composition API, <script setup>                       │  │
│  │  • Tailwind v4 design tokens                             │  │
│  │  • Composables for shortcuts, drag, focus, theme         │  │
│  │  • Format viewers: pdfjs / videojs / vue-reader (EPUB) / │  │
│  │    comic (CBZ·CBR) / marked (Markdown) /                 │  │
│  │    music-metadata / exifr                                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                       │ HTTP / WebSocket                       │
│                               ▼                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Go backend                                              │  │
│  │  • Gorilla mux router                                    │  │
│  │  • Storm v3 / bbolt — users · shares · settings ·        │  │
│  │    tags · audit · webhooks · sessions                    │  │
│  │  • afero filesystem abstraction                          │  │
│  │  • JWT auth (or proxy / noauth / hooks)                  │  │
│  │  • Embedded frontend assets (one binary deploys)         │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

### Tech stack

| Layer | Choice |
| --- | --- |
| Backend | **Go 1.25** |
| DB | **Storm v3 / bbolt** |
| Frontend | **Vue 3 + TypeScript** |
| State | **Pinia** |
| Styling | **Tailwind v4** |
| Build | **Vite + Rolldown** |
| Routing | **vue-router 5** |
| i18n | **vue-i18n** |
| Uploads | **TUS (tus-js-client)** |
| Virtual list | **vue-virtual-scroller** |
| PDF | **pdfjs-dist 6** |
| Video | **video.js 8** + **ffmpeg** thumbnails |
| Audio metadata (read) | **music-metadata 11** |
| Audio tags (read/write) | **id3v2 + go-flac** (MP3 / FLAC) · **TagLib**/Wasm (M4A / OGG) |
| Archives | **mholt/archives** (zip / 7z / rar / tar) · **yeka/zip** (password-protected ZIP) |
| EPUB | **vue-reader + epub.js** |
| Code / text edit | **native `<textarea>`** |
| Markdown | **marked + KaTeX** |
| EXIF | **exifr 7** |

---

## Credits

This project began as a fork of [filebrowser/filebrowser](https://github.com/filebrowser/filebrowser) by [@hacdias](https://github.com/hacdias) and the original maintainers.

---

## License

Apache License 2.0. See [LICENSE](LICENSE). Third-party components bundled in the binary — notably **TagLib** (LGPL-2.1), used for M4A / OGG tag editing — are listed in [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md).

---

<div align="center">

**Built with love by csummers-dev.**

</div>
