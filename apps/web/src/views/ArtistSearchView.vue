<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  getSessionId,
  searchArtists,
  setSessionId,
  type Artist,
} from '../api/client'

const query = ref('')
const artists = ref<Artist[]>([])
const loading = ref(false)
const error = ref('')
const manualSession = ref('')
const router = useRouter()

async function onSearch() {
  error.value = ''
  if (!query.value.trim()) return
  if (!getSessionId()) {
    error.value = 'Please connect Spotify first.'
    return
  }
  loading.value = true
  try {
    const res = await searchArtists(query.value.trim())
    artists.value = res.artists
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function pick(a: Artist) {
  router.push({
    path: '/setlist',
    query: { ref: a.external_ref, name: a.name },
  })
}

function applyManualSession() {
  if (manualSession.value.trim()) {
    setSessionId(manualSession.value.trim())
  }
}
</script>

<template>
  <section>
    <div class="gt-tape-title">Find a band</div>
    <p class="gt-sub gt-screen-sub">Who are you about to go see?</p>

    <form class="gt-row" @submit.prevent="onSearch">
      <input v-model="query" class="gt-input" placeholder="artist name…" />
      <button class="gt-btn gt-btn--sm" type="submit" :disabled="loading">GO</button>
    </form>

    <p v-if="error" class="gt-panel gt-screen-message" role="alert">{{ error }}</p>

    <div v-if="artists.length" class="gt-search-results">
      <div class="gt-eyebrow">{{ artists.length }} RESULTS</div>
      <button
        v-for="a in artists"
        :key="a.external_ref"
        type="button"
        class="gt-result-card"
        @click="pick(a)"
      >
        <span>
          <span class="gt-result-card__name">{{ a.name }}</span>
          <span v-if="a.disambiguation" class="gt-result-card__meta">{{ a.disambiguation }}</span>
        </span>
        <span class="gt-result-arrow">→</span>
      </button>
    </div>
    <p v-else-if="!loading && query" class="gt-empty-note">No artists found.</p>

    <button class="gt-link gt-link--hand gt-plain-link gt-festival-link" type="button" @click="router.push('/festival')">
      Going to a festival? Make the whole lineup →
    </button>

    <details class="gt-details">
      <summary>Troubleshooting</summary>
      <p>
        If the backend is running without <code>WEB_REDIRECT_URL</code> set,
        <code>/auth/callback</code> returns a JSON
        <code>{ "session_id": "…" }</code> page instead of redirecting here.
        Paste the UUID below to resume.
      </p>
      <div class="gt-row">
        <input v-model="manualSession" class="gt-input" placeholder="session id" />
        <button class="gt-btn gt-btn--sm" type="button" @click="applyManualSession">USE</button>
      </div>
    </details>
  </section>
</template>
