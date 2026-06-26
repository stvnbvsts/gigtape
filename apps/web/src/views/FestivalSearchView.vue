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
  <section>
    <div class="gt-tape-title gt-tape-title--ink">Festival mode</div>
    <p class="gt-sub gt-screen-sub">One big tape for the whole lineup.</p>

    <form class="gt-row" @submit.prevent="onSearch">
      <input v-model="query" class="gt-input" placeholder="festival or venue…" />
      <button class="gt-btn gt-btn--sm" type="submit" :disabled="loading">GO</button>
    </form>

    <p v-if="error" class="gt-panel gt-screen-message" role="alert">{{ error }}</p>
    <p v-if="loading" class="gt-loading">Digging through recent shows…</p>

    <div v-if="events.length && selectedEventIdx === null" class="gt-search-results">
      <div class="gt-eyebrow">{{ events.length }} EVENTS</div>
      <button
        v-for="(e, i) in events"
        :key="i"
        type="button"
        class="gt-result-card"
        @click="pickEvent(i)"
      >
        <span>
          <span class="gt-result-card__name">{{ e.name }}</span>
          <span class="gt-result-card__meta">
            {{ e.date }} · {{ e.location }} · {{ e.artists.length }} artists
          </span>
        </span>
        <span class="gt-result-arrow">→</span>
      </button>
    </div>

    <template v-if="selectedEventIdx !== null">
      <div class="gt-festival-event">
        <div class="gt-event-name">{{ events[selectedEventIdx].name }}</div>
        <div class="gt-event-meta">
          {{ events[selectedEventIdx].date }} · {{ events[selectedEventIdx].location }} ·
          {{ lineup.filter((r) => r.include).length }} of {{ lineup.length }} artists in
        </div>
      </div>

      <div>
        <div
          v-for="(row, i) in lineup"
          :key="i"
          class="gt-artist-block"
        >
          <button
            type="button"
            class="gt-artist"
            :class="{ 'gt-artist--off': !row.include }"
            @click="row.include = !row.include"
          >
            <span class="gt-check" :class="{ 'gt-check--on': row.include }">{{ row.include ? '✓' : '' }}</span>
            <span class="gt-artist__name">{{ row.artist_name }}</span>
            <span
              class="gt-artist__meta"
              :class="{ 'gt-artist__meta--nodata': !row.hasSetlist && row.tracks.length === 0 }"
            >
              <template v-if="row.hasSetlist">{{ row.tracks.length }} songs</template>
              <template v-else-if="row.tracks.length">added</template>
              <template v-else>no setlist yet</template>
            </span>
          </button>

          <div v-if="row.tracks.length" class="gt-mini-tracks">
            <div v-for="(track, trackIdx) in row.tracks" :key="trackIdx" class="gt-mini-track">
              <span>{{ track.title }}</span>
              <button type="button" @click="removeTrack(row, trackIdx)">✗</button>
            </div>
          </div>
          <form class="gt-row gt-manual-song" @submit.prevent="addTrack(row)">
            <input v-model="row.draft_title" class="gt-input gt-input--hand" :placeholder="`+ add song for ${row.artist_name}`" />
            <button class="gt-track-add" type="submit">ADD</button>
          </form>
        </div>
      </div>

      <form class="gt-row gt-manual-artist" @submit.prevent="addManual">
        <input v-model="manualName" class="gt-input gt-input--hand" placeholder="+ add an artist they missed…" />
        <button class="gt-track-add" type="submit">ADD</button>
      </form>

      <button class="gt-btn gt-btn--block gt-create-btn" type="button" @click="proceedToMode">
        CONTINUE →
      </button>
      <p class="gt-attribution gt-center gt-bottom-caption">LINEUPS VIA SETLIST.FM · MAY BE PARTIAL</p>
    </template>

    <template v-else-if="searched && !loading && events.length === 0">
      <p class="gt-empty-note">No events found.</p>
      <button class="gt-btn gt-btn--block" type="button" @click="startManualFestival">BUILD MANUALLY →</button>
    </template>
  </section>
</template>
