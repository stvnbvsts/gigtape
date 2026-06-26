<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { toSpotifyAppURI, type FestivalResultEntry } from '../api/client'

const results = ref<FestivalResultEntry[]>([])
const eventName = ref('Festival mix')
const mode = ref<'merged' | 'per_artist'>('merged')

onMounted(() => {
  const raw = sessionStorage.getItem('gigtape_festival_results')
  if (raw) results.value = JSON.parse(raw) as FestivalResultEntry[]
  const metaRaw = sessionStorage.getItem('gigtape_festival_result_meta')
  if (metaRaw) {
    const meta = JSON.parse(metaRaw) as { event_name?: string; mode?: 'merged' | 'per_artist' }
    eventName.value = meta.event_name || eventName.value
    mode.value = meta.mode || mode.value
  }
})

const skipped = computed(() => {
  const all = new Set<string>()
  for (const r of results.value) {
    for (const s of r.skipped_artists || []) all.add(s)
  }
  return [...all]
})
const isMerged = computed(() => mode.value === 'merged' || results.value.length <= 1)
const totalMatched = computed(() =>
  results.value.reduce((sum, r) => sum + (r.matched_tracks?.length || 0), 0),
)
const totalArtists = computed(() => new Set(results.value.flatMap((r) => r.matched_tracks.map((t) => t.artist_name))).size)
</script>

<template>
  <section>
    <div class="gt-center gt-result-eyebrow">
      ★ {{ isMerged ? 'your festival tape is ready' : 'your tapes are ready' }} ★
    </div>

    <template v-if="isMerged && results[0]">
      <div class="gt-cassette gt-result-cassette gt-tilt-l">
        <div class="gt-cassette__label">
          <div class="gt-cassette-head">
            <span class="gt-cassette__name">{{ eventName }}</span>
            <span class="gt-side">MIX</span>
          </div>
          <div class="gt-cassette-meta">{{ totalMatched }} songs · {{ totalArtists }} artists</div>
        </div>
      </div>
      <div class="gt-center">
        <a
          v-if="results[0].playlist_url && toSpotifyAppURI(results[0].playlist_url)"
          class="gt-btn gt-btn--spotify gt-result-open"
          :href="toSpotifyAppURI(results[0].playlist_url)!"
        >
          ▸ Open in Spotify
        </a>
        <a
          v-else-if="results[0].playlist_url"
          class="gt-btn gt-btn--spotify gt-result-open"
          :href="results[0].playlist_url"
          target="_blank"
          rel="noopener"
        >
          ▸ Open in Spotify
        </a>
        <p v-else class="gt-panel gt-screen-message">Could not be created. {{ results[0].error || '' }}</p>
      </div>
    </template>

    <template v-else>
      <div class="gt-search-results">
        <a
          v-for="(r, i) in results"
          :key="i"
          class="gt-result-card gt-result-card--spotify"
          :href="r.playlist_url || '#'"
          target="_blank"
          rel="noopener"
        >
          <span class="gt-result-arrow">▸</span>
          <span class="gt-result-card__name">
            {{ r.matched_tracks[0]?.artist_name || `Playlist #${i + 1}` }}
          </span>
          <span class="gt-result-card__meta">{{ r.matched_tracks.length }} songs</span>
        </a>
      </div>
    </template>

    <div v-if="results.some((r) => r.unmatched_tracks.length)" class="gt-panel gt-result-panel">
      <div class="gt-panel__label">COULDN'T FIND THESE ON SPOTIFY —</div>
      <template v-for="(r, i) in results" :key="i">
        <div v-for="(t, j) in r.unmatched_tracks" :key="`${i}-${j}`" class="gt-panel__item">· {{ t }}</div>
      </template>
    </div>

    <section v-if="skipped.length" class="gt-panel gt-result-panel">
      <div class="gt-panel__label">SKIPPED — NO SETLIST DATA YET —</div>
      <div v-for="(s, i) in skipped" :key="i" class="gt-panel__item">· {{ s }}</div>
    </section>

    <p class="gt-center gt-result-signoff">
      <router-link class="gt-link" to="/">start over</router-link>
    </p>
  </section>
</template>
