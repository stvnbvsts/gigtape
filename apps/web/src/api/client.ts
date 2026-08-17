// Typed fetch wrapper and API client. The session ID is module-level state set
// after OAuth callback completes; every request sends it in the X-Session-ID
// header.

export type MusicService = 'spotify' | 'apple_music'

export interface Artist {
  name: string
  disambiguation: string
  external_ref: string
}

export interface Track {
  title: string
  artist_name: string
}

export interface Setlist {
  event_name: string
  date: string
  tracks: Track[]
  source_attribution: string
  track_count: number
}

export interface PlaylistResult {
  playlist_url: string
  matched_tracks: Track[]
  unmatched_tracks: string[]
  skipped_artists: string[]
}

export interface ArtistPlaylistRequest {
  artist_ref: string
  artist_name: string
  setlist_index: number
  event_date: string
  tracks: Track[]
}

export interface Event {
  name: string
  date: string
  location: string
  artists: Artist[]
  lineup_complete: boolean
}

export interface FestivalArtistEntry {
  artist_ref: string
  artist_name: string
  include: boolean
  tracks: Track[]
}

export interface FestivalPlaylistRequest {
  event_name: string
  event_date: string
  mode: 'merged' | 'per_artist'
  artists: FestivalArtistEntry[]
}

export interface FestivalResultEntry extends PlaylistResult {
  error?: string
}

export interface AppleMusicDeveloperToken {
  developer_token: string
  storefront: string
}

export interface PlaylistSummary {
  id: string
  name: string
  description: string
  track_count: number
  owner_name: string
}

export interface SourcePlaylist {
  name: string
  description: string
  tracks: Track[]
}

// Spotify's open.spotify.com links open the web player even with the app
// installed (desktop). The spotify: URI scheme launches the native app
// directly when present, falling back to nothing otherwise — so callers
// should keep the open.spotify.com link as a fallback.
export function toSpotifyAppURI(playlistUrl: string): string | null {
  const m = playlistUrl.match(/open\.spotify\.com\/playlist\/([A-Za-z0-9]+)/)
  return m ? `spotify:playlist:${m[1]}` : null
}

let sessionId = ''
let sessionService: MusicService = 'spotify'

export function setSessionId(id: string, service: MusicService = 'spotify') {
  sessionId = id
  sessionService = service
  try {
    localStorage.setItem('gigtape_session_id', id)
    localStorage.setItem('gigtape_session_service', service)
    localStorage.setItem(serviceSessionKey(service), id)
  } catch {}
}

export function getSessionId(): string {
  if (sessionId) return sessionId
  try {
    const stored = localStorage.getItem('gigtape_session_id')
    if (stored) sessionId = stored
  } catch {}
  return sessionId
}

export function getSessionService(): MusicService {
  try {
    const stored = localStorage.getItem('gigtape_session_service')
    if (stored === 'spotify' || stored === 'apple_music') sessionService = stored
  } catch {}
  return sessionService
}

export function getServiceSessionId(service: MusicService): string {
  try {
    return localStorage.getItem(serviceSessionKey(service)) || ''
  } catch {
    return ''
  }
}

export function serviceLabel(service: MusicService = getSessionService()): string {
  return service === 'apple_music' ? 'Apple Music' : 'Spotify'
}

export function clearSessionId() {
  sessionId = ''
  sessionService = 'spotify'
  try {
    localStorage.removeItem('gigtape_session_id')
    localStorage.removeItem('gigtape_session_service')
  } catch {}
}

function serviceSessionKey(service: MusicService): string {
  return `gigtape_${service}_session_id`
}

// OAuth error banner state. The /auth/callback redirect appends
// ?oauth_error=<code> when the handshake fails and WEB_REDIRECT_URL is set;
// main.ts lifts it here so views can render a banner without plumbing it
// through the router.
let oauthError = ''

export function setOAuthError(code: string) {
  oauthError = code
}

export function consumeOAuthError(): string {
  const e = oauthError
  oauthError = ''
  return e
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers || {})
  if (sessionId || getSessionId()) headers.set('X-Session-ID', getSessionId())
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')

  const resp = await fetch(`/api${path}`, { ...init, headers })
  if (!resp.ok) {
    const body = await resp.json().catch(() => ({ error: 'unknown', message: resp.statusText }))
    throw new Error(body.message || body.error || `HTTP ${resp.status}`)
  }
  return resp.json() as Promise<T>
}

export function getAuthUrl(): Promise<{ auth_url: string }> {
  return request('/auth/login')
}

export function getAppleMusicDeveloperToken(): Promise<AppleMusicDeveloperToken> {
  return request('/auth/apple-music/developer-token')
}

export function createAppleMusicSession(
  musicUserToken: string,
): Promise<{ session_id: string; service: MusicService }> {
  return request('/auth/apple-music/session', {
    method: 'POST',
    body: JSON.stringify({ music_user_token: musicUserToken }),
  })
}

declare global {
  interface Window {
    MusicKit?: {
      configure(config: {
        developerToken: string
        app: { name: string; build: string }
      }): MusicKitInstance
      getInstance(): MusicKitInstance
    }
  }
}

interface MusicKitInstance {
  authorize(): Promise<string>
}

let musicKitScript: Promise<void> | null = null

function loadMusicKit(): Promise<void> {
  if (window.MusicKit) return Promise.resolve()
  if (musicKitScript) return musicKitScript
  musicKitScript = new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = 'https://js-cdn.music.apple.com/musickit/v1/musickit.js'
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Could not load Apple Music. Please try again.'))
    document.head.appendChild(script)
  })
  return musicKitScript
}

export async function connectAppleMusic(): Promise<{ session_id: string; service: MusicService }> {
  const cfg = await getAppleMusicDeveloperToken()
  await loadMusicKit()
  if (!window.MusicKit) throw new Error('Apple Music is unavailable. Please try again.')
  window.MusicKit.configure({
    developerToken: cfg.developer_token,
    app: { name: 'Gigtape', build: 'web' },
  })
  const musicUserToken = await window.MusicKit.getInstance().authorize()
  const session = await createAppleMusicSession(musicUserToken)
  setSessionId(session.session_id, session.service)
  return session
}

export function searchArtists(q: string): Promise<{ artists: Artist[] }> {
  return request(`/artists/search?q=${encodeURIComponent(q)}`)
}

export function getSetlists(
  artistRef: string,
  artistName: string,
): Promise<{ setlists: Setlist[]; short_warning: boolean }> {
  return request(
    `/setlists?artist_ref=${encodeURIComponent(artistRef)}&artist_name=${encodeURIComponent(artistName)}`,
  )
}

export function createArtistPlaylist(body: ArtistPlaylistRequest): Promise<PlaylistResult> {
  return request('/playlists/artist', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function searchEvents(q: string): Promise<{ events: Event[] }> {
  return request(`/events/search?q=${encodeURIComponent(q)}`)
}

export function createFestivalPlaylist(
  body: FestivalPlaylistRequest,
): Promise<{ results: FestivalResultEntry[] }> {
  return request('/playlists/festival', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function listSpotifyTransferPlaylists(
  spotifySessionId = getServiceSessionId('spotify'),
): Promise<{ playlists: PlaylistSummary[] }> {
  return request('/transfer/spotify/playlists', {
    headers: { 'X-Session-ID': spotifySessionId },
  })
}

export function getSpotifyTransferPlaylist(
  playlistId: string,
  spotifySessionId = getServiceSessionId('spotify'),
): Promise<SourcePlaylist> {
  return request(`/transfer/spotify/playlists/${encodeURIComponent(playlistId)}`, {
    headers: { 'X-Session-ID': spotifySessionId },
  })
}

export function transferSpotifyToAppleMusic(
  playlistId: string,
  spotifySessionId = getServiceSessionId('spotify'),
  appleMusicSessionId = getServiceSessionId('apple_music'),
): Promise<PlaylistResult> {
  return request('/transfer/spotify-to-apple-music', {
    method: 'POST',
    headers: {
      'X-Spotify-Session-ID': spotifySessionId,
      'X-Apple-Music-Session-ID': appleMusicSessionId,
    },
    body: JSON.stringify({ playlist_id: playlistId }),
  })
}
