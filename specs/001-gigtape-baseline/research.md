# Research: Gigtape — Baseline (Phase 1)

**Date**: 2026-04-21
**Branch**: `001-gigtape-baseline`
**Plan**: [plan.md](plan.md)

All NEEDS CLARIFICATION items from Technical Context are resolved below.

---

## 1. Go Workspace (go.work) for Monorepo

**Decision**: Go workspaces (`go.work`, Go 1.18+) with separate `go.mod` per package and app.

**Rationale**: Workspaces let local modules reference each other without publishing to a registry.
`go build ./...` and `go test ./...` work from the repo root. Module boundaries are enforced at
build time — if `packages/domain` accidentally imports an external package, `go mod tidy` on
that module makes it explicit.

**Alternatives considered**: Single root `go.mod` — simpler but allows accidental cross-layer
imports; no compile-time enforcement of hexagonal boundaries. Separate git repos — overkill for
a solo project.

**Setup**:
```
go work init
go work use ./packages/domain
go work use ./packages/usecases
go work use ./packages/adapters/setlistfm
go work use ./packages/adapters/spotify
go work use ./apps/api
go work use ./apps/cli
```

---

## 2. setlist.fm REST API

**Decision**: setlist.fm REST API v1 (`api.setlist.fm/rest/1.0`). Auth via `x-api-key` header.

**Rate limits**: Free tier — approximately 2 req/s. Client enforces a 500ms ticker between
requests. 429 responses trigger exponential backoff: 1s → 2s → 4s → 8s (max 3 retries).

**Relevant endpoints for Phase 1**:

| Endpoint | Purpose |
|---|---|
| `GET /search/artists?artistName={name}` | Artist disambiguation — returns candidates |
| `GET /artist/{mbid}/setlists?p=1` | Recent setlists for artist (sorted date desc) |
| `GET /search/setlists?eventName={name}` | Festival lineup search |

**Festival search**: setlist.fm supports event-level search via `eventName`. Results include per-artist setlists where available. Incomplete lineups are expected — the hybrid approach (auto-populate from setlist.fm, manual fill for gaps) directly maps to this API behavior.

**Attribution**: Every setlist response includes a URL. The adapter populates `Setlist.SourceAttribution` as `"setlist.fm • {url}"`. This string must be rendered wherever setlist data appears.

---

## 3. Spotify API — OAuth 2.0 and Track Operations

**Decision**: Authorization Code with PKCE for the web SPA (no client secret in the browser).
For the CLI: local redirect server on an ephemeral port (`http://localhost:{port}/callback`).

**Rationale**: PKCE is the correct flow for public clients (SPAs, CLIs) per OAuth 2.0 Security
Best Practices. The CLI pattern (ephemeral local server + open browser) is used by the GitHub
CLI, Spotify CLI, and the Go `golang.org/x/oauth2` package supports it natively.

**Required Spotify OAuth scopes**:

| Scope | Purpose |
|---|---|
| `playlist-modify-private` | Create private playlists |
| `playlist-read-private` | Check for existing playlist names to avoid collision |

**Key endpoints**:

| Endpoint | Purpose |
|---|---|
| `GET /v1/search?type=track&q=track:{title}+artist:{artist}` | Track search |
| `POST /v1/users/{user_id}/playlists` | Create playlist (body: `"public": false`) |
| `POST /v1/playlists/{id}/tracks` | Add tracks in batches of 100 (Spotify limit) |

**Track matching strategy**: Query `track:{title} artist:{artistName}`. Take first result.
If zero results returned, add title to `PlaylistResult.UnmatchedTracks`. No fuzzy matching
in Phase 1 — keep it simple, surface misses to the user.

**Rate limit handling**: Spotify returns 429 with `Retry-After` header (seconds). Adapter reads
this header and sleeps for the indicated duration before retrying. Exponential backoff (1s base)
used as fallback when `Retry-After` is absent.

---

## 4. In-Memory Session Store

**Decision**: `sync.Map` keyed by UUID v4 session ID. Sessions expire after 1 hour (configurable
via `SESSION_TTL_MINUTES` env var). Background goroutine cleans up expired sessions every 10
minutes.

**Session ID delivery**: After OAuth callback, the API returns the session ID in the response
body. The Vue SPA stores it in memory (not `localStorage` — stateless design) and sends it as
`X-Session-ID` on all subsequent requests. The CLI stores it in a temp file for the duration of
the interactive session.

**Rationale**: `sync.Map` is safe for concurrent reads/writes without additional locking. UUID v4
from `crypto/rand` provides sufficient uniqueness for 25 beta users. No external cache (Redis,
memcached) needed — that is a Phase 2 concern if user count grows.

**Session struct** (in `apps/api/middleware/session.go`, not in domain):
```go
type Session struct {
    ID        string
    Token     *oauth2.Token
    UserID    string
    CreatedAt time.Time
    ExpiresAt time.Time
}
```

---

## 5. CLI Library

**Decision**: `cobra` (`github.com/spf13/cobra`).

**Rationale**: Cobra is the de facto standard for Go CLIs (kubectl, Hugo, GitHub CLI). It provides
subcommand routing, flag parsing, help generation, and shell completion. The mental model maps
cleanly to Symfony Console (commands + flags + arguments) which the developer knows.

**CLI commands (Phase 1)**:

| Command | Description |
|---|---|
| `gigtape auth` | Initiate Spotify OAuth; opens browser; waits for callback |
| `gigtape artist [name]` | Single-artist flow: search → preview → create playlist |
| `gigtape festival [name]` | Festival flow: search → preview lineup → create playlist(s) |

**Bubbletea considered**: Better for interactive TUI with real-time UI updates. Not needed for
Phase 1 — cobra with `--flags` and basic `fmt.Println` output is sufficient and simpler.

---

## 6. Vue 3 SPA — Structure and Capacitor Compatibility

**Decision**: Vue 3 + Vite. Vue Router in history mode (Gin serves `index.html` as catch-all for
non-API routes). No SSR. No Pinia in Phase 1. Fetch-based API client in `src/api/client.ts`.

**Rationale**: History mode URLs (no `#`) are cleaner and Gin can serve the catch-all trivially.
Capacitor works with both hash and history mode — history mode is the better default.

**Capacitor readiness**: No Capacitor-specific code in Phase 1, but the SPA structure requires:
- No server-side rendering dependencies
- All API calls relative (`/api/...`, proxied by Vite in dev; served by Gin in prod)
- No `window.location` hard-coding; use `vue-router` for navigation

**State management**: Component-level `ref`/`reactive` only. No Pinia until Phase 2 introduces
enough shared state to justify it.

**Vite dev proxy**: `vite.config.ts` proxies `/api/*` to `http://localhost:8080` in development.
In production, Gin serves both the API and the built Vue assets from `apps/web/dist/`.

---

## 7. Structured Logging

**Decision**: `log/slog` (Go standard library, 1.21+). JSON handler in production, text handler
in local development. No third-party logging library.

**Log fields** (present on every structured log entry):

| Field | Type | Source |
|---|---|---|
| `session_id` | string | Middleware, injected into context |
| `use_case` | string | Use case name (e.g., `"CreatePlaylistFromArtist"`) |
| `adapter` | string | Adapter name when logging at adapter boundary |
| `artist` | string | Artist name being processed |
| `error` | string | Error message (on error entries only) |

**Logging rules**:
- Errors logged at use case boundary — not inside adapters (unless adapter catches and re-wraps)
- No logging inside tight loops (e.g., per-track search) — log at batch level
- `sentry.CaptureException(err)` called for unexpected errors (not "no setlist found" — that is a
  normal domain outcome, not an exception)

---

## 8. Testing Strategy

**Decision**: Standard Go `testing` package. `github.com/stretchr/testify` for assertions.
Table-driven tests for use cases. Hand-rolled fakes in `packages/usecases/fakes/`.

**Fake adapter pattern**:
```go
type FakeSetlistProvider struct {
    Artists  []domain.Artist
    Setlists []domain.Setlist
    Err      error
}

func (f *FakeSetlistProvider) SearchArtists(_ context.Context, _ string) ([]domain.Artist, error) {
    return f.Artists, f.Err
}
```

**No mocking framework**: Hand-rolled fakes are idiomatic Go, more transparent than `gomock`
or `testify/mock`, and appropriate for a solo developer who will read these tests six months later.

**Integration tests**: Guarded by build tag `//go:build integration`. Run separately with
`go test -tags integration ./...`. Require real API keys. Not run in normal `go test ./...`.

**HTTP handler tests**: `net/http/httptest` with fake use case implementations injected.
No live adapters needed to test handlers.
