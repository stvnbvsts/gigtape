<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  consumeOAuthError,
  getAuthUrl,
  getSessionId,
  searchArtists,
  setSessionId,
  type Artist,
} from '../api/client'

const query = ref('')
const artists = ref<Artist[]>([])
const loading = ref(false)
const error = ref('')
const oauthBanner = ref('')
const manualSession = ref('')
const router = useRouter()

onMounted(() => {
  const code = consumeOAuthError()
  if (code) {
    oauthBanner.value =
      code === 'profile_error'
        ? 'We could not retrieve your Spotify profile. Please try again.'
        : 'OAuth handshake failed. Please try connecting your Spotify account again.'
  }
})

async function connectSpotify() {
  oauthBanner.value = ''
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

    <p v-if="oauthBanner" class="banner" role="alert">{{ oauthBanner }}</p>

    <button type="button" @click="connectSpotify">Connect Spotify</button>
    <p>
      Looking for a festival?
      <router-link to="/festival">Festival mode →</router-link>
    </p>

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

    <details class="troubleshooting">
      <summary>Troubleshooting</summary>
      <p>
        If the backend is running without <code>WEB_REDIRECT_URL</code> set,
        <code>/auth/callback</code> returns a JSON
        <code>{ "session_id": "…" }</code> page instead of redirecting here.
        Paste the UUID below to resume.
      </p>
      <input v-model="manualSession" placeholder="session id" />
      <button type="button" @click="applyManualSession">Use session</button>
    </details>
  </section>
</template>

<style scoped>
.artist-search {
  max-width: 640px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
.banner {
  background: #fff3f3;
  border: 1px solid #e0a0a0;
  color: #900;
  padding: 0.5rem 1rem;
  border-radius: 4px;
}
.error {
  color: #b00;
}
.troubleshooting {
  margin-top: 2rem;
  font-size: 0.9em;
  color: #555;
}
</style>
