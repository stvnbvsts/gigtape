## Purpose

Allows web users to copy their own Spotify playlists into Apple Music without changing or deleting the original Spotify playlist.

## ADDED Requirements

### Requirement: Transfer requires both services
The system SHALL require an authenticated Spotify connection and an authenticated Apple Music web connection before transferring a playlist.

#### Scenario: Missing Spotify connection
- **WHEN** the user opens the transfer tool without a valid Spotify connection
- **THEN** the system prompts the user to connect Spotify before listing source playlists

#### Scenario: Missing Apple Music connection
- **WHEN** the user attempts to copy a Spotify playlist without a valid Apple Music connection
- **THEN** the system prompts the user to connect Apple Music before starting the copy

#### Scenario: Both connections available
- **WHEN** the user has valid Spotify and Apple Music connections
- **THEN** the system allows the user to select a Spotify playlist for transfer

### Requirement: List user's Spotify playlists
The system SHALL list Spotify playlists owned by or available to the authenticated Spotify user for selection as transfer sources.

#### Scenario: Playlists listed
- **WHEN** the authenticated Spotify user opens the transfer source selector
- **THEN** the system shows that user's Spotify playlists with enough identifying information to choose one

#### Scenario: No playlists found
- **WHEN** the authenticated Spotify user has no selectable playlists
- **THEN** the system shows an empty state and does not offer a transfer action

### Requirement: Preview selected Spotify playlist
The system SHALL let the user preview the selected Spotify playlist before copying it to Apple Music.

#### Scenario: Playlist preview loaded
- **WHEN** the user selects a Spotify playlist
- **THEN** the system shows the playlist title, description, and tracks in source order

#### Scenario: Playlist read fails
- **WHEN** the selected Spotify playlist cannot be read
- **THEN** the system reports the read failure and does not create an Apple Music playlist

### Requirement: Copy playlist to Apple Music
The system SHALL create a new Apple Music playlist from the selected Spotify playlist without modifying the Spotify playlist.

#### Scenario: Successful copy
- **WHEN** the user confirms transfer of a selected Spotify playlist
- **THEN** the system creates a new Apple Music playlist with the source title, source description, and matched tracks in source order

#### Scenario: Source playlist is not modified
- **WHEN** a playlist transfer succeeds or partially succeeds
- **THEN** the original Spotify playlist remains unchanged

#### Scenario: Unmatched Apple Music tracks
- **WHEN** one or more Spotify tracks cannot be matched in Apple Music
- **THEN** the transfer result lists the unmatched tracks and reports any playlist that was created

### Requirement: Transfer is copy-only
The system SHALL NOT delete, move, rename, or otherwise mutate the source Spotify playlist as part of transfer.

#### Scenario: User completes transfer
- **WHEN** the transfer operation finishes
- **THEN** no destructive action is performed against the source Spotify playlist
