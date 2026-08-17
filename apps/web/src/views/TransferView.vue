<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  connectAppleMusic,
  getAuthUrl,
  getServiceSessionId,
  getSpotifyTransferPlaylist,
  listSpotifyTransferPlaylists,
  transferSpotifyToAppleMusic,
  type PlaylistResult,
  type PlaylistSummary,
  type SourcePlaylist,
} from '../api/client'

const spotifySession = ref('')
const appleSession = ref('')
const playlists = ref<PlaylistSummary[]>([])
const selectedId = ref('')
const preview = ref<SourcePlaylist | null>(null)
const result = ref<PlaylistResult | null>(null)
const loading = ref(false)
const error = ref('')

const hasSpotify = computed(() => !!spotifySession.value)
const hasApple = computed(() => !!appleSession.value)
const canCopy = computed(() => hasSpotify.value && hasApple.value && !!selectedId.value && !!preview.value)

onMounted(() => {
  refreshSessions()
  if (hasSpotify.value) loadPlaylists()
})

function refreshSessions() {
  spotifySession.value = getServiceSessionId('spotify')
  appleSession.value = getServiceSessionId('apple_music')
}

async function connectSpotify() {
  error.value = ''
  localStorage.setItem('gigtape_post_spotify_auth_path', '/transfer')
  const { auth_url } = await getAuthUrl()
  window.location.href = auth_url
}

async function connectApple() {
  error.value = ''
  loading.value = true
  try {
    await connectAppleMusic()
    refreshSessions()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function loadPlaylists() {
  error.value = ''
  loading.value = true
  try {
    const res = await listSpotifyTransferPlaylists(spotifySession.value)
    playlists.value = res.playlists
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function selectPlaylist(id: string) {
  selectedId.value = id
  result.value = null
  error.value = ''
  loading.value = true
  try {
    preview.value = await getSpotifyTransferPlaylist(id, spotifySession.value)
  } catch (e) {
    preview.value = null
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function copyPlaylist() {
  if (!canCopy.value) return
  error.value = ''
  loading.value = true
  try {
    result.value = await transferSpotifyToAppleMusic(
      selectedId.value,
      spotifySession.value,
      appleSession.value,
    )
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section>
    <div class="gt-tape-title">Copy a playlist</div>
    <p class="gt-sub gt-screen-sub">Spotify to Apple Music.</p>

    <div class="gt-transfer-grid">
      <div class="gt-panel gt-transfer-panel">
        <div class="gt-panel__label">SPOTIFY</div>
        <p class="gt-panel__item">{{ hasSpotify ? 'Connected' : 'Connect Spotify to pick a source playlist.' }}</p>
        <button v-if="!hasSpotify" class="gt-btn gt-btn--spotify gt-btn--block" type="button" @click="connectSpotify">
          Connect Spotify
        </button>
      </div>

      <div class="gt-panel gt-transfer-panel">
        <div class="gt-panel__label">APPLE MUSIC</div>
        <p class="gt-panel__item">{{ hasApple ? 'Connected' : 'Connect Apple Music to create the copy.' }}</p>
        <button v-if="!hasApple" class="gt-btn gt-btn--apple gt-btn--block" type="button" :disabled="loading" @click="connectApple">
          Connect Apple Music
        </button>
      </div>
    </div>

    <p v-if="error" class="gt-panel gt-screen-message" role="alert">{{ error }}</p>

    <div v-if="hasSpotify" class="gt-search-results">
      <div class="gt-eyebrow">YOUR SPOTIFY PLAYLISTS</div>
      <p v-if="!loading && playlists.length === 0" class="gt-empty-note">No Spotify playlists found.</p>
      <button
        v-for="p in playlists"
        :key="p.id"
        type="button"
        class="gt-result-card gt-result-card--spotify"
        @click="selectPlaylist(p.id)"
      >
        <span>
          <span class="gt-result-card__name">{{ p.name }}</span>
          <span class="gt-result-card__meta">{{ p.track_count }} songs</span>
        </span>
        <span class="gt-result-arrow">→</span>
      </button>
    </div>

    <section v-if="preview" class="gt-panel gt-result-panel">
      <div class="gt-panel__label">{{ preview.name }}</div>
      <p v-if="preview.description" class="gt-panel__item">{{ preview.description }}</p>
      <div v-for="(t, i) in preview.tracks" :key="`${t.title}-${i}`" class="gt-panel__item">
        {{ i + 1 }}. {{ t.title }} — {{ t.artist_name }}
      </div>
      <button class="gt-btn gt-btn--apple gt-btn--block" type="button" :disabled="!canCopy || loading" @click="copyPlaylist">
        Copy to Apple Music
      </button>
    </section>

    <section v-if="result" class="gt-panel gt-result-panel">
      <div class="gt-panel__label">COPY COMPLETE</div>
      <p class="gt-panel__item">{{ result.matched_tracks.length }} songs copied.</p>
      <a class="gt-btn gt-btn--apple gt-result-open" :href="result.playlist_url" target="_blank" rel="noopener">
        ▸ Open in Apple Music
      </a>
      <template v-if="result.unmatched_tracks.length">
        <div class="gt-panel__label">COULDN'T FIND THESE ON APPLE MUSIC —</div>
        <div v-for="(t, i) in result.unmatched_tracks" :key="i" class="gt-panel__item">· {{ t }}</div>
      </template>
    </section>
  </section>
</template>
