# Quickstart: Gigtape (Phase 1)

**Purpose**: Validate the complete Phase 1 pipeline end-to-end in local development.
**Date**: 2026-04-21

---

## Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Go | 1.22+ | `go version` to verify |
| Node | 20+ | `node --version` to verify |
| setlist.fm API key | — | Register at setlist.fm/settings/api |
| Spotify Developer app | — | Create at developer.spotify.com/dashboard |

---

## Environment Setup

Create `.env` in the repo root (git-ignored):

```env
SETLISTFM_API_KEY=your_setlistfm_key_here
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
SPOTIFY_REDIRECT_URI=http://localhost:8080/auth/callback
SESSION_TTL_MINUTES=60
SENTRY_DSN=
LOG_FORMAT=text
```

In your Spotify Developer Dashboard, add `http://localhost:8080/auth/callback` as an allowed
redirect URI.

---

## Initialize Go Workspace

Run once after cloning:

```bash
go work init
go work use \
  ./packages/domain \
  ./packages/usecases \
  ./packages/adapters/setlistfm \
  ./packages/adapters/spotify \
  ./apps/api \
  ./apps/cli
```

---

## Run the API Server

```bash
cd apps/api
go run .
# Listening on http://localhost:8080
```

---

## Run the Vue Frontend (separate terminal)

```bash
cd apps/web
npm install
npm run dev
# Vite dev server on http://localhost:5173
# /api/* requests proxied to http://localhost:8080
```

---

## Validate the CLI First

The CLI validates the entire core pipeline before the web UI exists. Run these checks in order:

```bash
# Build the CLI
cd apps/cli
go build -o gigtape .

# 1. Authenticate with Spotify
./gigtape auth
# Opens browser → complete OAuth flow → prints "Authenticated as {name}"

# 2. Single artist flow
./gigtape artist "Radiohead"
# Expected output:
#   Found: Radiohead (rock band from Abingdon, UK) [y/n]: y
#   Setlist: Glastonbury 2024 (12 songs)
#   Attribution: setlist.fm • https://...
#   Creating playlist...
#   ✓ Playlist created: https://open.spotify.com/playlist/...
#   ✓ 12 songs added
#   ✗ 0 songs not found

# 3. Festival flow
./gigtape festival "Glastonbury 2024"
# Expected output:
#   Found: Glastonbury 2024 — Worthy Farm, UK
#   Lineup (8 artists found, lineup may be incomplete):
#     1. Coldplay (18 songs)
#     2. SZA (14 songs)
#     ...
#   Mode [merged/per-artist]: merged
#   Creating playlist...
#   ✓ Playlist created: https://open.spotify.com/playlist/...
```

---

## Validation Checklist

Run through these in order. Each item must pass before moving on.

### Core Pipeline

- [ ] `go test ./packages/domain/...` passes with zero failures
- [ ] `go test ./packages/usecases/...` passes with zero failures (uses fakes, no live APIs)
- [ ] `go test ./packages/adapters/setlistfm/...` passes (unit tests; no live API)
- [ ] `go test ./packages/adapters/spotify/...` passes (unit tests; no live API)
- [ ] `go test ./apps/api/...` passes (handler tests with httptest)

### CLI Smoke Test

- [ ] `gigtape auth` completes OAuth and confirms Spotify account name
- [ ] `gigtape artist "Radiohead"` returns a setlist with ≥1 song
- [ ] A private playlist appears in the authenticated Spotify account
- [ ] The playlist link is valid and opens Spotify
- [ ] Songs not found on Spotify are listed explicitly (not silently skipped)
- [ ] `gigtape festival "Glastonbury"` returns a lineup with ≥2 artists
- [ ] Merged playlist contains songs from multiple artists in lineup order
- [ ] `source_attribution` is printed to terminal wherever setlist data appears

### Web UI Smoke Test

- [ ] OAuth flow completes in browser; session established
- [ ] Artist search returns results with disambiguation info
- [ ] Setlist preview shows songs with attribution
- [ ] Track removal from preview is reflected in the created playlist
- [ ] Playlist link returned after creation opens correctly in Spotify
- [ ] Festival flow: lineup loads, artist deselection works, playlist created

### Resilience Checks

- [ ] Searching for an artist with no setlists returns empty state (not an error page)
- [ ] Searching for an unknown artist returns empty state with "add manually" option
- [ ] Searching for an unknown festival returns empty state with "add artists manually" option
- [ ] Songs not found on Spotify appear in the result summary, not swallowed

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| 429 from Spotify | Rate limit hit | Wait and retry; verify backoff is implemented |
| Empty setlist returned | Artist has no recent shows | Try a more active artist |
| Playlist not created | Wrong OAuth scope | Verify `playlist-modify-private` in scope list |
| Session expired mid-flow | TTL too short | Increase `SESSION_TTL_MINUTES` |
| OAuth callback fails | Redirect URI mismatch | Check `SPOTIFY_REDIRECT_URI` matches dashboard setting |
| Festival lineup empty | Festival not in setlist.fm | Try manual artist entry workflow |
| Songs in wrong order in merged playlist | Lineup order bug | Verify `Event.Artists` order preserved through use case |

---

## Integration Tests (Optional)

Requires real API keys. Not run by default:

```bash
export RUN_INTEGRATION=true
go test -tags integration ./packages/adapters/...
```

These tests hit the live setlist.fm and Spotify APIs. Run sparingly — they consume API quota.
