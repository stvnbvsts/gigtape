<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { getAuthUrl, getSessionId, searchArtists, setSessionId, type Artist } from '../api/client'

const query = ref('')
const artists = ref<Artist[]>([])
const loading = ref(false)
const error = ref('')
const manualSession = ref('')
const router = useRouter()

async function connectSpotify() {
  try {
    const { auth_url } = await getAuthUrl()
    window.location.href = auth_url
  } catch (e) {
    error.value = (e as Error).message
  }
}

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
  <section class="artist-search">
    <h1>Gigtape</h1>
    <p>Create a Spotify playlist from an artist's recent setlists.</p>

    <button type="button" @click="connectSpotify">Connect Spotify</button>

    <details class="manual-session">
      <summary>Already authenticated? Paste session ID</summary>
      <input v-model="manualSession" placeholder="session id from /auth/callback" />
      <button type="button" @click="applyManualSession">Use session</button>
    </details>

    <form @submit.prevent="onSearch">
      <input v-model="query" placeholder="Artist name" />
      <button type="submit" :disabled="loading">Search</button>
    </form>

    <p v-if="error" class="error">{{ error }}</p>

    <ul v-if="artists.length" class="artist-list">
      <li v-for="a in artists" :key="a.external_ref">
        <button type="button" @click="pick(a)">
          <strong>{{ a.name }}</strong>
          <span v-if="a.disambiguation"> — {{ a.disambiguation }}</span>
        </button>
      </li>
    </ul>
    <p v-else-if="!loading && query">No artists found.</p>
  </section>
</template>

<style scoped>
.artist-search {
  max-width: 640px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
.error {
  color: #b00;
}
.manual-session {
  margin: 0.5rem 0;
}
</style>
