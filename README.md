# Gigtape

Gigtape creates private Spotify playlists from setlist.fm data.

Search for an artist or festival, preview the songs from recent setlists,
remove or add tracks, and create a private Spotify playlist with a direct link.
It is useful before a concert, festival, or tour stop when you want a playlist
that sounds like the show you are about to see.

## Features

- Spotify PKCE OAuth for web, API, and CLI flows
- Artist search with disambiguation
- Recent setlist preview from setlist.fm
- Track editing before playlist creation
- Festival playlist mode with artist deselection and manual artist additions
- Merged festival playlists or one playlist per artist
- Explicit reporting for tracks Spotify could not match
- Explicit reporting for skipped festival artists
- setlist.fm attribution wherever setlist data is shown
- Per-session API rate limiting
- Optional Sentry reporting for API and CLI errors

## Architecture

Gigtape is a Go workspace monorepo with three user-facing surfaces:

```text
apps/
  api/                 Gin REST API
  cli/                 Cobra CLI
  web/                 Vue 3 + Vite SPA
packages/
  domain/              stdlib-only entities and port interfaces
  usecases/            application use cases
  adapters/
    setlistfm/         setlist.fm client and providers
    spotify/           Spotify OAuth, search, and playlist creation
```

The code follows a ports-and-adapters shape. `packages/domain` is the center:
it defines entities and interfaces without importing framework, HTTP, Spotify,
or setlist.fm code. Delivery apps wire the use cases to adapters at their
composition roots.

## Requirements

| Requirement | Version | Notes |
|---|---:|---|
| Docker | 24+ with Compose v2 | Recommended for API + web |
| Go | 1.22+ | Required for native API or CLI |
| Node | 20+ | Required for native web development |
| setlist.fm API key | - | Create one in your setlist.fm account settings |
| Spotify Developer app | - | Create one in the Spotify Developer Dashboard |

In your Spotify Developer Dashboard, add these redirect URIs:

- `http://127.0.0.1:8080/auth/callback` for the web and API flow
- `http://127.0.0.1/callback` for the native CLI flow; Spotify allows loopback IP redirects to add a dynamic port at authorization time

The CLI chooses an available local port when you run `gigtape auth`; add the
exact URI it prints if Spotify rejects the callback.

## Environment

Copy the example file and fill in your credentials:

```bash
cp .env.example .env
```

Required and common values:

```env
SETLISTFM_API_KEY=your_setlistfm_key_here
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
SPOTIFY_REDIRECT_URI=http://127.0.0.1:8080/auth/callback
SESSION_TTL_MINUTES=60
WEB_REDIRECT_URL=http://localhost:5173/
SENTRY_DSN=
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=gigtape@dev
LOG_FORMAT=text
```

`SPOTIFY_CLIENT_SECRET` is currently kept for future compatibility; the app
uses PKCE and does not require the secret for the implemented OAuth flow.

For native commands, load the environment first:

```bash
set -a && source .env && set +a
```

## Run With Docker

Docker runs the API and web UI. The CLI intentionally remains native because
its OAuth flow opens your browser and listens on an ephemeral loopback port.

```bash
docker compose up --build
```

Then open:

- Web UI: `http://localhost:5173`
- API: `http://localhost:8080`

The web container serves the built SPA with nginx and proxies `/api/*` to the
API service on the internal Compose network. Browser redirects still use
`localhost`, so the Spotify dashboard redirect URI remains
`http://127.0.0.1:8080/auth/callback`.

For hosted deployments, set the web service's `API_BASE_URL` to the full API
origin, including the scheme, for example:

```bash
API_BASE_URL=https://your-api-service.up.railway.app
```

This keeps browser calls same-origin through `/api/*` and avoids mobile Safari
following cross-origin HTTP-to-HTTPS redirects.

Useful Docker commands:

```bash
make docker-up
make docker-build
make docker-down
```

## Run Natively

Initialize the Go workspace once after cloning:

```bash
go work sync
```

Start the API:

```bash
go run ./apps/api
# Listening on :8080
```

Start the web UI in another terminal:

```bash
cd apps/web
npm install
npm run dev
# Vite dev server on http://localhost:5173
```

The native Vite dev server proxies `/api/*` to `http://localhost:8080` by
default. To point it somewhere else, create `apps/web/.env.local`:

```env
VITE_API_PROXY_TARGET=http://localhost:8080
```

Build and run the CLI natively:

```bash
go build -o gigtape ./apps/cli
./gigtape --help
```

Common Make targets:

```bash
make build-api
make build-cli
make serve-web
make test
```

## Web Usage

1. Open `http://localhost:5173`.
2. Click **Connect Spotify** and complete consent in the browser.
3. Search for an artist.
4. Pick the correct artist result.
5. Preview the setlist, remove songs, or add manual tracks.
6. Create the private Spotify playlist.
7. Open the returned Spotify playlist link.

Festival mode is available from the web UI. Search for a festival or venue,
review the lineup, deselect artists, add manual artists if needed, and choose
either a merged playlist or one playlist per artist.

## CLI Usage

Authenticate first:

```bash
./gigtape auth
```

Create a playlist from an artist's latest setlist:

```bash
./gigtape artist "Radiohead"
```

Create a festival playlist:

```bash
./gigtape festival "Glastonbury 2024"
```

CLI OAuth tokens are cached at `$TMPDIR/gigtape-token.json` until they expire.

## API Usage

Start OAuth:

```bash
curl -s http://localhost:8080/auth/login | jq
```

Open the returned `auth_url`, complete Spotify consent, and copy the returned
`session_id` if the callback is configured to return JSON. Protected API routes
require this header:

```text
X-Session-ID: <session-id>
```

Available endpoints:

| Method | Path | Purpose |
|---|---|---|
| GET | `/auth/login` | Start Spotify PKCE OAuth |
| GET | `/auth/callback` | Spotify OAuth callback |
| GET | `/artists/search?q=` | Search artist candidates |
| GET | `/setlists?artist_ref=&artist_name=` | Fetch recent setlists |
| POST | `/playlists/artist` | Create an artist playlist |
| GET | `/events/search?q=` | Search festival/event candidates |
| POST | `/playlists/festival` | Create merged or per-artist festival playlists |

Example artist search:

```bash
curl -s "http://localhost:8080/artists/search?q=Radiohead" \
  -H "X-Session-ID: $SESSION_ID" | jq
```

## Testing

Run the Go test suite:

```bash
make test
```

Build the web app:

```bash
cd apps/web
npm ci
npm run build
```

Check the domain package stays framework-free:

```bash
( cd packages/domain && go list -f '{{ join .Imports "\n" }}' ./... | sort -u )
# expected: context, fmt, time
```

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `401 session_not_found` | Missing or unknown `X-Session-ID` | Re-run the Spotify login flow |
| `401 session_expired` | Session TTL expired | Re-authenticate |
| `429 rate_limited` | Per-session API rate limit | Wait and retry; honor `Retry-After` |
| Spotify redirect error | Redirect URI missing in Spotify app | Add the exact callback URI to the Spotify dashboard |
| CLI says `not authenticated` | Token file missing or expired | Run `./gigtape auth` |
| Playlist not created | Missing Spotify scope or auth issue | Re-authenticate and verify app scopes |
| No setlists found | Artist or event has no setlist.fm data | Try another artist/event or add tracks manually in the web UI |
| Some tracks are missing | Spotify search could not match them | Review the unmatched tracks list |

## Known Caveats

- setlist.fm does not expose authoritative complete festival lineups, so
  festival search results should be treated as potentially partial.
- Festival search relies on setlist.fm setlist and venue data. Queries that
  depend on curated festival names may return fewer results than expected.
- Per-artist festival mode creates playlists sequentially, so large lineups can
  take a while.
- Sentry is disabled unless `SENTRY_DSN` is set.

## Roadmap

Phase 1 is the current baseline: Spotify playlist creation from setlist.fm
artist and festival-style setlist data, delivered through the web app, REST
API, and CLI. It is intentionally stateless, with in-memory sessions and no
stored accounts, playlist history, or preferences.

Phase 2 is not implemented yet. The planned direction is:

- Ticketmaster-backed event discovery so users can search upcoming concerts
  and festivals from a more structured event source.
- A `DiscoverUpcomingConcerts` use case that can sit beside the current
  setlist and playlist use cases.
- Additional delivery surfaces such as Telegram and Discord bots.
- Stronger state management or caching if the beta user count grows beyond the
  current local/in-memory design.
- More complete automated coverage around interactive CLI and full API flows.

Apple Music support is considered later-phase work and is not part of the
current Phase 2 scope.
