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
  draft_title: string
}

const router = useRouter()
const query = ref('')
const events = ref<Event[]>([])
const selectedEventIdx = ref<number | null>(null)
const lineup = ref<LineupRow[]>([])
const manualName = ref('')
const loading = ref(false)
const error = ref('')
const searched = ref(false)

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
    searched.value = true
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
        draft_title: '',
      })
    } catch {
      rows.push({
        artist_ref: a.external_ref,
        artist_name: a.name,
        include: true,
        tracks: [],
        hasSetlist: false,
        attribution: '',
        draft_title: '',
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
    draft_title: '',
  })
  manualName.value = ''
}

function addTrack(row: LineupRow) {
  const title = row.draft_title.trim()
  if (!title) return
  row.tracks.push({ title, artist_name: row.artist_name })
  row.draft_title = ''
}

function removeTrack(row: LineupRow, index: number) {
  row.tracks.splice(index, 1)
}

function startManualFestival() {
  const name = query.value.trim() || 'Manual festival'
  const today = new Date().toISOString().slice(0, 10)
  events.value = [
    {
      name,
      date: today,
      location: 'manual',
      artists: [],
      lineup_complete: false,
    },
  ]
  selectedEventIdx.value = 0
  lineup.value = []
  searched.value = true
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
            <span v-else-if="row.tracks.length"> ({{ row.tracks.length }} manual songs)</span>
            <span v-else class="no-setlist"> (no setlist found)</span>
          </label>
          <div v-if="row.attribution" class="attribution">{{ row.attribution }}</div>
          <ol v-if="row.tracks.length" class="track-editor">
            <li v-for="(track, trackIdx) in row.tracks" :key="trackIdx">
              <span>{{ track.title }}</span>
              <button type="button" @click="removeTrack(row, trackIdx)">Remove</button>
            </li>
          </ol>
          <form class="manual-track" @submit.prevent="addTrack(row)">
            <input v-model="row.draft_title" :placeholder="`Add song for ${row.artist_name}`" />
            <button type="submit">Add song</button>
          </form>
        </li>
      </ul>

      <form class="manual-artist" @submit.prevent="addManual">
        <input v-model="manualName" placeholder="Add artist manually" />
        <button type="submit">Add</button>
      </form>

      <button type="button" @click="proceedToMode">Continue</button>
    </template>

    <template v-else-if="searched && !loading && events.length === 0">
      <p>No events found.</p>
      <button type="button" @click="startManualFestival">Build manually</button>
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
.track-editor {
  margin-left: 1.5rem;
}
.track-editor li {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
.manual-track {
  margin: 0.4rem 0 0.8rem 1.5rem;
}
.error {
  color: #b00;
}
</style>
