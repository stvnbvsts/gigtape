## Context

See proposal.md for motivation. The current codebase already has a provider-neutral `PlaylistDestination` behavior contract in the domain layer and Spotify implements it for setlist-based playlist creation. The web/API authentication and session model are Spotify-specific, and the CLI also assumes Spotify. Apple Music requires a different web-centered authorization shape: backend requests need an Apple developer token, while user-library writes require a Music User Token obtained from the web app through MusicKit.

The repo also contains an `applemusic` adapter module placeholder with only `go.mod`; no Apple Music behavior exists yet.

## Goals / Non-Goals

**Goals:**

- Preserve the existing ports-and-adapters architecture.
- Add provider-aware web sessions while keeping setlist preview independent of music service.
- Reuse the existing playlist creation use cases for Spotify and Apple Music destinations.
- Add a separate transfer use case instead of folding source-playlist reading into setlist creation.
- Keep Apple Music token material out of long-term storage.
- Keep the CLI's existing Spotify behavior stable.

**Non-Goals:**

- No Apple Music CLI flow.
- No arbitrary Spotify playlist URL transfer.
- No account database, playlist history, or persistent multi-account profile.
- No destructive transfer behavior.
- No bidirectional Apple Music to Spotify transfer.

## Decisions

### Provider-aware session model

Represent authenticated web sessions with a music service identifier and service-specific credential payload. Spotify sessions continue to carry OAuth token and user ID. Apple Music sessions carry a Music User Token and enough metadata to create Apple Music API clients. Session expiry and deletion remain in-memory and short-lived.

Alternatives considered:

- Separate session stores per provider. This would avoid a tagged credential payload but would duplicate expiration, lookup, and middleware behavior.
- Persisted user accounts. This is broader than the current stateless beta model and would add security/migration work not required for this feature.

### Apple Music user authorization happens in the web app

Use MusicKit on the web to authorize Apple Music and provide the resulting Music User Token to the API for session creation. The API signs or serves Apple developer tokens using configured Apple Music developer credentials, then uses both the developer token and Music User Token for catalog search and library playlist writes.

Alternatives considered:

- Native CLI Apple Music auth. MusicKit's browser-centered flow makes this significantly less straightforward and is explicitly out of scope.
- API-only OAuth-style Apple Music login. Apple Music library access does not mirror Spotify's PKCE flow for this use case.

### Keep `PlaylistDestination` for creation; add source behavior for transfer

Apple Music playlist creation should implement the existing destination behavior. Transfer should introduce a separate source-reader behavior for listing and reading Spotify playlists, because "read source playlist" is not the same responsibility as "create destination playlist."

Expected transfer flow:

```text
Spotify playlist source
  -> domain playlist with title, description, ordered tracks
  -> Apple Music destination
  -> transfer result with target URL and unmatched tracks
```

Alternatives considered:

- Add read methods to `PlaylistDestination`. This would make every destination look like a source even when the feature only needs Spotify reads.
- Build transfer directly in handlers. That would bypass the use case layer and make CLI/API/web reuse harder later.

### Preserve metadata in the domain playlist shape

The current domain playlist has name and tracks. Transfer needs source description preservation, so extend the domain playlist representation or introduce a transfer-specific playlist model that carries title, description, tracks, and source metadata. The final implementation should choose the smallest model that avoids leaking provider-specific fields into setlist creation.

Alternatives considered:

- Encode description only in the transfer request. This keeps the existing playlist model unchanged but complicates reuse of Apple Music destination creation.
- Provider-specific transfer DTOs. This would make the first transfer path quick but would work against future additional services.

### Matching remains best-effort with explicit unmatched reporting

Both Apple Music creation and transfer should search Apple Music catalog songs using title and artist data and add only matched tracks. Unmatched tracks remain visible to the user, matching the current Spotify behavior.

Alternatives considered:

- Fail the entire playlist when any track is unmatched. This is harsher than current Gigtape behavior and poor for live setlist data.
- Use ISRC-only matching. ISRC can improve precision for Spotify-to-Apple transfer but setlist data does not provide it, so title/artist matching remains required.

## Risks / Trade-offs

- Apple Music developer token signing and key management adds sensitive configuration -> Store private key material in environment/secrets, document required variables, and avoid logging token material.
- MusicKit authorization is browser-specific -> Keep Apple Music web-only and preserve Spotify-only CLI behavior.
- Track matching quality can vary between catalogs -> Report unmatched tracks explicitly and preserve successful partial creation.
- Provider-aware sessions increase API complexity -> Keep the session abstraction small and keep service-specific logic at adapter/composition boundaries.
- Transfer can create duplicate Apple Music playlists on retry -> Make the UI clear that transfer creates a new playlist and report the created target URL after completion.

## Migration Plan

1. Introduce provider-aware session and auth endpoints while preserving existing Spotify route compatibility where practical.
2. Add Apple Music configuration and web authorization flow behind the provider selection UI.
3. Wire Apple Music as an alternate destination for existing web artist and festival playlist creation.
4. Add Spotify source-reading endpoints and the transfer use case/UI.
5. Update docs and tests.

Rollback is to hide Apple Music and transfer entry points in the web UI while leaving existing Spotify setlist creation routes intact.
