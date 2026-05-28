# Feature Specification: Gigtape — Baseline (Phase 1)

**Feature Branch**: `001-gigtape-baseline`
**Created**: 2026-04-21
**Status**: Draft
**Input**: User description: baseline specification for Gigtape

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Single Artist Playlist (Priority: P1)

A music fan is attending a solo artist concert next weekend. They open Gigtape, connect
their Spotify account, search for the artist by name, and see the most recent setlist.
They review the songs, remove one they dislike, then confirm. Gigtape creates a Spotify
playlist and returns a direct link the user can open immediately.

**Why this priority**: The single-artist flow is the simplest end-to-end path through
the system and the foundation all other flows build on. Delivering this first proves
the core value proposition.

**Independent Test**: Can be fully tested by connecting Spotify, searching for a single
artist with a known setlist, and verifying a playlist appears in the user's Spotify
account containing the expected songs.

**Acceptance Scenarios**:

1. **Given** a user has connected their Spotify account, **When** they search for an
   artist with a known recent setlist, **Then** they see the setlist with song titles
   and can proceed to create a playlist.
2. **Given** a user is previewing a setlist, **When** they remove a song and confirm,
   **Then** the created playlist omits that song.
3. **Given** a user confirms a setlist, **When** the playlist is created successfully,
   **Then** the user receives a direct link to the Spotify playlist.
4. **Given** an artist has multiple recent setlists, **When** the user views the
   setlist, **Then** the latest setlist is shown by default with an option to choose
   a different one.

---

### User Story 2 — Festival Playlist (Priority: P2)

A music fan is attending a multi-day festival. They open Gigtape, connect Spotify,
search for the festival, and Gigtape retrieves the full lineup. For each artist, it
fetches the latest setlist. The user sees all artists with song counts, deselects
artists they don't care about, and chooses between one merged playlist or one playlist
per artist. Gigtape creates the playlist(s) and returns the link(s).

**Why this priority**: Festivals are the hardest manual problem (tens of artists,
hundreds of songs) and the strongest differentiator for Gigtape. The single-artist
flow (P1) must be stable first.

**Independent Test**: Can be fully tested by searching for a known festival, verifying
all lineup artists appear with setlist data, deselecting artists, choosing a playlist
mode, and confirming the correct number of playlists appear in the user's Spotify
account.

**Acceptance Scenarios**:

1. **Given** a user searches for a festival, **When** the lineup is retrieved,
   **Then** they see all artists in the lineup, each with a song count sourced from
   their latest setlist.
2. **Given** a user is reviewing a festival lineup, **When** they deselect an artist,
   **Then** that artist's songs are excluded from all created playlists.
3. **Given** a user chooses "one merged playlist," **When** playlists are created,
   **Then** a single Spotify playlist is created containing all selected artists'
   songs.
4. **Given** a user chooses "one playlist per artist," **When** playlists are created,
   **Then** one Spotify playlist per selected artist is created, each named with the
   artist name and date.
5. **Given** an artist in the festival lineup has no setlist on setlist.fm, **When**
   the user reviews the lineup, **Then** that artist is shown with a "no setlist
   found" indicator and the user can add songs manually or skip that artist.

---

### Edge Cases

- What happens when no setlist is found for an artist?
  → Manual song entry is offered as a fallback. The flow never reaches a dead end.
- What happens when a setlist has fewer than 6 songs?
  → The user is warned with song count shown. They decide to proceed, add more songs,
  or abort.
- What happens when a song from the setlist is not found on Spotify?
  → The song is surfaced in the result summary after playlist creation. It is never
  silently skipped.
- What happens when a song matches the wrong version or artist on Spotify?
  → Search uses both artist name and song title to reduce false matches.
- What happens when an artist name is ambiguous?
  → A confirmation step shows the matched artist name and origin before proceeding.
- What happens when multiple recent setlists vary significantly?
  → The latest is used by default; the user can select a different one.
- What happens when Spotify rate limits are hit?
  → Requests are throttled and progress feedback is shown to the user.
- What happens when an OAuth token expires mid-session?
  → The user is prompted to re-authenticate with a clear, human-readable message.
- What happens when the user denies OAuth permissions?
  → The app explains what permissions are needed and why, and offers a retry.
- What happens when a playlist name already exists in the user's Spotify account?
  → The date is appended to make the name unique. Existing playlists are never
  overwritten silently.
- What happens when the festival lineup is incomplete?
  → What was found is shown; manual additions are always permitted.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to connect their Spotify account via OAuth before
  creating playlists.
- **FR-002**: Users MUST be able to search for an artist by name and retrieve their
  most recent setlist from setlist.fm.
- **FR-003**: Users MUST be able to search for a festival by name; where setlist.fm
  returns matching festival data, the lineup and per-artist setlists are retrieved
  automatically. Where the lineup is incomplete or the festival is not found, the user
  MUST be able to add artists manually to fill the gaps.
- **FR-004**: Users MUST be able to preview a setlist and add or remove individual
  songs before confirming.
- **FR-005**: When multiple recent setlists exist for an artist, the system MUST use
  the latest by default and MUST allow the user to select a different one.
- **FR-006**: Users MUST be able to create one Spotify playlist per artist or one merged
  playlist for all artists (festival flow only). In the merged playlist, songs MUST be
  grouped by artist in the order artists appear in the festival lineup.
- **FR-007**: Playlist names MUST include the artist name and date to prevent collisions
  with existing playlists. Playlists MUST be created as private (visible only to the
  owner; shareable via direct link).
- **FR-008**: After playlist creation, users MUST receive a direct link to each created
  playlist.
- **FR-009**: When no setlist is found for an artist, the system MUST offer manual song
  entry as a fallback and MUST NOT present a dead end.
- **FR-010**: Songs not found on Spotify MUST be surfaced in the result summary after
  creation. Failures MUST NOT be silently skipped.
- **FR-011**: Song searches MUST use both artist name and song title to reduce false
  matches.
- **FR-012**: When an artist name is ambiguous, the system MUST show a confirmation
  step before proceeding.
- **FR-013**: Setlist.fm attribution MUST be displayed wherever setlist data is shown.
- **FR-014**: The system MUST comply with Spotify Developer Policy at all times.
- **FR-015**: OAuth tokens MUST be session-scoped and discarded after use. No user data
  is persisted between sessions.

### Out of Scope for v1 (Phase 1)

- Persistent user accounts or login
- Playlist history or saved preferences
- Apple Music integration
- Bandsintown, Songkick integration
- Ticketmaster event discovery flow (Phase 2)
- Telegram bot, Discord bot delivery surfaces (Phase 2)
- Monetization of any kind

### Key Entities

- **Artist**: A musical performer identified by name; resolved to a specific entity via
  a confirmation step when ambiguous. Has zero or more setlists.
- **Setlist**: An ordered list of songs performed by an artist at a specific show.
  Sourced from setlist.fm. Has a date and a venue.
- **Song**: A title within a setlist. May or may not match a track on Spotify.
- **Festival**: A multi-artist event with an ordered lineup. Each lineup slot contains
  an artist and may have an associated setlist.
- **Playlist**: A named, ordered collection of Spotify tracks created in the user's
  Spotify account. Named with artist name and date; never overwrites an existing
  playlist.
- **PlaylistCreationResult**: A structured outcome of the playlist creation operation.
  Carries: created playlists with links, songs not found on Spotify, and any partial
  failures. Partial success is a first-class outcome — never collapsed into a generic
  error.
- **Session**: Ephemeral, in-memory scope that holds the OAuth token and working state
  for a single user interaction. Discarded after use.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user with a known artist can go from opening Gigtape to having a
  Spotify playlist ready in under 2 minutes for the single-artist flow.
- **SC-002**: A user can complete the festival flow for a 10-artist lineup in under
  5 minutes.
- **SC-003**: Songs not found on Spotify are always surfaced to the user — zero
  silent failures in any playlist creation result.
- **SC-004**: The app supports up to 25 concurrent beta users without degraded
  experience.
- **SC-005**: All Spotify OAuth flows complete successfully without cryptic errors
  for 95% of users on first attempt.
- **SC-006**: Partial successes (some songs found, some not) are presented clearly
  enough that users understand what was and was not created.

## Clarifications

### Session 2026-04-21

- Q: When a user searches for a festival in Phase 1, what actually happens? → A: Hybrid — festival name search via setlist.fm where available; manual artist entry to fill gaps in the lineup.
- Q: Should playlists be created as public or private in the user's Spotify account? → A: Private — visible only to the playlist owner; shareable via direct link.
- Q: How are songs ordered in a merged festival playlist? → A: Grouped by artist in the order artists appear in the festival lineup.

## Assumptions

- Users have an active Spotify account with sufficient permissions to create playlists.
- setlist.fm has reasonably current setlist data for mainstream artists; coverage gaps
  are handled by the manual fallback (FR-009).
- The festival lineup source in Phase 1 is setlist.fm festival search with manual artist
  entry as a fallback for gaps; full Ticketmaster integration is Phase 2.
- The web app delivery surface is sufficient for Phase 1 (no mobile app required).
- Sessions are short-lived (minutes to an hour); no long-running token refresh is
  required for v1.
- The beta user base (~25 users) does not require horizontal scaling or rate-limit
  pooling across users in v1.
- setlist.fm and Spotify API availability is assumed; the app degrades gracefully when
  either is unavailable but does not implement caching or offline mode in v1.
