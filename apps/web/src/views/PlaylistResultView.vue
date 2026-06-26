<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { toSpotifyAppURI, type PlaylistResult } from '../api/client'

const result = ref<PlaylistResult | null>(null)

onMounted(() => {
  const raw = sessionStorage.getItem('gigtape_last_result')
  if (raw) {
    result.value = JSON.parse(raw) as PlaylistResult
  }
})

const hasUnmatched = computed(() => (result.value?.unmatched_tracks?.length ?? 0) > 0)
const appURI = computed(() => (result.value ? toSpotifyAppURI(result.value.playlist_url) : null))
const artistName = computed(() => result.value?.matched_tracks?.[0]?.artist_name || 'Gigtape')
</script>

<template>
  <section class="gt-center">
    <template v-if="result">
      <div class="gt-result-eyebrow">★ your tape is ready ★</div>

      <div class="gt-cassette gt-result-cassette">
        <div class="gt-cassette__label">
          <div class="gt-cassette-head">
            <span class="gt-cassette__name">
              {{ artistName }}<br />
              <span class="gt-cassette-subtitle">live · setlist mix</span>
            </span>
            <span class="gt-side">SIDE A</span>
          </div>
          <div class="gt-cassette__divider"></div>
          <div class="gt-cassette__reels">
            <div class="gt-reel"><div class="gt-reel__hub"></div></div>
            <div class="gt-reel gt-reel--fast"><div class="gt-reel__hub"></div></div>
          </div>
        </div>
      </div>

      <p class="gt-result-copy">{{ result.matched_tracks.length }} songs added to Spotify.</p>

      <a
        v-if="appURI"
        class="gt-btn gt-btn--spotify gt-result-open"
        :href="appURI"
      >
        ▸ Open in Spotify
      </a>
      <a
        v-else
        class="gt-btn gt-btn--spotify gt-result-open"
        :href="result.playlist_url"
        target="_blank"
        rel="noopener"
      >
        ▸ Open in Spotify
      </a>
      <a
        v-if="appURI"
        class="gt-browser-fallback"
        :href="result.playlist_url"
        target="_blank"
        rel="noopener"
      >
        no app installed? open in browser
      </a>

      <!-- Unmatched tracks are always displayed explicitly — never hidden or collapsed. -->
      <div v-if="hasUnmatched" class="gt-panel gt-result-panel">
        <div class="gt-panel__label">COULDN'T FIND THESE ON SPOTIFY —</div>
        <div v-for="(t, i) in result.unmatched_tracks" :key="i" class="gt-panel__item">· {{ t }}</div>
      </div>

      <p class="gt-result-signoff">Now go. Pass it on. ♪</p>
      <router-link class="gt-link" to="/">make another tape</router-link>
    </template>
    <template v-else>
      <p class="gt-empty-note">No result to show.</p>
      <router-link class="gt-link" to="/">start again</router-link>
    </template>
  </section>
</template>
