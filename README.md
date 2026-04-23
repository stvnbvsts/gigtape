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
| 3 | **User Story 1 — Single-artist playlist (MVP)** | ✅ Done |
| 4 | User Story 2 — Festival playlist | ⏳ Not started |
| 5 | Polish (Sentry, rate limit, audit) | ⏳ Not started |

What you can do today:

- Authenticate with Spotify (PKCE OAuth)
- Search for an artist by name
- Fetch recent setlists for that artist from setlist.fm
- Preview, edit (remove / manually add) the track list
- Create a private Spotify playlist and receive a direct link
- See which tracks Spotify couldn't match (never silently dropped)

What is **not** available yet: festival mode (`gigtape festival`), rate limiting
middleware, Sentry integration.

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
SENTRY_DSN=
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

## How to test Phase 3 — single-artist playlist

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

## Endpoints available in Phase 3

| Method | Path | Purpose |
|---|---|---|
| GET | `/auth/login` | Start PKCE OAuth; returns `{ auth_url }` |
| GET | `/auth/callback` | OAuth callback; returns `{ session_id }` |
| GET | `/artists/search?q=` | Disambiguation candidates |
| GET | `/setlists?artist_ref=&artist_name=` | Recent setlists (desc by date) |
| POST | `/playlists/artist` | Create playlist from artist's setlist |

Festival routes (`/events/search`, `/playlists/festival`) are **not** wired
yet — Phase 4.

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
| 429 from setlist.fm / Spotify | Rate limit | Backoff is implemented; just retry after a few seconds |
| Setlist has 0 songs / "No setlists found" | Artist has no recent shows on setlist.fm | Try a more active artist; manual track entry works in the web UI |

## Next

- **Phase 4** (US2): `gigtape festival`, `/events/search`,
  `/playlists/festival`, merged + per-artist modes.
- **Phase 5**: Sentry, per-session rate limit, error-path audit, full
  `quickstart.md` validation run.
