# Gigtape

Create private Spotify playlists from setlist.fm data. Given an artist, Gigtape
fetches recent setlists, lets you preview and edit the track list, and creates
a private playlist with a direct link.

Built in Go with Hexagonal Architecture (Ports & Adapters), a Gin REST API, a
Cobra CLI, and a Vue 3 SPA.

## Status

Progress against [specs/001-gigtape-baseline/tasks.md](specs/001-gigtape-baseline/tasks.md):

| Phase | Description | Status |
|---|---|---|
| 1 | Setup (workspace, modules, scaffolds) | ✅ Done |
| 2 | Foundational (domain, session, OAuth shell) | ✅ Done |
| 3 | User Story 1 — Single-artist playlist (MVP) | ✅ Done |
| 4 | User Story 2 — Festival playlist | ✅ Done |
| 5 | **Polish (Sentry, rate limit, audit)** | ✅ Done |

Phase 1 of the baseline is complete. What you can do today:

- Authenticate with Spotify (PKCE OAuth)
- **Single-artist flow** — search → preview/edit setlist → create private playlist
- **Festival flow** — search festival → review lineup with per-artist song counts
  → deselect artists / add manual ones → choose **merged** (one playlist) or
  **per-artist** (one playlist per artist) mode → receive Spotify links
- See which tracks Spotify couldn't match (never silently dropped)
- See which artists were skipped (deselected or no setlist data)

Phase 5 added:

- **Sentry** SDK initialization in API + CLI; unexpected adapter errors are
  captured at the use-case boundary through an `ErrorReporter` port (the
  usecases package has zero Sentry imports). Set `SENTRY_DSN` to enable; leave
  empty to disable without code changes.
- **Per-session rate limit** on all protected API routes — token bucket at
  ~2 req/s, burst 4, keyed by `X-Session-ID`. 429 `rate_limited` with a
  `Retry-After: 1` header when the bucket is empty.
- **Structured error logging** at use-case boundaries with `slog.Error`
  (fields: `use_case`, `artist`/`event`, `session_id` on API, `error`).
- **Error-path audit**: every `_` discard of an error in the adapters was
  either replaced with explicit handling or annotated with a `slog.Warn`
  fallback (e.g. unparseable setlist.fm dates).
- **Attribution audit**: `SourceAttribution` is rendered everywhere setlist
  data appears — CLI (`artist`, `festival`), web (SetlistPreview, per-row
  in FestivalSearch).

## Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| Go | 1.22+ | `go version` |
| Node | 20+ | `node --version` (only for the web UI) |
| setlist.fm API key | — | [setlist.fm/settings/api](https://www.setlist.fm/settings/api) |
| Spotify Developer app | — | [developer.spotify.com/dashboard](https://developer.spotify.com/dashboard) |

In your Spotify Developer Dashboard, add **both** of these as allowed Redirect URIs:

- `http://localhost:8080/auth/callback` — used by the web + API flow
- `http://127.0.0.1:<dynamic-port>/callback` — used by the CLI (ephemeral
  loopback; you will paste the exact port printed at login time, see below)

## Environment

Copy the example and fill in the blanks:

```bash
cp .env.example .env
```

Required keys (see [.env.example](.env.example)):

```env
SETLISTFM_API_KEY=your_setlistfm_key_here
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret       # unused by PKCE, keep for future
SPOTIFY_REDIRECT_URI=http://localhost:8080/auth/callback
SESSION_TTL_MINUTES=60
SENTRY_DSN=                                             # leave empty to disable Sentry
SENTRY_ENVIRONMENT=development                          # optional; defaults to "development"
SENTRY_RELEASE=gigtape@dev                              # optional; defaults to gigtape@dev
LOG_FORMAT=text                                         # or "json"
```

Load the vars before running anything:

```bash
set -a && source .env && set +a
```

## Workspace setup

One-time sync after cloning:

```bash
go work sync
```

Domain-purity check (must show only stdlib imports):

```bash
( cd packages/domain && go list -f '{{ join .Imports "\n" }}' ./... | sort -u )
# expected: context, fmt, time
```

## Build & run

The `Makefile` at the repo root wraps the common commands:

```bash
make build-api     # go build ./apps/api
make build-cli     # go build -o gigtape ./apps/cli
make serve-web     # cd apps/web && npm run dev
make test          # go test ./packages/...
```

### API server

```bash
go run ./apps/api
# Listening on :8080
```

Quick smoke test (no session required):

```bash
curl -s http://localhost:8080/auth/login | jq
# { "auth_url": "https://accounts.spotify.com/authorize?..." }
```

### CLI

```bash
go build -o gigtape ./apps/cli
./gigtape --help
```

### Web UI

```bash
cd apps/web
npm install      # first time only
npm run dev
# Vite dev server on http://localhost:5173
# /api/* is proxied to http://localhost:8080
```

## How to test Phase 3 & 4 — single-artist and festival playlists

### Option A — CLI (fastest)

This is the flow validated by Phase 3's checkpoint.

```bash
# 1. Authenticate with Spotify (opens browser, waits on ephemeral loopback port)
./gigtape auth
# → "Opening browser for Spotify authorization…"
# → Complete consent in browser.
# → "Authenticated as <your_spotify_user_id>"
#
# Token is cached at $TMPDIR/gigtape-token.json until it expires.

# 2. Create a playlist from an artist's latest setlist
./gigtape artist "Radiohead"
#
# Expected interaction:
#   Multiple artists found:
#     1. Radiohead — English rock band
#     2. Radiohead — tribute band
#   Pick number: 1
#
#   Setlist: Coachella 2024 (17 songs, 2024-04-12)
#   Attribution: setlist.fm • https://www.setlist.fm/setlist/...
#
#      1. Let Down
#      2. Lucky
#      ...
#
#   Remove tracks by number (comma-separated), or Enter to keep all:
#   Create playlist with 17 tracks? [y/n]: y
#
#   ✓ Playlist created: https://open.spotify.com/playlist/...
#   ✓ 17 songs added
#   ✗ 0 songs not found
```

**Things to verify:**

- A private playlist appears in your Spotify account named
  `"<Artist> — <YYYY-MM-DD>"`.
- `SourceAttribution` is printed after the setlist preview (setlist.fm policy,
  FR-013).
- If any track has no Spotify match, it's listed under "✗ N songs not found on
  Spotify" — never silently dropped.
- If the most recent setlist has fewer than 6 songs, you'll see a warning.

Festival flow (CLI):

```bash
./gigtape festival "Glastonbury 2024"
#
# Expected interaction:
#   Multiple events found:
#     1. Glastonbury Festival — 2024-06-28 (Pilton, United Kingdom, 40 artists)
#     ...
#   Pick number: 1
#
#   Lineup (40 artists found, lineup may be incomplete):
#      1. Coldplay (22 songs)
#           attribution: setlist.fm • https://...
#      2. SZA (18 songs)
#           attribution: setlist.fm • https://...
#      3. Some Opener (no setlist found)
#     ...
#
#   Deselect artists by number (comma-separated), or Enter to include all: 3
#   Mode [merged/per-artist]: merged
#
#   ✓ Playlist #1: https://open.spotify.com/playlist/...
#     40 songs added
#     Skipped artists: Some Opener
```

Per-artist mode produces one playlist per included artist and prints each URL.
Artists with no setlist (or explicitly deselected) are listed under "Skipped
artists". FR-013 compliance: each artist's `SourceAttribution` is printed next
to its song count.

### Option B — API directly

The REST contract is in [specs/001-gigtape-baseline/contracts/api.md](specs/001-gigtape-baseline/contracts/api.md).

OAuth with the API requires a browser round-trip. The simplest path:

```bash
# 1. Get the auth URL and open it
AUTH_URL=$(curl -s http://localhost:8080/auth/login | jq -r .auth_url)
open "$AUTH_URL"      # macOS; on Linux use xdg-open

# 2. Complete consent in browser. Spotify redirects to
#    http://localhost:8080/auth/callback?code=…&state=…
#    The API responds with { "session_id": "<uuid>" } — copy it.

SESSION_ID=<paste-uuid-here>

# 3. Search for an artist
curl -s "http://localhost:8080/artists/search?q=Radiohead" \
     -H "X-Session-ID: $SESSION_ID" | jq

# 4. Pull recent setlists (use external_ref from step 3)
ARTIST_REF=a74b1b7f-71a5-4011-9441-d0b5e4122711
curl -s "http://localhost:8080/setlists?artist_ref=$ARTIST_REF&artist_name=Radiohead" \
     -H "X-Session-ID: $SESSION_ID" | jq

# 5. Create the playlist (pass the tracks you want)
curl -s -X POST http://localhost:8080/playlists/artist \
     -H "X-Session-ID: $SESSION_ID" \
     -H "Content-Type: application/json" \
     -d '{
       "artist_ref":"'"$ARTIST_REF"'",
       "artist_name":"Radiohead",
       "setlist_index":0,
       "event_date":"2024-04-12",
       "tracks":[
         {"title":"Creep","artist_name":"Radiohead"},
         {"title":"Karma Police","artist_name":"Radiohead"}
       ]
     }' | jq
```

Festival flow:

```bash
# 1. Search festivals (the query may include a year — it's parsed out and
#    passed to setlist.fm's `year` parameter; the remainder goes to venueName).
curl -s "http://localhost:8080/events/search?q=Glastonbury%202024" \
     -H "X-Session-ID: $SESSION_ID" | jq

# 2. Create merged festival playlist
curl -s -X POST http://localhost:8080/playlists/festival \
     -H "X-Session-ID: $SESSION_ID" \
     -H "Content-Type: application/json" \
     -d '{
       "event_name":"Glastonbury 2024",
       "event_date":"2024-06-28",
       "mode":"merged",
       "artists":[
         {"artist_ref":"...mbid-coldplay...","artist_name":"Coldplay","include":true,
          "tracks":[{"title":"Yellow","artist_name":"Coldplay"}]},
         {"artist_ref":"...mbid-sza...","artist_name":"SZA","include":false,"tracks":[]}
       ]
     }' | jq
#
# Per-artist mode: set "mode":"per_artist". Response returns HTTP 207 when at
# least one playlist in the batch succeeded and at least one failed.
```

Error shapes are documented in `contracts/api.md`. 401 codes:
`session_not_found` and `session_expired`.

### Option C — Web UI

The SPA covers the same flow.

1. `npm run dev` in `apps/web` (and `go run ./apps/api` in another terminal).
2. Open `http://localhost:5173`.
3. Click **Connect Spotify** → complete consent in browser.
4. Spotify redirects to the API's `/auth/callback`, which currently returns
   `{"session_id":"<uuid>"}` as JSON. **Phase 3 shim:** copy the UUID, return
   to `http://localhost:5173`, expand *"Already authenticated? Paste session
   ID"*, paste it, click **Use session**. (A backend-to-SPA redirect with
   `?session_id=` is already supported — it just isn't emitted by the API yet.
   Slotted for a later phase.)
5. Search an artist → pick a disambiguation → preview the setlist (with
   attribution) → remove tracks or add manual ones → **Create Playlist**.
6. Result view shows the Spotify link, matched count, and an explicit
   "Not found on Spotify" list.

Festival flow (web): from the home page click **"Festival mode →"**, or go
directly to `http://localhost:5173/festival`. Search → pick an event → review
the lineup (checkboxes per artist, song count shown, "no setlist found"
indicator) → add manual artists if needed → **Continue** → choose **merged**
or **per-artist** mode → **Create Playlists**. Result view shows one card per
created playlist plus a "Skipped artists" panel.

## Endpoints available

| Method | Path | Purpose |
|---|---|---|
| GET | `/auth/login` | Start PKCE OAuth; returns `{ auth_url }` |
| GET | `/auth/callback` | OAuth callback; returns `{ session_id }` |
| GET | `/artists/search?q=` | Artist disambiguation candidates |
| GET | `/setlists?artist_ref=&artist_name=` | Recent setlists (desc by date) |
| POST | `/playlists/artist` | Create playlist from artist's setlist |
| GET | `/events/search?q=` | Festival/event search (venueName + optional year) |
| POST | `/playlists/festival` | Create merged or per-artist festival playlist(s) |

HTTP 207 Multi-Status is returned by `/playlists/festival` in `per_artist`
mode when at least one playlist succeeded and at least one failed.

## Tests

No test files are committed in Phase 3 (tasks.md did not request them, but the
fakes infrastructure in `packages/usecases/fakes/` is in place for when they
are added).

Build check across all modules:

```bash
go build ./apps/api ./apps/cli
( cd packages/domain && go build ./... )
( cd packages/usecases && go build ./... )
( cd packages/adapters/setlistfm && go build ./... )
( cd packages/adapters/spotify && go build ./... )
```

## Project layout

See [specs/001-gigtape-baseline/plan.md](specs/001-gigtape-baseline/plan.md)
for the full rationale.

```
packages/
  domain/              stdlib-only entities + port interfaces
  usecases/            use case logic + hand-rolled fakes for tests
  adapters/
    setlistfm/         SetlistProvider implementation
    spotify/           PlaylistDestination + OAuth PKCE
apps/
  api/                 Gin REST server (composition root: main.go)
  cli/                 Cobra CLI (composition root: main.go)
  web/                 Vue 3 + Vite SPA
```

The domain package has zero external imports — the Go compiler enforces this.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `401 session_not_found` | No or expired `X-Session-ID` | Re-run `/auth/login` flow |
| `401 session_expired` | OAuth token TTL exceeded | Re-authenticate |
| CLI: `not authenticated` | Token file missing | Run `./gigtape auth` first |
| CLI browser: OAuth redirect error | Loopback URI not allowed | Add `http://127.0.0.1:<port>/callback` to Spotify app (the exact port is printed when `gigtape auth` starts) |
| Playlist not created | Missing OAuth scope | Verify `playlist-modify-private` in the Spotify app config |
| Web: "Please connect Spotify first" | Session ID not captured | Use the "Paste session ID" shim (see Option C, step 4) |
| 429 from setlist.fm / Spotify | Rate limit (upstream) | Backoff is implemented; retry after a few seconds |
| 429 `rate_limited` from the Gigtape API | Per-session rate limit (~2 req/s) | Honour the `Retry-After` header, or space requests |
| Setlist has 0 songs / "No setlists found" | Artist has no recent shows on setlist.fm | Try a more active artist; manual track entry works in the web UI |

## Next

Phase 1 of the baseline is complete. Remaining work lives in `plan.md` as
Phase 2 / 3 placeholders (Ticketmaster + Apple Music adapters, Telegram +
Discord surfaces, `DiscoverUpcomingConcerts` use case) — not scheduled yet.

## Known caveats

- The session-ID handoff from API callback to SPA is still manual — backend
  `/auth/callback` returns JSON; the SPA offers a "Paste session ID" fallback.
  A redirect-to-SPA with `?session_id=` is wired in the client but not yet
  emitted by the API handler.
- setlist.fm has no `eventName` parameter on `/search/setlists`. The event
  provider parses a year out of the query (e.g. `"Glastonbury 2024"` →
  `venueName=Glastonbury`, `year=2024`) and groups returned setlists by
  `(venue, eventDate)`. Festival searches that rely on curated event names
  rather than venue names may return fewer results than expected.
- `lineup_complete` is always `false` — setlist.fm doesn't surface ground
  truth for full lineups; the UI treats every lineup as potentially partial
  and allows manual artist additions.
- Per-artist festival mode makes one Spotify request per artist with no
  parallelism; a 20-artist festival takes proportionally longer. Rate-limit
  handling (429 + `Retry-After`) is in place.
