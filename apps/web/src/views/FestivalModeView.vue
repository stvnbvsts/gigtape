<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  createFestivalPlaylist,
  type FestivalArtistEntry,
  type FestivalResultEntry,
} from '../api/client'

interface LineupRow extends FestivalArtistEntry {
  hasSetlist: boolean
  attribution: string
}

interface FestivalState {
  event_name: string
  event_date: string
  lineup: LineupRow[]
}

const router = useRouter()
const state = ref<FestivalState | null>(null)
const mode = ref<'merged' | 'per_artist'>('merged')
const creating = ref(false)
const error = ref('')

onMounted(() => {
  const raw = sessionStorage.getItem('gigtape_festival_state')
  if (!raw) {
    router.replace('/festival')
    return
  }
  state.value = JSON.parse(raw) as FestivalState
})

async function submit() {
  if (!state.value) return
  creating.value = true
  error.value = ''
  try {
    const res = await createFestivalPlaylist({
      event_name: state.value.event_name,
      event_date: state.value.event_date,
      mode: mode.value,
      artists: state.value.lineup.map((r) => ({
        artist_ref: r.artist_ref,
        artist_name: r.artist_name,
        include: r.include,
        tracks: r.tracks,
      })),
    })
    sessionStorage.setItem(
      'gigtape_festival_results',
      JSON.stringify(res.results as FestivalResultEntry[]),
    )
    router.push('/festival/result')
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <section class="festival-mode" v-if="state">
    <h1>Choose playlist mode</h1>
    <p>{{ state.event_name }} — {{ state.event_date }}</p>

    <fieldset class="mode-options">
      <label>
        <input type="radio" value="merged" v-model="mode" />
        <strong>One merged playlist</strong>
        <small>All selected artists' tracks in lineup order, one playlist.</small>
      </label>
      <label>
        <input type="radio" value="per_artist" v-model="mode" />
        <strong>One playlist per artist</strong>
        <small>One playlist for each selected artist.</small>
      </label>
    </fieldset>

    <button type="button" :disabled="creating" @click="submit">
      {{ creating ? 'Creating…' : 'Create Playlists' }}
    </button>

    <p v-if="error" class="error">{{ error }}</p>
  </section>
</template>

<style scoped>
.festival-mode {
  max-width: 640px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
.mode-options label {
  display: block;
  padding: 0.5rem 0;
}
.mode-options small {
  display: block;
  color: #555;
  margin-left: 1.5rem;
}
.error {
  color: #b00;
}
</style>
