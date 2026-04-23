<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { FestivalResultEntry } from '../api/client'

const results = ref<FestivalResultEntry[]>([])

onMounted(() => {
  const raw = sessionStorage.getItem('gigtape_festival_results')
  if (raw) results.value = JSON.parse(raw) as FestivalResultEntry[]
})

const skipped = computed(() => {
  const all = new Set<string>()
  for (const r of results.value) {
    for (const s of r.skipped_artists || []) all.add(s)
  }
  return [...all]
})
</script>

<template>
  <section class="festival-result">
    <h1>Festival playlists</h1>

    <article v-for="(r, i) in results" :key="i" class="result-card">
      <h2>Playlist #{{ i + 1 }}</h2>
      <p v-if="r.playlist_url">
        <a :href="r.playlist_url" target="_blank" rel="noopener">Open in Spotify</a>
      </p>
      <p v-else class="error">
        Could not be created. {{ r.error || '' }}
      </p>
      <p>{{ r.matched_tracks.length }} tracks matched.</p>
      <template v-if="r.unmatched_tracks.length">
        <h3>Not found on Spotify</h3>
        <ul>
          <li v-for="(t, j) in r.unmatched_tracks" :key="j">{{ t }}</li>
        </ul>
      </template>
    </article>

    <section v-if="skipped.length" class="skipped">
      <h2>Skipped artists</h2>
      <ul>
        <li v-for="(s, i) in skipped" :key="i">{{ s }}</li>
      </ul>
    </section>

    <p>
      <router-link to="/">Back to home</router-link>
    </p>
  </section>
</template>

<style scoped>
.festival-result {
  max-width: 680px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
.result-card {
  border: 1px solid #ddd;
  padding: 1rem;
  margin: 1rem 0;
  border-radius: 4px;
}
.error {
  color: #b00;
}
.skipped {
  background: #faf6ee;
  padding: 0.5rem 1rem;
  border-radius: 4px;
}
</style>
