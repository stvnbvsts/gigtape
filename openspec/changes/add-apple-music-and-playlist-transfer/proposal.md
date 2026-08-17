## Why

Gigtape currently creates playlists only in Spotify, but the setlist-to-playlist workflow should support Apple Music users without making them use a Spotify account. Users also need a copy-only way to bring their own Spotify playlists into Apple Music while preserving playlist metadata and track order.

## What Changes

- Add Apple Music as a web-only playlist destination for artist and festival setlist flows.
- Let users choose Spotify or Apple Music as their music service at app start, with the selected service determining the playlist destination for normal creation flows.
- Keep CLI playlist creation Spotify-only for this change.
- Add a transfer tool that requires both a Spotify connection and an Apple Music web connection.
- Let users list/select their own Spotify playlists and copy one playlist to Apple Music.
- Preserve transferred playlist title, description, and track order.
- Report tracks that cannot be matched in Apple Music.
- Keep transfer copy-only; no source playlist deletion, mutation, or move behavior.

## Capabilities

### New Capabilities

- `music-service-playlist-creation`: Provider-aware web playlist creation for Spotify and Apple Music from setlist and festival data.
- `playlist-transfer`: Copy-only transfer of the authenticated user's Spotify playlists into Apple Music.

### Modified Capabilities

- None.

## Impact

- Adds Apple Music web authentication and playlist creation support, including developer token and Music User Token handling.
- Extends session/auth concepts from Spotify-only to provider-aware web sessions.
- Adds a Spotify playlist source/read capability for the transfer workflow.
- Adds API endpoints and web UI states for provider selection, Apple Music setlist creation, Spotify playlist selection, transfer preview, and transfer result reporting.
- Updates docs and environment configuration for Apple Music developer credentials.
- Adds tests around Apple Music matching/creation, provider-aware sessions, Spotify playlist reading, and transfer orchestration.
