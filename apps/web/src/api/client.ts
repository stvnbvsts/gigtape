// Typed fetch wrapper and API client. The session ID is module-level state set
// after OAuth callback completes; every request sends it in the X-Session-ID
// header (see contracts/api.md).

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

let sessionId = ''

export function setSessionId(id: string) {
  sessionId = id
  try {
    localStorage.setItem('gigtape_session_id', id)
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

export function clearSessionId() {
  sessionId = ''
  try {
    localStorage.removeItem('gigtape_session_id')
  } catch {}
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
