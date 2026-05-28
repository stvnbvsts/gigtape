# API Contract: Gigtape REST API (Phase 1)

**Date**: 2026-04-21
**Branch**: `001-gigtape-baseline`
**Base URL**: `http://localhost:8080` (local development — no deployment target in Phase 1)
**Content-Type**: `application/json` (all requests and responses)
**Session auth**: `X-Session-ID: {uuid}` header required on all endpoints except `/auth/*`

---

## Error Response Shape

All error responses follow this shape:

```json
{
  "error": "machine_readable_code",
  "message": "Human-readable explanation suitable for display to the user."
}
```

**Common error codes**:

| Code | HTTP Status | Meaning |
|---|---|---|
| `session_not_found` | 401 | `X-Session-ID` header missing or session does not exist |
| `session_expired` | 401 | Session exists but OAuth token has expired; re-authenticate |
| `artist_not_found` | 404 | No artist matched the search query |
| `setlist_not_found` | 404 | No setlist found for the given artist |
| `upstream_error` | 502 | setlist.fm or Spotify returned an unexpected error |
| `rate_limited` | 429 | Upstream rate limit hit; see `Retry-After` response header |

---

## Authentication

### GET /auth/login

Initiates Spotify OAuth 2.0 PKCE flow. Returns the authorization URL.

**Response** `200 OK`:
```json
{
  "auth_url": "https://accounts.spotify.com/authorize?client_id=...&code_challenge=...&scope=playlist-modify-private+playlist-read-private&..."
}
```

---

### GET /auth/callback

OAuth callback handler. Spotify redirects here with `code` and `state` query parameters.
Creates an in-memory session and returns the session ID.

**Query params**: `code` (required), `state` (required)

**Response** `200 OK`:
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response** `400 Bad Request` (state mismatch or missing code):
```json
{
  "error": "oauth_error",
  "message": "OAuth handshake failed. Please try connecting your Spotify account again."
}
```

---

## Artists

### GET /artists/search

Search for artists by name. Returns candidates for the disambiguation confirmation step.
Returns an empty array when no artists match — never a 404.

**Query params**: `q` (required) — artist name search term

**Headers**: `X-Session-ID` (required)

**Response** `200 OK`:
```json
{
  "artists": [
    {
      "name": "Radiohead",
      "disambiguation": "rock band from Abingdon, UK",
      "external_ref": "a74b1b7f-71a5-4011-9441-d0b5e4122711"
    },
    {
      "name": "Radiohead",
      "disambiguation": "cover band",
      "external_ref": "b99f1b7f-22a5-9911-b441-d0b5e4199abc"
    }
  ]
}
```

**Response** `200 OK` (no matches):
```json
{
  "artists": []
}
```

---

## Setlists

### GET /setlists

Fetch recent setlists for a specific artist. Returns setlists most recent first.
Returns an empty array when no setlists exist — never a 404.

**Query params**: `artist_ref` (required) — opaque `external_ref` from `/artists/search`

**Headers**: `X-Session-ID` (required)

**Response** `200 OK`:
```json
{
  "setlists": [
    {
      "event_name": "Glastonbury 2024 — Pyramid Stage",
      "date": "2024-06-28",
      "tracks": [
        { "title": "Creep", "artist_name": "Radiohead" },
        { "title": "Karma Police", "artist_name": "Radiohead" },
        { "title": "Paranoid Android", "artist_name": "Radiohead" }
      ],
      "source_attribution": "setlist.fm • https://www.setlist.fm/setlist/radiohead/2024/worthy-farm-pilton-england-abc123.html",
      "track_count": 3
    }
  ]
}
```

**Notes**:
- `track_count` is a convenience field (equals `len(tracks)`).
- The UI MUST display `source_attribution` wherever setlist data is shown.
- When `track_count < 6`, the UI MUST warn the user before they proceed.

**Response** `200 OK` (no setlists found):
```json
{
  "setlists": []
}
```

---

## Events (Festivals)

### GET /events/search

Search for a festival or event by name. Uses setlist.fm's event search.
Partial lineups are expected and valid — the UI allows manual additions.

**Query params**: `q` (required) — festival or event name

**Headers**: `X-Session-ID` (required)

**Response** `200 OK`:
```json
{
  "events": [
    {
      "name": "Glastonbury 2024",
      "date": "2024-06-26",
      "location": "Worthy Farm, Pilton, Somerset, UK",
      "artists": [
        {
          "name": "Coldplay",
          "disambiguation": "",
          "external_ref": "cc197bad-dc9c-440d-a5b5-d52ba2e14234"
        },
        {
          "name": "SZA",
          "disambiguation": "",
          "external_ref": "9c9f1152-36a2-4f41-b56f-4f9948c9b341"
        }
      ],
      "lineup_complete": false
    }
  ]
}
```

**Notes**:
- `lineup_complete: false` signals to the UI that the lineup may be partial and manual additions
  should be offered.
- `artists` preserves source order — this is the order used for merged playlist track grouping.

**Response** `200 OK` (no matches):
```json
{
  "events": []
}
```

---

## Playlists

### POST /playlists/artist

Create a private Spotify playlist from a single artist's (possibly user-edited) setlist.

**Headers**: `X-Session-ID` (required)

**Request body**:
```json
{
  "artist_ref": "a74b1b7f-71a5-4011-9441-d0b5e4122711",
  "artist_name": "Radiohead",
  "setlist_index": 0,
  "tracks": [
    { "title": "Creep", "artist_name": "Radiohead" },
    { "title": "Karma Police", "artist_name": "Radiohead" }
  ]
}
```

**Notes**:
- `tracks` is the final user-edited track list. The user may have added or removed songs from
  the fetched setlist. This list is authoritative.
- `setlist_index` records which setlist the user selected (0 = latest). Informational only.
- `artist_name` used for playlist naming: `"{artist_name} — {YYYY-MM-DD}"`.

**Response** `200 OK`:
```json
{
  "playlist_url": "https://open.spotify.com/playlist/3cEYpjA9oz9GiPac4AsH21",
  "matched_tracks": [
    { "title": "Creep", "artist_name": "Radiohead" },
    { "title": "Karma Police", "artist_name": "Radiohead" }
  ],
  "unmatched_tracks": [],
  "skipped_artists": []
}
```

**Response** `200 OK` (partial success — some tracks not found):
```json
{
  "playlist_url": "https://open.spotify.com/playlist/3cEYpjA9oz9GiPac4AsH21",
  "matched_tracks": [
    { "title": "Creep", "artist_name": "Radiohead" }
  ],
  "unmatched_tracks": ["Some Rare B-Side Title"],
  "skipped_artists": []
}
```

**Response** `401 Unauthorized`:
```json
{
  "error": "session_expired",
  "message": "Your Spotify session has expired. Please reconnect your account."
}
```

---

### POST /playlists/festival

Create private Spotify playlist(s) from a festival lineup. Supports merged (one playlist) or
per-artist (one playlist per selected artist) modes.

**Headers**: `X-Session-ID` (required)

**Request body**:
```json
{
  "event_name": "Glastonbury 2024",
  "event_date": "2024-06-26",
  "mode": "merged",
  "artists": [
    {
      "artist_ref": "cc197bad-dc9c-440d-a5b5-d52ba2e14234",
      "artist_name": "Coldplay",
      "include": true,
      "tracks": [
        { "title": "Yellow", "artist_name": "Coldplay" },
        { "title": "Fix You", "artist_name": "Coldplay" }
      ]
    },
    {
      "artist_ref": "9c9f1152-36a2-4f41-b56f-4f9948c9b341",
      "artist_name": "SZA",
      "include": false,
      "tracks": []
    }
  ]
}
```

**Notes**:
- `mode`: `"merged"` or `"per_artist"`.
- `include: false` skips the artist entirely. Their songs are excluded from all playlists and
  their name is added to `skipped_artists` in the result.
- `tracks` is the user-edited list per artist. Artists with `include: true` and empty `tracks`
  are treated as having no setlist — surfaced in `skipped_artists`.
- In `"merged"` mode, tracks are ordered by artist in the order they appear in this array.

**Response** `200 OK`:
```json
{
  "results": [
    {
      "playlist_url": "https://open.spotify.com/playlist/3cEYpjA9oz9GiPac4AsH21",
      "matched_tracks": [
        { "title": "Yellow", "artist_name": "Coldplay" },
        { "title": "Fix You", "artist_name": "Coldplay" }
      ],
      "unmatched_tracks": [],
      "skipped_artists": ["SZA"]
    }
  ]
}
```

**Response** `207 Multi-Status` (per_artist mode, some playlists created, some failed):
```json
{
  "results": [
    {
      "playlist_url": "https://open.spotify.com/playlist/abc123",
      "matched_tracks": [...],
      "unmatched_tracks": [],
      "skipped_artists": []
    },
    {
      "playlist_url": "",
      "matched_tracks": [],
      "unmatched_tracks": [],
      "skipped_artists": ["Artist With No Tracks"],
      "error": "No tracks available for this artist."
    }
  ]
}
```

**Notes**:
- `207 Multi-Status` used for `per_artist` mode when at least one playlist succeeded and at
  least one did not. Frontend renders all results individually.
- `200 OK` used when the outcome is a single result (merged mode) or all results succeeded.
