# Data Model: Gigtape — Baseline (Phase 1)

**Date**: 2026-04-21
**Branch**: `001-gigtape-baseline`
**Plan**: [plan.md](plan.md)

All types below live in `packages/domain`. This package has **zero external imports**.
The Go compiler enforces this — any external import in `packages/domain/go.mod` is a
constitution violation.

---

## Domain Entities

### Artist

```go
// artist.go
type Artist struct {
    Name           string
    Disambiguation string // e.g., "rock band from Abingdon, UK" — shown in confirmation step
    ExternalRef    string // opaque adapter token (MBID for setlist.fm); domain never interprets it
}
```

**Rules**:
- `Name` is always non-empty.
- `ExternalRef` is treated as an opaque string by domain. Adapters assign meaning to it.
- `Disambiguation` may be empty; UI only shows it when non-empty.

---

### Track

```go
// track.go
type Track struct {
    Title      string
    ArtistName string // denormalized — needed for Spotify search query (artist+title)
}
```

**Rules**:
- `Title` is always non-empty.
- `ArtistName` duplicates the parent artist name intentionally — track can appear in a merged
  playlist and must carry its own attribution for the search query.

---

### Setlist

```go
// setlist.go
type Setlist struct {
    Artist            Artist
    EventName         string    // e.g., "Glastonbury 2024 Main Stage"
    Date              time.Time
    Tracks            []Track
    SourceAttribution string    // MUST be displayed wherever setlist data appears (setlist.fm policy)
}
```

**Rules**:
- `Tracks` is an empty slice (not nil) when no songs are known. Callers check `len(Tracks)`.
- `SourceAttribution` is always populated by the setlistfm adapter. Never empty on data returned
  from the adapter — empty string indicates a bug in the adapter.
- Setlists with `len(Tracks) < 6` are valid domain objects; the use case issues a warning to the
  caller, not an error.

---

### Event

```go
// event.go
type Event struct {
    Name     string
    Date     time.Time
    Location string
    Artists  []Artist // lineup in source order; order preserved for merged playlist grouping
}
```

**Rules**:
- `Artists` preserves source order. Merged playlist song grouping follows this order.
- `Artists` may be a partial lineup (some artists not returned by setlist.fm). This is normal —
  the use case surfaces what it found and allows manual additions.

---

### Playlist

```go
// playlist.go
type Playlist struct {
    Name      string
    Tracks    []Track
    CreatedAt time.Time
}
```

**Naming convention**:
- Single artist: `"{ArtistName} — {YYYY-MM-DD}"`
- Merged festival: `"{FestivalName} — {YYYY-MM-DD}"`
- Per-artist festival: `"{ArtistName} — {FestivalName} — {YYYY-MM-DD}"`

**Rules**:
- Name uniqueness is enforced by the PlaylistDestination adapter (appends date; if collision
  persists, adapter returns the existing name + suffix).
- `Tracks` is the final user-edited list — may differ from the fetched setlist.

---

### PlaylistResult

```go
// result.go
type PlaylistResult struct {
    PlaylistURL     string   // direct link to created playlist (e.g., Spotify URL)
    MatchedTracks   []Track  // tracks successfully found and added
    UnmatchedTracks []string // song titles not found in the music service
    SkippedArtists  []string // artist names with no setlist and no manual tracks provided
}
```

**Rules**:
- Always returned non-nil by use cases — never replaced with a plain `error`.
- `UnmatchedTracks` contains song *titles* as strings, not `Track` structs — sufficient for
  the UI summary ("3 songs couldn't be found").
- `SkippedArtists` populated when a festival artist had no setlist and the user provided no
  manual tracks, or the user explicitly deselected the artist.
- A result with a non-empty `PlaylistURL` and non-empty `UnmatchedTracks` is a valid partial
  success — the playlist was created, some songs were not found.

---

## Port Interfaces

```go
// ports.go
package domain

import "context"

// SetlistProvider fetches setlists from a setlist data source.
type SetlistProvider interface {
    // SearchArtists returns candidate artists for disambiguation.
    // Returns an empty slice (not an error) when no artists match.
    SearchArtists(ctx context.Context, name string) ([]Artist, error)

    // GetSetlists returns recent setlists for the artist, most recent first.
    // Returns an empty slice (not an error) when no setlists exist.
    GetSetlists(ctx context.Context, artist Artist) ([]Setlist, error)
}

// EventProvider fetches festival and event lineups.
type EventProvider interface {
    // SearchEvents returns events matching the given name.
    // Returns an empty slice (not an error) when no events match.
    SearchEvents(ctx context.Context, name string) ([]Event, error)
}

// PlaylistDestination creates playlists in a music service.
type PlaylistDestination interface {
    // CreatePlaylist creates the playlist and returns a structured result.
    // Returns a non-nil PlaylistResult even on partial failure.
    // Returns an error only for unrecoverable failures (e.g., auth failure, service down).
    CreatePlaylist(ctx context.Context, playlist Playlist) (PlaylistResult, error)
}
```

**Design notes**:
- All methods take `context.Context` as first argument — idiomatic Go, enables timeouts and
  cancellation propagation.
- "No results" is not an error — callers check `len(slice)`. Errors are reserved for unexpected
  failures (network error, auth failure, malformed response).
- `EventProvider` is separate from `SetlistProvider` because the implementations may differ
  (setlist.fm implements both in Phase 1; Ticketmaster implements only `EventProvider` in Phase 2).

---

## Session (Infrastructure — not in domain)

```go
// apps/api/middleware/session.go
type Session struct {
    ID        string        // UUID v4
    Token     *oauth2.Token // from golang.org/x/oauth2
    UserID    string        // Spotify user ID (needed for playlist creation endpoint)
    CreatedAt time.Time
    ExpiresAt time.Time     // CreatedAt + SESSION_TTL_MINUTES
}
```

**Location**: `apps/api/middleware/` — session management is an infrastructure concern, not a
domain concern. The domain knows nothing about sessions or OAuth tokens.

---

## Entity Relationships

```
Event (Festival)
  └── []Artist           (lineup, ordered)
        └── []Setlist    (fetched per artist; user selects one)
              └── []Track

Playlist
  └── []Track            (assembled from selected Setlist(s), user-edited)
        ↓
  PlaylistResult         (returned by PlaylistDestination.CreatePlaylist)
        ├── MatchedTracks
        ├── UnmatchedTracks
        └── SkippedArtists
```

---

## Canonical Term Reference

| Concept | Canonical Go Name | Avoid |
|---|---|---|
| Music performer | `Artist` | Musician, Band, Act |
| Song in a setlist | `Track` | Song, Item |
| Concert programme | `Setlist` | Program, Show, Gig |
| Concert or festival | `Event` | Gig, Concert (generic) |
| Music streaming playlist | `Playlist` | — |
| Operation outcome | `PlaylistResult` | Response, Output |
| Partial/full success | `PlaylistResult` (always) | binary `bool` or `error` |
