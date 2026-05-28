<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  createArtistPlaylist,
  getSetlists,
  type Setlist,
  type Track,
} from '../api/client'
import TrackList from '../components/TrackList.vue'

const route = useRoute()
const router = useRouter()

const setlists = ref<Setlist[]>([])
const selectedIdx = ref(0)
const tracks = ref<Track[]>([])
const manualTitle = ref('')
const error = ref('')
const loading = ref(false)
const shortWarning = ref(false)
const creating = ref(false)

const artistRef = (route.query.ref as string) || ''
const artistName = (route.query.name as string) || ''

onMounted(async () => {
  loading.value = true
  try {
    const res = await getSetlists(artistRef, artistName)
    setlists.value = res.setlists
    shortWarning.value = res.short_warning
    if (setlists.value.length > 0) {
      tracks.value = [...setlists.value[0].tracks]
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})

function selectSetlist(i: number) {
  selectedIdx.value = i
  tracks.value = [...setlists.value[i].tracks]
}

function removeTrack(index: number) {
  tracks.value.splice(index, 1)
}

function addManualTrack() {
  const title = manualTitle.value.trim()
  if (!title) return
  tracks.value.push({ title, artist_name: artistName })
  manualTitle.value = ''
}

async function createPlaylist() {
  creating.value = true
  error.value = ''
  try {
    const current = setlists.value[selectedIdx.value]
    const eventDate = current ? current.date : new Date().toISOString().slice(0, 10)
    const result = await createArtistPlaylist({
      artist_ref: artistRef,
      artist_name: artistName,
      setlist_index: selectedIdx.value,
      event_date: eventDate,
      tracks: tracks.value,
    })
    router.push({ path: '/result', state: { result: JSON.parse(JSON.stringify(result)) } })
    sessionStorage.setItem('gigtape_last_result', JSON.stringify(result))
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <section class="setlist-preview">
    <h1>{{ artistName }}</h1>
    <p v-if="loading">Loading setlists…</p>
    <p v-if="error" class="error">{{ error }}</p>

    <template v-if="!loading && setlists.length > 0">
      <div v-if="setlists.length > 1" class="setlist-selector">
        <label>Choose setlist:</label>
        <select :value="selectedIdx" @change="selectSetlist(Number(($event.target as HTMLSelectElement).value))">
          <option v-for="(s, i) in setlists" :key="i" :value="i">
            {{ s.date }} — {{ s.event_name }} ({{ s.track_count }} songs)
          </option>
        </select>
      </div>

      <p class="event-meta">
        {{ setlists[selectedIdx].date }} — {{ setlists[selectedIdx].event_name }}
      </p>
      <!-- setlist.fm attribution is required wherever setlist data appears -->
      <p class="attribution">{{ setlists[selectedIdx].source_attribution }}</p>

      <p v-if="shortWarning" class="warning">
        ⚠ Only {{ setlists[selectedIdx].track_count }} songs — setlist may be incomplete.
      </p>

      <TrackList :tracks="tracks" @remove="removeTrack" />

      <form class="manual-add" @submit.prevent="addManualTrack">
        <input v-model="manualTitle" placeholder="Add a track manually" />
        <button type="submit">Add</button>
      </form>

      <button type="button" :disabled="creating || tracks.length === 0" @click="createPlaylist">
        {{ creating ? 'Creating…' : 'Create Playlist' }}
      </button>
    </template>

    <template v-else-if="!loading">
      <!-- Empty state: FR-009 — allow manual entry when no setlist is available. -->
      <p>No setlists found for this artist.</p>
      <p>Add tracks manually:</p>
      <TrackList :tracks="tracks" @remove="removeTrack" />
      <form class="manual-add" @submit.prevent="addManualTrack">
        <input v-model="manualTitle" placeholder="Track title" />
        <button type="submit">Add</button>
      </form>
      <button type="button" :disabled="creating || tracks.length === 0" @click="createPlaylist">
        {{ creating ? 'Creating…' : 'Create Playlist' }}
      </button>
    </template>
  </section>
</template>

<style scoped>
.setlist-preview {
  max-width: 640px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
.attribution {
  font-size: 0.9em;
  color: #555;
}
.warning {
  color: #a60;
}
.error {
  color: #b00;
}
</style>
