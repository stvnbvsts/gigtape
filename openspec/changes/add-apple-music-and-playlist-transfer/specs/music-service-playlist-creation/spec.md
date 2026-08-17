## Purpose

Allows web users to create setlist-based Gigtape playlists in the music service they choose at the start of the session, beginning with Spotify and Apple Music.

## ADDED Requirements

### Requirement: User chooses a playlist service
The system SHALL let web users start a playlist creation session by choosing either Spotify or Apple Music as the destination music service.

#### Scenario: Start Spotify session
- **WHEN** the user chooses Spotify at app start and completes authentication
- **THEN** the system starts a Spotify-backed session for subsequent setlist and playlist creation requests

#### Scenario: Start Apple Music session
- **WHEN** the user chooses Apple Music at app start and completes authorization in the web app
- **THEN** the system starts an Apple Music-backed session for subsequent setlist and playlist creation requests

#### Scenario: Session service is visible to protected flows
- **WHEN** a protected web request is made with a valid session
- **THEN** the system can determine which music service the session targets

### Requirement: Create artist playlist in selected service
The system SHALL create a private artist setlist playlist in the music service selected for the active web session.

#### Scenario: Spotify artist playlist
- **WHEN** an authenticated Spotify web session submits an artist playlist creation request
- **THEN** the system creates the playlist in Spotify and returns a Spotify playlist URL

#### Scenario: Apple Music artist playlist
- **WHEN** an authenticated Apple Music web session submits an artist playlist creation request
- **THEN** the system creates the playlist in Apple Music and returns an Apple Music playlist URL

#### Scenario: Artist playlist unmatched tracks
- **WHEN** one or more requested tracks cannot be matched in the selected music service
- **THEN** the response lists those unmatched tracks without hiding the successfully created playlist

### Requirement: Create festival playlists in selected service
The system SHALL create festival playlists in the music service selected for the active web session while preserving the existing merged and per-artist modes.

#### Scenario: Merged festival playlist
- **WHEN** an authenticated web session submits a merged festival playlist request
- **THEN** the system creates one playlist in the selected music service using the included artists' edited tracks

#### Scenario: Per-artist festival playlists
- **WHEN** an authenticated web session submits a per-artist festival playlist request
- **THEN** the system creates one playlist per included artist with tracks in the selected music service

#### Scenario: Festival skipped and unmatched reporting
- **WHEN** festival artists are skipped or tracks cannot be matched in the selected music service
- **THEN** the response reports skipped artists and unmatched tracks for user review

### Requirement: Apple Music support is web-only
The system SHALL limit Apple Music playlist creation to the web app for this change.

#### Scenario: CLI remains Spotify-only
- **WHEN** a user runs the CLI playlist creation flow
- **THEN** the CLI continues to use Spotify authentication and Spotify playlist creation behavior

### Requirement: Service-specific authentication errors are clear
The system SHALL present authentication and authorization errors using the selected music service's name.

#### Scenario: Apple Music authorization fails
- **WHEN** Apple Music authorization fails or expires during a web session
- **THEN** the user sees an Apple Music-specific reconnection message

#### Scenario: Spotify authorization fails
- **WHEN** Spotify authentication fails or expires during a web session
- **THEN** the user sees a Spotify-specific reconnection message
