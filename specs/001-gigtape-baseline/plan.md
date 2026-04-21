# Implementation Plan: Gigtape — Baseline (Phase 1)

**Branch**: `001-gigtape-baseline` | **Date**: 2026-04-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-gigtape-baseline/spec.md`

## Summary

Gigtape automates the creation of Spotify pre-show playlists from setlist.fm data. Given an
artist or festival name, the system fetches the relevant setlist(s), lets the user preview and
edit the track selection, and creates a private Spotify playlist with a direct link. The system
is built in Go using Hexagonal Architecture (Ports & Adapters) with a Gin REST API backend, a
Vue 3 SPA frontend, and a Go CLI as the first validated delivery surface.

## Technical Context

**Language/Version**: Go 1.22+ (backend, CLI, adapters); Node 20+ / Vue 3 (frontend)
**Primary Dependencies**: Gin (HTTP routing), cobra (CLI), Vue 3 + Vite (frontend), go.work (workspace)
**Storage**: N/A — in-memory session store only (`sync.Map` + TTL)
**Testing**: Go `testing` + `net/http/httptest`; testify for assertions; hand-rolled fakes via port interfaces
**Target Platform**: Local development only (macOS/Linux) — no deployment target in Phase 1
**Project Type**: REST API + CLI tool + SPA frontend (monorepo, 3 deployable components)
**Performance Goals**: Single-artist flow <2 min total; festival flow (10 artists) <5 min; individual setlist fetch <2s
**Constraints**: Stateless — no database; sessions discarded after use; ~25 beta users; setlist.fm attribution mandatory; Spotify Developer Policy compliance
**Scale/Scope**: Beta — up to 25 concurrent users; no horizontal scaling required in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Hexagonal Architecture**: `packages/domain` has zero external imports — it is the
  architectural center enforced at the compiler level. Port interfaces are defined in domain.
  Adapters import domain to implement ports; domain never imports adapters. Delivery apps
  (api, cli) import use cases and compose adapters at `main.go` only.
- [x] **II. Domain Integrity**: Domain entities (Artist, Setlist, Track, Event, Playlist) carry
  no HTTP clients, Spotify URIs, setlist.fm MBIDs as typed fields, or infrastructure concerns.
  `PlaylistResult` is a first-class structured partial-success type in domain — never collapsed
  into a binary error.
- [x] **III. Simplicity Mandate**: Phase 1 has exactly 3 deployable components: `apps/api`,
  `apps/cli`, `apps/web`. Phase 2/3 directories exist in the repo as empty placeholders only —
  no code, no interfaces shaped for future phases. Each dependency (Gin, cobra, go.work, Vue)
  is justified by a concrete Phase 1 requirement.
- [x] **IV. Stateless by Design**: No persistent storage. Session store is an in-memory
  `sync.Map` with TTL and a background cleanup goroutine. OAuth tokens are session-scoped and
  discarded on expiry or after playlist creation.
- [x] **V. Resilience**: setlistfm and spotify adapters implement exponential backoff on 429
  responses. Every use case returns a non-nil `PlaylistResult` — success, partial success, or
  explicit failure with a human-readable message. No errors swallowed.
- [x] **VI. Observability**: Structured JSON logging via `log/slog` (stdlib, Go 1.21+) with
  fields `session_id`, `use_case`, `adapter`, `artist`, `error`. Errors logged at use case
  boundary. Sentry SDK in `apps/api` and `apps/cli` for unexpected adapter failures.
- [x] **VII. Testability**: All use cases accept port interfaces — `SetlistProvider`,
  `EventProvider`, `PlaylistDestination` — and are tested with hand-rolled fakes in
  `packages/usecases/fakes/`. No test requires a live API call. `httptest` for API handlers.
- [x] **VIII. Phased Delivery**: Phase 1 is self-contained: domain + usecases + setlistfm +
  spotify + api + cli + web. Phase 2/3 adapters and delivery surfaces do not influence Phase 1
  design. `DiscoverUpcomingConcerts` use case is not implemented until Phase 2.
- [x] **IX. Code Quality**: All errors handled explicitly with `if err != nil`. No `_` discard
  on errors in production paths. Naming follows Go conventions. Code readable without this
  document as context.

All gates pass. No complexity justification required.

## Project Structure

### Documentation (this feature)

```text
specs/001-gigtape-baseline/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── api.md           # Phase 1 output — REST API contract
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
gigtape/
├── packages/
│   ├── domain/                      # entities + port interfaces (zero external imports)
│   │   ├── artist.go
│   │   ├── setlist.go
│   │   ├── track.go
│   │   ├── event.go
│   │   ├── playlist.go
│   │   ├── result.go                # PlaylistResult
│   │   ├── ports.go                 # SetlistProvider, EventProvider, PlaylistDestination
│   │   └── go.mod
│   ├── usecases/                    # use case implementations
│   │   ├── create_from_artist.go
│   │   ├── create_from_festival.go
│   │   ├── preview_setlist.go
│   │   ├── fakes/                   # hand-rolled fakes for testing
│   │   │   ├── fake_setlist_provider.go
│   │   │   └── fake_playlist_destination.go
│   │   └── go.mod
│   └── adapters/
│       ├── setlistfm/               # implements SetlistProvider + EventProvider
│       │   ├── client.go            # HTTP client with rate limit ticker
│       │   ├── setlist_provider.go
│       │   ├── event_provider.go
│       │   └── go.mod
│       ├── spotify/                 # implements PlaylistDestination
│       │   ├── auth.go              # OAuth 2.0 PKCE flow
│       │   ├── search.go            # track search (artist+title)
│       │   ├── playlist.go          # playlist creation + track batching
│       │   └── go.mod
│       ├── ticketmaster/            # Phase 2 — empty placeholder
│       └── applemusic/              # Phase 3 — empty placeholder
├── apps/
│   ├── api/                         # Gin REST API server
│   │   ├── main.go                  # composition root: wire adapters → use cases → handlers
│   │   ├── handlers/
│   │   │   ├── auth.go
│   │   │   ├── setlist.go
│   │   │   └── playlist.go
│   │   ├── middleware/
│   │   │   ├── session.go           # in-memory session store
│   │   │   └── logger.go
│   │   └── go.mod
│   ├── cli/                         # Go CLI (cobra)
│   │   ├── main.go
│   │   ├── cmd/
│   │   │   ├── root.go
│   │   │   ├── auth.go
│   │   │   ├── artist.go
│   │   │   └── festival.go
│   │   └── go.mod
│   ├── web/                         # Vue 3 + Vite SPA
│   │   ├── src/
│   │   │   ├── views/               # page-level components
│   │   │   ├── components/          # reusable UI components
│   │   │   └── api/                 # fetch wrapper + typed API client
│   │   │       └── client.ts
│   │   ├── package.json
│   │   └── vite.config.ts
│   ├── telegram/                    # Phase 2 — empty placeholder
│   └── discord/                     # Phase 2 — empty placeholder
├── go.work
└── Makefile
```

**Structure Decision**: Go workspace monorepo. `packages/domain` is the architectural center
with zero external imports — the Go compiler enforces this boundary. Use cases depend on domain
only. Adapters depend on domain to implement port interfaces. Delivery apps compose everything
at `main.go`. If a package in `packages/domain` ever imports an external package, that is an
immediate constitution violation detectable by `go list -deps`.

## Complexity Tracking

No violations. All constitution gates pass for Phase 1 scope.
