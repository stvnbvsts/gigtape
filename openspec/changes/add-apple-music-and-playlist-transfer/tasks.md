## 1. Domain and Session Foundations

- [ ] 1.1 Add provider-aware session state for web requests while preserving existing Spotify session behavior.
- [ ] 1.2 Add or extend domain playlist models to carry title, optional description, ordered tracks, and playlist result metadata needed by creation and transfer.
- [ ] 1.3 Add a playlist source/read port for listing source playlists and reading a selected playlist with ordered tracks.
- [ ] 1.4 Add unit tests for provider-aware session lookup, expiry messaging, and playlist metadata preservation.

## 2. Apple Music Adapter and API Wiring

- [ ] 2.1 Implement Apple Music developer token generation/configuration with tests that avoid logging or persisting token material.
- [ ] 2.2 Implement Apple Music catalog track search using title and artist matching with explicit not-found behavior.
- [ ] 2.3 Implement Apple Music playlist creation and track addition as a `PlaylistDestination`.
- [ ] 2.4 Add API endpoints needed by the web app to obtain/configure Apple Music authorization and create Apple Music-backed sessions from Music User Tokens.
- [ ] 2.5 Wire destination selection in the API composition root so protected playlist creation routes use Spotify or Apple Music based on the active session provider.
- [ ] 2.6 Add adapter and handler tests for successful Apple Music creation, partial unmatched results, auth failures, and upstream failures.

## 3. Web Provider Choice and Setlist Creation

- [ ] 3.1 Update the landing/start flow so users can choose Spotify or Apple Music before normal setlist playlist creation.
- [ ] 3.2 Add MusicKit web authorization for Apple Music and submit the Music User Token to the API session endpoint.
- [ ] 3.3 Update client session state to track selected music service and display service-specific auth/expiry errors.
- [ ] 3.4 Update artist and festival playlist creation views/results so labels, open links, and unmatched-track copy reflect the selected service.
- [ ] 3.5 Keep CLI flows and docs explicitly Spotify-only for Apple Music scope.
- [ ] 3.6 Add web build/type checks and focused component/client tests where existing test infrastructure supports them.

## 4. Spotify Playlist Source

- [ ] 4.1 Implement Spotify playlist listing for the authenticated user's selectable playlists.
- [ ] 4.2 Implement Spotify playlist reading with title, description, and tracks in source order.
- [ ] 4.3 Add API endpoints for listing Spotify transfer sources and previewing a selected source playlist.
- [ ] 4.4 Add tests for empty playlist lists, pagination, playlist read failures, and ordered track extraction.

## 5. Transfer Use Case and API

- [ ] 5.1 Implement a copy-only transfer use case that reads a Spotify playlist source and creates a new Apple Music destination playlist.
- [ ] 5.2 Ensure transfer preserves source title, source description, and matched track order.
- [ ] 5.3 Ensure transfer reports Apple Music unmatched tracks and never mutates the source Spotify playlist.
- [ ] 5.4 Add API handler(s) for transfer confirmation and transfer result reporting.
- [ ] 5.5 Add use case and API tests for successful transfer, partial unmatched transfer, missing provider sessions, and source read failure.

## 6. Transfer Web Experience

- [ ] 6.1 Add a transfer entry point that communicates both Spotify and Apple Music connections are required.
- [ ] 6.2 Add connection states for missing Spotify, missing Apple Music, and both connected.
- [ ] 6.3 Add Spotify playlist selector and empty-state handling.
- [ ] 6.4 Add selected playlist preview with title, description, and ordered tracks.
- [ ] 6.5 Add transfer confirmation and result views with Apple Music target link and unmatched-track reporting.

## 7. Documentation and Verification

- [ ] 7.1 Update README and environment examples with Apple Music developer credentials and web-only Apple Music scope.
- [ ] 7.2 Document transfer behavior as copy-only and limited to the authenticated user's Spotify playlists.
- [ ] 7.3 Run Go tests for domain, use cases, Spotify adapter, Apple Music adapter, API, and CLI packages.
- [ ] 7.4 Run the web build/type-check workflow.
- [ ] 7.5 Validate the OpenSpec change artifacts before implementation begins.
