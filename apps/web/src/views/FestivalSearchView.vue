<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  getSessionId,
  getSetlists,
  searchEvents,
  type Event,
  type Track,
} from '../api/client'

interface LineupRow {
  artist_ref: string
  artist_name: string
  include: boolean
  tracks: Track[]
  hasSetlist: boolean
  attribution: string
}

const router = useRouter()
const query = ref('')
const events = ref<Event[]>([])
const selectedEventIdx = ref<number | null>(null)
const lineup = ref<LineupRow[]>([])
const manualName = ref('')
const loading = ref(false)
const error = ref('')

async function onSearch() {
  error.value = ''
  if (!getSessionId()) {
    error.value = 'Please connect Spotify first.'
    return
  }
  if (!query.value.trim()) return
  loading.value = true
  try {
    const res = await searchEvents(query.value.trim())
    events.value = res.events
    selectedEventIdx.value = null
    lineup.value = []
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function pickEvent(i: number) {
  selectedEventIdx.value = i
  loading.value = true
  const ev = events.value[i]
  const rows: LineupRow[] = []
  for (const a of ev.artists) {
    try {
      const res = await getSetlists(a.external_ref, a.name)
      const top = res.setlists[0]
      rows.push({
        artist_ref: a.external_ref,
        artist_name: a.name,
        include: true,
        tracks: top ? top.tracks : [],
        hasSetlist: !!top,
        attribution: top ? top.source_attribution : '',
      })
    } catch {
      rows.push({
        artist_ref: a.external_ref,
        artist_name: a.name,
        include: true,
        tracks: [],
        hasSetlist: false,
        attribution: '',
      })
    }
  }
  lineup.value = rows
  loading.value = false
}

function addManual() {
  const name = manualName.value.trim()
  if (!name) return
  lineup.value.push({
    artist_ref: '',
    artist_name: name,
    include: true,
    tracks: [],
    hasSetlist: false,
    attribution: '',
  })
  manualName.value = ''
}

function proceedToMode() {
  const ev = events.value[selectedEventIdx.value!]
  sessionStorage.setItem(
    'gigtape_festival_state',
    JSON.stringify({
      event_name: ev.name,
      event_date: ev.date,
      lineup: lineup.value,
    }),
  )
  router.push('/festival/mode')
}
</script>

<template>
  <section class="festival-search">
    <h1>Festival playlist</h1>

    <form @submit.prevent="onSearch">
      <input v-model="query" placeholder="Festival name (e.g. 'Glastonbury 2024')" />
      <button type="submit" :disabled="loading">Search</button>
    </form>

    <p v-if="error" class="error">{{ error }}</p>

    <ul v-if="events.length && selectedEventIdx === null" class="event-list">
      <li v-for="(e, i) in events" :key="i">
        <button type="button" @click="pickEvent(i)">
          <strong>{{ e.name }}</strong>
          <span> — {{ e.date }} ({{ e.location }}, {{ e.artists.length }} artists)</span>
        </button>
      </li>
    </ul>

    <template v-if="selectedEventIdx !== null">
      <h2>Lineup</h2>
      <ul class="lineup">
        <li v-for="(row, i) in lineup" :key="i">
          <label>
            <input type="checkbox" v-model="row.include" />
            <strong>{{ row.artist_name }}</strong>
            <span v-if="row.hasSetlist"> ({{ row.tracks.length }} songs)</span>
            <span v-else class="no-setlist"> (no setlist found)</span>
          </label>
          <div v-if="row.attribution" class="attribution">{{ row.attribution }}</div>
        </li>
      </ul>

      <form class="manual-artist" @submit.prevent="addManual">
        <input v-model="manualName" placeholder="Add artist manually" />
        <button type="submit">Add</button>
      </form>

      <button type="button" @click="proceedToMode">Continue</button>
    </template>
  </section>
</template>

<style scoped>
.festival-search {
  max-width: 680px;
  margin: 2rem auto;
  font-family: system-ui, sans-serif;
}
.attribution {
  font-size: 0.85em;
  color: #555;
  margin-left: 1.5rem;
}
.no-setlist {
  color: #a60;
}
.error {
  color: #b00;
}
</style>
