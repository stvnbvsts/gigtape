<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { PlaylistResult } from '../api/client'

const result = ref<PlaylistResult | null>(null)

onMounted(() => {
  const raw = sessionStorage.getItem('gigtape_last_result')
  if (raw) {
    result.value = JSON.parse(raw) as PlaylistResult
  }
})

const hasUnmatched = computed(() => (result.value?.unmatched_tracks?.length ?? 0) > 0)
</script>

<template>
  <section class="playlist-result">
    <template v-if="result">
      <h1>Playlist created</h1>
      <p>
        <a :href="result.playlist_url" target="_blank" rel="noopener">Open in Spotify</a>
      </p>
      <p>{{ result.matched_tracks.length }} tracks added.</p>

      <!-- Unmatched tracks are always displayed explicitly — never hidden or collapsed. -->
      <template v-if="hasUnmatched">
        <h2>Not found on Spotify</h2>
        <ul>
          <li v-for="(t, i) in result.unmatched_tracks" :key="i">{{ t }}</li>
        </ul>
      </template>

      <p>
        <router-link to="/">Create another</router-link>
      </p>
    </template>
    <template v-else>
      <p>No result to show. <router-link to="/">Start again</router-link>.</p>
    </template>
  </section>
</template>

<style scoped>
.playlist-result {
  max-width: 640px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
</style>
