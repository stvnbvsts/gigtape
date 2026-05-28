# Tasks: Gigtape — Baseline (Phase 1)

**Input**: Design documents from `specs/001-gigtape-baseline/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api.md, research.md, quickstart.md
**Branch**: `001-gigtape-baseline`

**Organization**: Tasks are grouped by user story to enable independent implementation
and testing of each story. No test tasks are included (not requested in spec).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared dependencies)
- **[Story]**: Which user story this task belongs to ([US1], [US2])
- Exact file paths are included in all descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Repo skeleton, module initialization, and tooling — must be done before any code.

- [x] T001 Create full directory skeleton at repo root: `packages/domain/`, `packages/usecases/fakes/`, `packages/adapters/setlistfm/`, `packages/adapters/spotify/`, `packages/adapters/ticketmaster/`, `packages/adapters/applemusic/`, `apps/api/handlers/`, `apps/api/middleware/`, `apps/cli/cmd/`, `apps/web/src/views/`, `apps/web/src/components/`, `apps/web/src/api/`
- [x] T002 [P] Create `packages/domain/go.mod` with module name `gigtape/domain`, go 1.22, and zero external dependencies
- [x] T003 [P] Create `packages/usecases/go.mod` with module name `gigtape/usecases`, go 1.22, and `require gigtape/domain`
- [x] T004 [P] Create `packages/adapters/setlistfm/go.mod` with module name `gigtape/adapters/setlistfm`, go 1.22, require `gigtape/domain` and `github.com/stretchr/testify`
- [x] T005 [P] Create `packages/adapters/spotify/go.mod` with module name `gigtape/adapters/spotify`, go 1.22, require `gigtape/domain` and `golang.org/x/oauth2`
- [x] T006 [P] Create `apps/api/go.mod` with module name `gigtape/api`, go 1.22, require `github.com/gin-gonic/gin`, `gigtape/usecases`, `gigtape/adapters/setlistfm`, `gigtape/adapters/spotify`, `golang.org/x/oauth2`, `github.com/getsentry/sentry-go`
- [x] T007 [P] Create `apps/cli/go.mod` with module name `gigtape/cli`, go 1.22, require `github.com/spf13/cobra`, `gigtape/usecases`, `gigtape/adapters/setlistfm`, `gigtape/adapters/spotify`, `golang.org/x/oauth2`
- [x] T008 Create `go.work` at repo root and run `go work use` for all 6 modules: `packages/domain`, `packages/usecases`, `packages/adapters/setlistfm`, `packages/adapters/spotify`, `apps/api`, `apps/cli` (depends on T002–T007)
- [x] T009 [P] Create `apps/web/package.json` with Vue 3 and Vite dependencies and `apps/web/vite.config.ts` with Vue plugin and `/api/*` proxy pointing to `http://localhost:8080`
- [x] T010 [P] Create `Makefile` at repo root with targets: `build-api` (go build ./apps/api), `build-cli` (go build -o gigtape ./apps/cli), `serve-web` (cd apps/web && npm run dev), `test` (go test ./packages/...), `test-integration` (RUN_INTEGRATION=true go test -tags integration ./packages/adapters/...)
- [x] T011 [P] Create `.env.example` at repo root with all required keys: `SETLISTFM_API_KEY`, `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`, `SPOTIFY_REDIRECT_URI`, `SESSION_TTL_MINUTES`, `SENTRY_DSN`, `LOG_FORMAT`

**Checkpoint**: Run `go work sync` from repo root — should succeed with zero errors

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain package, session infrastructure, Spotify OAuth, and API routing shell.
Both user stories depend on these before any story-specific code can begin.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T012 [P] Create `packages/domain/artist.go` with `Artist` struct: `Name string`, `Disambiguation string`, `ExternalRef string`; add package-level doc comment stating zero external imports rule
- [x] T013 [P] Create `packages/domain/track.go` with `Track` struct: `Title string`, `ArtistName string`; add comment explaining ArtistName is denormalized for Spotify search
- [x] T014 [P] Create `packages/domain/setlist.go` with `Setlist` struct: `Artist Artist`, `EventName string`, `Date time.Time`, `Tracks []Track`, `SourceAttribution string`
- [x] T015 [P] Create `packages/domain/event.go` with `Event` struct: `Name string`, `Date time.Time`, `Location string`, `Artists []Artist`; add comment that Artists slice order is preserved for merged playlist grouping
- [x] T016 [P] Create `packages/domain/playlist.go` with `Playlist` struct: `Name string`, `Tracks []Track`, `CreatedAt time.Time`; add naming convention constants (single artist: `"{ArtistName} — {YYYY-MM-DD}"`, merged: `"{FestivalName} — {YYYY-MM-DD}"`, per-artist festival: `"{ArtistName} — {FestivalName} — {YYYY-MM-DD}"`)
- [x] T017 [P] Create `packages/domain/result.go` with `PlaylistResult` struct: `PlaylistURL string`, `MatchedTracks []Track`, `UnmatchedTracks []string`, `SkippedArtists []string`
- [x] T018 Create `packages/domain/ports.go` defining `SetlistProvider` interface (`SearchArtists`, `GetSetlists`), `EventProvider` interface (`SearchEvents`), and `PlaylistDestination` interface (`CreatePlaylist`) — all methods take `context.Context` as first arg (depends on T012–T017)
- [x] T019 [P] Create `apps/api/middleware/session.go` with `Session` struct (`ID string`, `Token *oauth2.Token`, `UserID string`, `CreatedAt time.Time`, `ExpiresAt time.Time`), `sync.Map` session store, `NewSession(token, userID) Session` constructor using `crypto/rand` UUID v4, `GetSession(id) (Session, bool)`, background cleanup goroutine that purges expired sessions every 10 minutes, `SESSION_TTL_MINUTES` read from env
- [x] T020 [P] Create `apps/api/middleware/logger.go` with Gin middleware that creates an `slog.Logger` with `session_id` field (read from `X-Session-ID` header), stores logger in Gin context, and logs request completion with method, path, status, and latency; uses JSON handler when `LOG_FORMAT=json`, text handler otherwise
- [x] T021 Create `packages/adapters/setlistfm/client.go` with `Client` struct holding API key and rate-limit ticker (500ms between requests); `do(ctx, method, path, params)` method that respects the ticker, reads response, and implements exponential backoff (1s→2s→4s→8s, max 3 retries) on HTTP 429 responses (depends on T018 being resolvable via workspace)
- [x] T022 Create `packages/adapters/spotify/auth.go` with PKCE OAuth 2.0 flow: `GenerateChallenge() (verifier, challenge string)`, `AuthURL(challenge, state string) string` building Spotify authorize URL with scopes `playlist-modify-private playlist-read-private`, `ExchangeCode(ctx, code, verifier string) (*oauth2.Token, error)`, `NewClient(ctx, token) *http.Client` (depends on T005 go.mod with oauth2)
- [x] T023 Create `apps/api/handlers/auth.go` implementing `GET /auth/login` (generate PKCE challenge + state, store state in session store pending map, return `{"auth_url": "..."}`) and `GET /auth/callback` (validate state param, exchange code for token, create session, return `{"session_id": "uuid"}`) (depends on T019, T022)
- [x] T024 Create `apps/api/main.go` as composition root: load env vars, initialize Sentry, create Gin router, register `middleware/logger` and `middleware/session` authentication middleware (validate `X-Session-ID` on all non-auth routes, return 401 `session_not_found` or `session_expired`), mount auth handler routes at `/auth/*`, start server on `:8080` (depends on T019, T020, T023)

**Checkpoint**: `go build ./apps/api` compiles; `GET /auth/login` returns a valid Spotify auth URL

---

## Phase 3: User Story 1 — Single Artist Playlist (Priority: P1) 🎯 MVP

**Goal**: A user can connect Spotify, search for an artist, preview and edit the setlist,
create a private playlist, and receive a direct Spotify link — end-to-end.

**Independent Test**: Run `gigtape auth` → `gigtape artist "Radiohead"` → confirm playlist
appears in Spotify account with expected songs and no silent failures.
Alternatively: complete OAuth in browser → `GET /artists/search?q=Radiohead` →
`GET /setlists?artist_ref={ref}` → `POST /playlists/artist` → verify playlist URL.

### Implementation for User Story 1

- [X] T025 [P] [US1] Create `packages/usecases/fakes/fake_setlist_provider.go` implementing `domain.SetlistProvider` with configurable `Artists []domain.Artist`, `Setlists []domain.Setlist`, `Err error` fields; used in use case unit tests
- [X] T026 [P] [US1] Create `packages/usecases/fakes/fake_playlist_destination.go` implementing `domain.PlaylistDestination` with configurable `Result domain.PlaylistResult`, `Err error`, and `Captured *domain.Playlist` fields to inspect what was passed
- [X] T026b [P] [US2] Create `packages/usecases/fakes/fake_event_provider.go` implementing `domain.EventProvider` with configurable `Events []domain.Event`, `Err error` fields; required for unit testing `apps/api/handlers/events.go` without a live setlist.fm adapter (constitution Principle VII)
- [X] T027 [US1] Create `packages/adapters/setlistfm/setlist_provider.go` implementing `domain.SetlistProvider`: `SearchArtists` calls `GET /search/artists?artistName={name}` and maps response to `[]domain.Artist` (ExternalRef = MBID); `GetSetlists` calls `GET /artist/{mbid}/setlists?p=1`, maps response to `[]domain.Setlist` sorted date descending, populates `SourceAttribution` as `"setlist.fm • {url}"` (depends on T021, T018)
- [X] T028 [US1] Create `packages/adapters/spotify/search.go` with `SearchTrack(ctx, track domain.Track, client *http.Client) (string, bool)` that queries `GET /v1/search?type=track&q=track:{title}+artist:{artistName}`, returns the first result's Spotify URI or `("", false)` if no match (depends on T022, T018)
- [X] T029 [US1] Create `packages/adapters/spotify/playlist.go` implementing `domain.PlaylistDestination`: `CreatePlaylist` POSTs to `/v1/users/{userID}/playlists` with `"public": false`, then adds tracks in batches of 100 via `POST /v1/playlists/{id}/tracks`; uses `SearchTrack` per track; reads `Retry-After` header on 429 and sleeps accordingly; returns `domain.PlaylistResult` with matched, unmatched, and playlist URL (depends on T022, T028)
- [X] T030 [US1] Create `packages/usecases/preview_setlist.go` with `PreviewSetlist` struct taking `SetlistProvider domain.SetlistProvider`; `SearchArtists(ctx, name)` delegates to provider; `GetSetlists(ctx, artist)` delegates to provider, returns empty slice (not error) if none found, and a warning flag if `len(setlist.Tracks) < 6` for the most recent setlist (depends on T018, T025)
- [X] T031 [US1] Create `packages/usecases/create_from_artist.go` with `CreatePlaylistFromArtist` struct taking `Destination domain.PlaylistDestination`; `Execute(ctx, artistName string, date time.Time, tracks []domain.Track) (domain.PlaylistResult, error)` builds `domain.Playlist` with name `"{artistName} — {YYYY-MM-DD}"`, calls `Destination.CreatePlaylist`, returns non-nil `PlaylistResult` always (depends on T018, T025, T026)
- [X] T032 [US1] Create `apps/api/handlers/setlist.go` with Gin handlers for `GET /artists/search?q={name}` (call `PreviewSetlist.SearchArtists`, return `{"artists": [...]}`) and `GET /setlists?artist_ref={ref}` (resolve artist from ref, call `PreviewSetlist.GetSetlists`, return `{"setlists": [...]}` with `source_attribution` and `track_count`; return `{"setlists": []}` when empty); inject `PreviewSetlist` use case via closure (depends on T030, T019, T018)
- [X] T033 [US1] Create `apps/api/handlers/playlist.go` with Gin handler for `POST /playlists/artist` (decode request body into `artist_ref`, `artist_name`, `setlist_index`, `tracks []Track`; call `CreatePlaylistFromArtist.Execute`; return `{playlist_url, matched_tracks, unmatched_tracks, skipped_artists}`; return 401 for session errors, 502 for upstream errors) (depends on T031, T019, T018)
- [X] T034 [US1] Wire US1 adapters and handlers into `apps/api/main.go`: instantiate `setlistfm.Client`, `setlistfm.SetlistProvider`, `spotify.PlaylistDestination`, `PreviewSetlist` use case, `CreatePlaylistFromArtist` use case; register routes `GET /artists/search`, `GET /setlists`, `POST /playlists/artist` (depends on T024, T027, T029, T030, T031, T032, T033)
- [X] T035 [P] [US1] Create `apps/cli/cmd/root.go` with Cobra root command, persistent `--env-file` flag to load `.env`, and `Execute()` entry point
- [X] T036 [P] [US1] Create `apps/cli/cmd/auth.go` with `gigtape auth` Cobra command: start ephemeral HTTP server on a random local port to receive OAuth callback, open browser to Spotify authorize URL, wait for callback, exchange code for token, store token in temp file at `os.TempDir()/gigtape-token.json`, print `"Authenticated as {display_name}"` (depends on T022, T035)
- [X] T037 [US1] Create `apps/cli/cmd/artist.go` with `gigtape artist [name]` Cobra command: load token from temp file, call `PreviewSetlist.SearchArtists` and print disambiguation list with `[y/n]` prompt, call `GetSetlists`, print setlist preview with song count and date, **print `setlist.SourceAttribution` immediately after the setlist preview** (FR-013), prompt user to remove songs interactively, call `CreatePlaylistFromArtist.Execute`, print playlist URL, print unmatched tracks explicitly (depends on T030, T031, T035, T036)
- [X] T038 [US1] Create `apps/cli/main.go` as composition root: load env, instantiate `setlistfm.Client`, `setlistfm.SetlistProvider`, `spotify.PlaylistDestination`, inject into use cases, register commands, call `root.Execute()` (depends on T035, T036, T037)
- [X] T039 [P] [US1] Create `apps/web/src/api/client.ts` with typed fetch wrapper that reads session ID from module state and sends `X-Session-ID` header; export `searchArtists(q: string)`, `getSetlists(artistRef: string)`, `createArtistPlaylist(body)` typed against the API contract in `contracts/api.md`
- [X] T040 [P] [US1] Create `apps/web/src/views/ArtistSearchView.vue` with search input, "Connect Spotify" button that calls `GET /auth/login` and redirects, and artist disambiguation list that shows name + disambiguation text for each result
- [X] T041 [P] [US1] Create `apps/web/src/views/SetlistPreviewView.vue` showing setlist date and event name, source attribution text (required), track list with remove buttons, setlist selector when multiple setlists available, warning banner when track count < 6, and "Create Playlist" confirm button; when `getSetlists` returns an empty array, render an empty-state panel with a manual track entry input (title + artist name fields, add/remove rows) so the user can proceed without a setlist — this satisfies FR-009 for the single-artist web flow
- [X] T042 [P] [US1] Create `apps/web/src/views/PlaylistResultView.vue` showing playlist link (open in Spotify), matched track count, and explicit list of unmatched track titles under a "Not found on Spotify" heading (never hidden or collapsed)
- [X] T043 [US1] Create `apps/web/src/components/TrackList.vue` reusable component accepting `tracks` prop and emitting `remove(index)` event; used by `SetlistPreviewView` and later `FestivalModeView` (depends on T041)
- [X] T044 [US1] Create `apps/web/src/main.ts` with Vue Router in history mode: route `/` → `ArtistSearchView`, route `/setlist` → `SetlistPreviewView`, route `/result` → `PlaylistResultView`; store session ID returned by `/auth/callback` in module-level variable used by `client.ts` (depends on T040, T041, T042)

**Checkpoint**: `gigtape auth` + `gigtape artist "Radiohead"` produces a Spotify playlist.
`go test ./packages/domain/... ./packages/usecases/...` passes with zero failures.

---

## Phase 4: User Story 2 — Festival Playlist (Priority: P2)

**Goal**: A user can search for a festival, review the full lineup with per-artist setlist
data, deselect artists, choose merged or per-artist playlist mode, and receive Spotify
playlist links for all created playlists.

**Independent Test**: Search for "Glastonbury 2024" → verify lineup with artist names and
song counts → deselect one artist → choose merged mode → confirm one Spotify playlist
contains songs from remaining artists in lineup order with skipped artists listed.

### Implementation for User Story 2

- [X] T045 [P] [US2] Create `packages/adapters/setlistfm/event_provider.go` implementing `domain.EventProvider`: `SearchEvents` calls `GET /search/setlists?eventName={name}`, groups results by event, maps to `[]domain.Event` preserving artist lineup order, sets `lineup_complete: false` when not all artists have setlists (depends on T021, T018)
- [X] T046 [US2] Create `packages/usecases/create_from_festival.go` with `CreatePlaylistFromFestival` struct taking `Destination domain.PlaylistDestination`; `Execute(ctx, req FestivalRequest) ([]domain.PlaylistResult, error)` where `FestivalRequest` contains `EventName`, `EventDate`, `Mode` ("merged"/"per_artist"), and `[]ArtistEntry{ArtistRef, ArtistName, Include bool, Tracks []Track}`; in `merged` mode: flatten included artists' tracks in lineup order into one playlist; in `per_artist` mode: one `CreatePlaylist` call per included artist; populate `SkippedArtists` for excluded or empty-track artists; always return non-nil results slice (depends on T018, T025, T026, T031)
- [X] T047 [US2] Create `apps/api/handlers/events.go` with Gin handler for `GET /events/search?q={name}`: call `EventProvider.SearchEvents`, return `{"events": [{name, date, location, artists, lineup_complete}]}`; return `{"events": []}` when no matches (depends on T019, T045, T018)
- [X] T048 [US2] Extend `apps/api/handlers/playlist.go` to add `POST /playlists/festival` handler: decode request body into `FestivalRequest`; call `CreatePlaylistFromFestival.Execute`; return `{"results": [...]}` with HTTP 200 for merged or all-succeeded per_artist, HTTP 207 for per_artist when at least one succeeded and at least one failed (depends on T046, T033)
- [X] T049 [US2] Wire US2 adapters and handlers into `apps/api/main.go`: instantiate `setlistfm.EventProvider`, `CreatePlaylistFromFestival` use case; register routes `GET /events/search` and `POST /playlists/festival` (depends on T034, T045, T046, T047, T048)
- [X] T050 [US2] Create `apps/cli/cmd/festival.go` with `gigtape festival [name]` Cobra command: call `EventProvider.SearchEvents`, print lineup as numbered list with song counts, prompt user to enter comma-separated numbers to deselect, prompt `"Mode [merged/per-artist]:"`, call `CreatePlaylistFromFestival.Execute`, print all playlist URLs, print skipped artists, print unmatched tracks per artist; **print each artist's `setlist.SourceAttribution` when displaying per-artist setlist data** (FR-013) (depends on T038, T045, T046)
- [X] T050b [US2] Update `apps/cli/main.go` to wire US2: instantiate `setlistfm.EventProvider`, inject into `CreatePlaylistFromFestival` use case, and register the festival command; mirrors T049 for the CLI surface (depends on T038, T045, T046, T050)
- [X] T051 [P] [US2] Extend `apps/web/src/api/client.ts` to add `searchEvents(q: string)` and `createFestivalPlaylist(body)` typed methods
- [X] T052 [P] [US2] Create `apps/web/src/views/FestivalSearchView.vue` with festival name search input; on results, display lineup list showing artist name + song count per artist; each artist row has a checkbox to include/exclude; show "no setlist found" indicator for artists without setlist data; include "Add artist manually" input at bottom of lineup
- [X] T053 [P] [US2] Create `apps/web/src/views/FestivalModeView.vue` with two clearly labelled options — "One merged playlist" and "One playlist per artist" — and a confirm button; pass selected artists and mode to `createFestivalPlaylist`
- [X] T054 [US2] Create `apps/web/src/views/FestivalResultView.vue` showing one result card per playlist (name, link, matched count, unmatched list); show a "Skipped artists" section when `skipped_artists` is non-empty (depends on T042 for shared result display patterns)
- [X] T055 [US2] Add festival flow routes to `apps/web/src/main.ts`: `/festival` → `FestivalSearchView`, `/festival/mode` → `FestivalModeView`, `/festival/result` → `FestivalResultView` (depends on T044, T052, T053, T054)

**Checkpoint**: `gigtape festival "Glastonbury"` produces merged playlist with songs from
multiple artists in lineup order. `go test ./packages/usecases/...` passes for festival use case.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Observability, compliance, and end-to-end validation.

- [X] T056 [P] Add Sentry SDK initialization to `apps/api/main.go` (read `SENTRY_DSN` env, call `sentry.Init`) and `apps/cli/main.go`; wrap unexpected adapter errors with `sentry.CaptureException(err)` at use case boundaries in `create_from_artist.go` and `create_from_festival.go`
- [X] T057 [P] Create `apps/api/middleware/ratelimit.go` with per-session request throttle using a `sync.Map` of token buckets; apply as Gin middleware before handler routes to protect upstream API quota at ~2 req/s per session
- [X] T058 [P] Audit all `if err != nil` paths in `packages/usecases/`, `packages/adapters/setlistfm/`, and `packages/adapters/spotify/` — confirm no error is discarded with `_`; add `slog.Error` log entries at use case boundaries with `session_id`, `use_case`, `artist`, and `error` fields
- [X] T059 [P] Audit all surfaces where setlist data appears and confirm `SourceAttribution` is rendered: (1) all views in `apps/web/src/views/` — add attribution display to any missing it; (2) `apps/cli/cmd/artist.go` (T037) and `apps/cli/cmd/festival.go` (T050) — verify attribution is printed after each setlist preview; setlist.fm policy: attribution MUST appear wherever setlist data is shown
- [X] T060 Run the complete `quickstart.md` validation checklist: `go test ./packages/domain/... ./packages/usecases/... ./packages/adapters/... ./apps/api/...`, CLI smoke tests (`gigtape auth`, `gigtape artist "Radiohead"`, `gigtape festival "Glastonbury"`), web UI smoke tests in browser; record and fix any failures

**Checkpoint**: All validation checklist items in `quickstart.md` pass. Phase 1 complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Requires T001–T011 complete — **blocks all user stories**
- **User Story 1 (Phase 3)**: Requires Phase 2 complete — no dependency on US2
- **User Story 2 (Phase 4)**: Requires Phase 2 complete — builds on US1 adapters (T027, T028, T029) and reuses `PlaylistDestination`
- **Polish (Phase 5)**: Requires all desired user stories complete

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2 — no dependency on US2
- **US2 (P2)**: Requires Phase 2 and T027 (`setlistfm.SetlistProvider`), T028, T029 from US1 — the setlistfm client and spotify playlist adapter are shared

### Within Each User Story

- Domain fakes (T025, T026) before use cases (T030, T031, T046)
- Use cases before API handlers (T032, T033, T047, T048)
- API handlers before `main.go` wiring (T034, T049)
- `cmd/root.go` (T035) before any CLI commands (T036, T037, T050)
- Vue views before router registration (T044, T055)

### Parallel Opportunities

- T002–T007: All `go.mod` files in parallel
- T009–T011: Web setup, Makefile, `.env.example` in parallel
- T012–T017: All domain entity files in parallel
- T019–T020: Session middleware and logger middleware in parallel
- T025–T026: Both fakes in parallel
- T027–T029: setlistfm provider and spotify adapters in parallel (both depend only on T021/T022)
- T035–T036: CLI root and auth command in parallel
- T039–T042: All Vue views and API client for US1 in parallel
- T051–T053: Festival API client extension and festival views in parallel

---

## Parallel Example: User Story 1

```bash
# All domain entities (Phase 2 start):
Task: "Create packages/domain/artist.go"       # T012
Task: "Create packages/domain/track.go"        # T013
Task: "Create packages/domain/setlist.go"      # T014
Task: "Create packages/domain/event.go"        # T015
Task: "Create packages/domain/playlist.go"     # T016
Task: "Create packages/domain/result.go"       # T017

# All adapters after T021 (setlistfm client) and T022 (spotify auth):
Task: "Create setlistfm/setlist_provider.go"   # T027
Task: "Create spotify/search.go"               # T028
Task: "Create spotify/playlist.go"             # T029

# All Vue views for US1:
Task: "Create ArtistSearchView.vue"            # T040
Task: "Create SetlistPreviewView.vue"          # T041
Task: "Create PlaylistResultView.vue"          # T042
Task: "Extend api/client.ts with US1 methods"  # T039
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T011)
2. Complete Phase 2: Foundational (T012–T024) — **this is the blocker**
3. Complete Phase 3: User Story 1 (T025–T044)
4. **STOP and VALIDATE**: `gigtape auth` + `gigtape artist "Radiohead"` + browser OAuth flow
5. Share with beta users for early feedback before building US2

### Incremental Delivery

1. Setup + Foundational → workspace compiles, API server starts, auth flow works
2. Add US1 → single artist flow end-to-end → CLI + web both functional → **MVP**
3. Add US2 → festival flow → CLI + web both functional
4. Polish → observability, compliance, full quickstart checklist

### Solo Developer Strategy

Phase 2 must complete before anything else. Then tackle US1 fully before starting US2 —
the setlistfm and spotify adapters built in US1 (T027–T029) are reused in US2, so US1
completion unblocks US2 immediately.

---

## Notes

- [P] tasks operate on different files with no shared state — safe to run in parallel
- [US1]/[US2] labels trace every task back to its user story for independent delivery
- Domain package (`packages/domain`) must never have external imports — verify with `go list -deps ./packages/domain/...`
- `source_attribution` must be rendered wherever setlist data appears (setlist.fm policy)
- All playlists are private — verify `"public": false` in Spotify API calls
- Songs not found on Spotify go in `UnmatchedTracks` (strings), never silently dropped
- Commit after each checkpoint at minimum; commit after logical groups within phases
