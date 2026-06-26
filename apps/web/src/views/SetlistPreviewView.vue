<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  clearSessionId,
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
const removedIndexes = ref<number[]>([])
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
      removedIndexes.value = []
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
  removedIndexes.value = []
}

function toggleTrack(index: number) {
  if (removedIndexes.value.includes(index)) {
    removedIndexes.value = removedIndexes.value.filter((i) => i !== index)
  } else {
    removedIndexes.value = [...removedIndexes.value, index]
  }
}

function addManualTrack() {
  const title = manualTitle.value.trim()
  if (!title) return
  tracks.value.push({ title, artist_name: artistName })
  manualTitle.value = ''
}

const keptTracks = computed(() => tracks.value.filter((_, i) => !removedIndexes.value.includes(i)))
const keptCount = computed(() => keptTracks.value.length)
const totalTime = computed(() => '00:00')
const currentSetlist = computed(() => setlists.value[selectedIdx.value])

function formatLocation(s: Setlist) {
  return s.event_name || 'recent show'
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
      tracks: keptTracks.value,
    })
    clearSessionId()
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
  <section>
    <div class="gt-tape-title">{{ artistName || 'Setlist' }}</div>
    <p class="gt-sub gt-screen-sub">songs they've been playing live ♪</p>

    <p v-if="loading" class="gt-loading">Loading setlists…</p>
    <p v-if="error" class="gt-panel gt-screen-message" role="alert">{{ error }}</p>

    <template v-if="!loading && setlists.length > 0">
      <div class="gt-eyebrow gt-section-label">PICK A SHOW</div>
      <div class="gt-tabs">
        <button
          v-for="(s, i) in setlists"
          :key="`${s.date}-${i}`"
          type="button"
          class="gt-tab"
          :class="{ 'gt-tab--active': selectedIdx === i }"
          @click="selectSetlist(i)"
        >
          <span class="gt-tab__title">{{ s.event_name || 'Recent show' }}</span>
          <span class="gt-tab__meta">{{ s.date }} · {{ formatLocation(s) }}</span>
        </button>
      </div>

      <!-- setlist.fm attribution is required wherever setlist data appears -->
      <p class="gt-attribution gt-tab-caption">
        {{ currentSetlist?.source_attribution || 'via setlist.fm' }} · {{ setlists.length }} recent shows
      </p>

      <div v-if="shortWarning" class="gt-panel gt-warning-row">
        <span aria-hidden="true">⚠</span>
        <span>
          Only {{ currentSetlist?.track_count || tracks.length }} songs logged for this show — the setlist
          may be incomplete. Add any you remember.
        </span>
      </div>

      <TrackList :tracks="tracks" :removed-indexes="removedIndexes" @toggle="toggleTrack">
        <form class="gt-track gt-manual-track" @submit.prevent="addManualTrack">
          <input v-model="manualTitle" class="gt-input gt-input--ghost" placeholder="+ add a song you remember…" />
          <button class="gt-track-add" type="submit">ADD</button>
        </form>
      </TrackList>

      <div class="gt-track-footer">
        <span>{{ keptCount }} SONGS · {{ totalTime }}</span>
        <span>SIDE A</span>
      </div>

      <button class="gt-btn gt-btn--block gt-create-btn" type="button" :disabled="creating || keptCount === 0" @click="createPlaylist">
        {{ creating ? 'CREATING…' : 'MAKE THE TAPE →' }}
      </button>
      <p class="gt-attribution gt-center gt-bottom-caption">SETLIST DATA PROVIDED BY SETLIST.FM</p>
    </template>

    <template v-else-if="!loading">
      <!-- Empty state: FR-009 — allow manual entry when no setlist is available. -->
      <div class="gt-panel gt-warning-row">
        <span aria-hidden="true">⚠</span>
        <span>No setlists found for this artist. Add the songs you remember.</span>
      </div>
      <TrackList :tracks="tracks" :removed-indexes="removedIndexes" @toggle="toggleTrack">
        <form class="gt-track gt-manual-track" @submit.prevent="addManualTrack">
          <input v-model="manualTitle" class="gt-input gt-input--ghost" placeholder="+ add a song you remember…" />
          <button class="gt-track-add" type="submit">ADD</button>
        </form>
      </TrackList>
      <div class="gt-track-footer">
        <span>{{ keptCount }} SONGS · {{ totalTime }}</span>
        <span>SIDE A</span>
      </div>
      <button class="gt-btn gt-btn--block gt-create-btn" type="button" :disabled="creating || keptCount === 0" @click="createPlaylist">
        {{ creating ? 'CREATING…' : 'MAKE THE TAPE →' }}
      </button>
    </template>
  </section>
</template>
