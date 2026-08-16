# syntax=docker/dockerfile:1

## Frontend builder: build the embedded SPA ONCE on the native build platform
## (its output is architecture-independent), then reuse it for every target
## arch. Pinned pnpm matches the repo's packageManager field.
FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend
WORKDIR /app/frontend
RUN npm install -g pnpm@10.34.1
# Dependency layer first (cached until the lockfile changes), then the build.
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
# → /app/frontend/dist, embedded into the Go binary via frontend/assets.go.
RUN pnpm run build

## Backend builder: cross-compile the Go binary for the requested target arch.
## Runs on the NATIVE build platform (Go cross-compiles, so no slow emulation)
## and emits a static, CGO-free binary for $TARGETOS/$TARGETARCH. VERSION /
## REVISION are stamped into the version package, mirroring `task build:backend`.
FROM --platform=$BUILDPLATFORM golang:1.25.13-alpine AS backend
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG REVISION=unknown
# Module-download layer — only re-runs when go.mod / go.sum change.
COPY go.mod go.sum ./
RUN go mod download
# Source + the prebuilt frontend assets (go:embed all:dist needs frontend/dist).
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags="-s -w -X 'github.com/csummers-dev/vitrine/v3/version.Version=${VERSION}' -X 'github.com/csummers-dev/vitrine/v3/version.CommitSHA=${REVISION}'" \
    -o /out/vitrine .

## Fetch runtime helper files (ca-certificates, mailcap, tini-static, JSON.sh).
FROM alpine:3.23 AS fetcher

# install and copy ca-certificates, mailcap, and tini-static; download JSON.sh
RUN apk update && \
    apk --no-cache add ca-certificates mailcap tini-static && \
    wget -O /JSON.sh https://raw.githubusercontent.com/dominictarr/JSON.sh/0d5e5c77365f63809bf6e77ef44a1f34b0e05840/JSON.sh

## Static ffmpeg for server-side video thumbnails (v1.3 S6-2). A single
## self-contained, multi-arch static binary — the cleanest way to add
## ffmpeg to the BusyBox image (no package manager, no shared libs). The
## app runtime-detects it and falls back to the generic video icon when
## absent, so this stays a convenience, never a hard dependency.
FROM mwader/static-ffmpeg:7.1 AS ffmpeg

## Final stage: lightweight Alpine runtime. Alpine (musl, like the previous
## BusyBox base) ships a package manager + busybox applets, so the existing
## init/healthcheck scripts keep working AND we can install poppler-utils for
## PDF cover thumbnails. pdftoppm is runtime-detected; absent → PDF rows fall
## back to the generic icon, so this stays a convenience, never a hard dep.
FROM alpine:3.23

# OCI provenance labels. VERSION / REVISION / CREATED are passed at build time
# (CI fills them from the git tag, commit sha, and build timestamp; a plain
# `docker build` gets the defaults below). In CI, docker/metadata-action also
# injects its own labels, which take precedence over these.
ARG VERSION=dev
ARG REVISION=unknown
ARG CREATED=""
LABEL org.opencontainers.image.title="vitrine" \
      org.opencontainers.image.description="A modern, polished File Browser fork" \
      org.opencontainers.image.source="https://github.com/csummers-dev/vitrine" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${CREATED}"

# Default UID/GID for the runtime user. The entrypoint (init.sh) overrides these
# at run time from the PUID/PGID env vars, so end users point the container at
# whatever uid owns their mounted data without rebuilding.
ENV UID=1000
ENV GID=1000

# poppler-utils → pdftoppm (PDF first-page rasterizer for row thumbnails);
# ttf-dejavu so text-heavy first pages render legibly; su-exec → drop from root
# to PUID:PGID in the entrypoint. Then create the default (1000) user.
RUN apk --no-cache add poppler-utils ttf-dejavu su-exec && \
    addgroup -g $GID user && \
    adduser -D -u $UID -G user user

# Copy binary (cross-compiled in the `backend` stage for THIS image's arch),
# scripts, and configurations into image with proper ownership.
COPY --chown=user:user --from=backend /out/vitrine /bin/vitrine
# S6-2: static ffmpeg → /usr/local/bin (on PATH for video-thumbnail
# generation). If PATH ever misses it, detection just fails gracefully.
COPY --from=ffmpeg /ffmpeg /usr/local/bin/ffmpeg
# #3: ffprobe enables the transcode remux fast-path (probe codecs → stream-
# copy when already H.264/AAC). Optional — without it everything is fully
# transcoded, which is correct, just slower.
COPY --from=ffmpeg /ffprobe /usr/local/bin/ffprobe
COPY --chown=user:user docker/common/ /
COPY --chown=user:user docker/alpine/ /
COPY --chown=user:user --from=fetcher /sbin/tini-static /bin/tini
COPY --from=fetcher /JSON.sh /JSON.sh
COPY --from=fetcher /etc/ca-certificates.conf /etc/ca-certificates.conf
COPY --from=fetcher /etc/ca-certificates /etc/ca-certificates
COPY --from=fetcher /etc/mime.types /etc/mime.types
COPY --from=fetcher /etc/ssl /etc/ssl

# Create data directories, set ownership, and ensure healthcheck script is executable
RUN mkdir -p /config /database /srv && \
    chown -R user:user /config /database /srv \
    && chmod +x /healthcheck.sh

# Define healthcheck script
HEALTHCHECK --start-period=2s --interval=5s --timeout=3s CMD /healthcheck.sh

# NOTE: no `USER` here on purpose. The container starts as root so the
# entrypoint (init.sh) can chown the app's own data to PUID:PGID and then drop
# to that unprivileged user via su-exec before running vitrine — so the app
# itself never runs as root. Operators who require strictly-non-root (no root at
# all) can set `user:` at run time; init.sh detects that and skips the root step.
VOLUME /srv /config /database

EXPOSE 80

ENTRYPOINT [ "tini", "--", "/init.sh" ]
