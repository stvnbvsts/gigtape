<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  clearSessionId,
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
    sessionStorage.setItem(
      'gigtape_festival_result_meta',
      JSON.stringify({ event_name: state.value.event_name, mode: mode.value }),
    )
    clearSessionId()
    router.push('/festival/result')
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <section v-if="state">
    <div class="gt-tape-title gt-tape-title--ink">How do you want it?</div>
    <p class="gt-sub gt-screen-sub">{{ state.event_name }} · {{ state.event_date }}</p>

    <div class="gt-eyebrow gt-section-label">HOW DO YOU WANT IT?</div>
    <div class="gt-mode-row">
      <button
        type="button"
        class="gt-mode"
        :class="{ 'gt-mode--active': mode === 'merged' }"
        @click="mode = 'merged'"
      >
        <span class="gt-mode__title">ONE BIG TAPE</span>
        <span class="gt-mode__sub">everyone, merged</span>
      </button>
      <button
        type="button"
        class="gt-mode"
        :class="{ 'gt-mode--active': mode === 'per_artist' }"
        @click="mode = 'per_artist'"
      >
        <span class="gt-mode__title">A TAPE EACH</span>
        <span class="gt-mode__sub">one per artist</span>
      </button>
    </div>

    <button class="gt-btn gt-btn--block gt-create-btn" type="button" :disabled="creating" @click="submit">
      {{
        creating
          ? 'CREATING…'
          : mode === 'merged'
            ? 'MAKE ONE BIG TAPE →'
            : `MAKE ${state.lineup.filter((r) => r.include).length} TAPES →`
      }}
    </button>
    <p class="gt-attribution gt-center gt-bottom-caption">LINEUPS VIA SETLIST.FM · MAY BE PARTIAL</p>

    <p v-if="error" class="gt-panel gt-screen-message" role="alert">{{ error }}</p>
  </section>
</template>
