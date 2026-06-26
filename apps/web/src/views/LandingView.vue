<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { consumeOAuthError, getAuthUrl } from '../api/client'

const router = useRouter()
const authUrl = ref('')
const error = ref('')
const oauthBanner = ref('')

onMounted(() => {
  const code = consumeOAuthError()
  if (code) {
    oauthBanner.value =
      code === 'profile_error'
        ? 'We could not retrieve your Spotify profile. Please try again.'
        : 'OAuth handshake failed. Please try connecting your Spotify account again.'
  }
  getAuthUrl()
    .then((r) => (authUrl.value = r.auth_url))
    .catch(() => {})
})

async function connectSpotify() {
  oauthBanner.value = ''
  if (authUrl.value) {
    window.location.href = authUrl.value
    return
  }
  try {
    const { auth_url } = await getAuthUrl()
    window.location.href = auth_url
  } catch (e) {
    error.value = (e as Error).message
  }
}
</script>

<template>
  <section>
    <p v-if="oauthBanner" class="gt-panel gt-screen-message" role="alert">{{ oauthBanner }}</p>
    <p v-if="error" class="gt-panel gt-screen-message" role="alert">{{ error }}</p>

    <div class="gt-cassette gt-hero-cassette">
      <div class="gt-cassette__label">
        <div class="gt-cassette-head">
          <span class="gt-cassette__name">Gigtape</span>
          <span class="gt-side">SIDE A</span>
        </div>
        <div class="gt-cassette__divider"></div>
        <div class="gt-cassette__reels">
          <div class="gt-reel"><div class="gt-reel__hub"></div></div>
          <div class="gt-reel gt-reel--fast"><div class="gt-reel__hub"></div></div>
        </div>
      </div>
    </div>

    <h1 class="gt-h1 gt-landing-title">Make the tape<br />before the show.</h1>
    <p class="gt-sub">
      Turn any band's live setlist into a Spotify playlist — a mixtape for the gig you're about
      to see. ♪
    </p>

    <button class="gt-btn gt-connect-btn" type="button" @click="connectSpotify">
      <span class="gt-spotify-dot"></span>
      Connect Spotify
    </button>

    <button class="gt-link gt-link--hand gt-plain-link gt-festival-link" type="button" @click="router.push('/festival')">
      Going to a festival? Make the whole lineup →
    </button>

    <div class="gt-footer-note">
      NO ACCOUNT STORED · NOTHING SAVED · IN &amp; OUT<br />
      SETLIST DATA PROVIDED BY SETLIST.FM
    </div>
  </section>
</template>
