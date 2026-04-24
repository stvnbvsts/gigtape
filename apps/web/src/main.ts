import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

import App from './App.vue'
import ArtistSearchView from './views/ArtistSearchView.vue'
import SetlistPreviewView from './views/SetlistPreviewView.vue'
import PlaylistResultView from './views/PlaylistResultView.vue'
import FestivalSearchView from './views/FestivalSearchView.vue'
import FestivalModeView from './views/FestivalModeView.vue'
import FestivalResultView from './views/FestivalResultView.vue'
import { setOAuthError, setSessionId } from './api/client'

// When /auth/callback redirects back with ?session_id=... or ?oauth_error=...,
// lift the values out of the URL into module-level state and scrub the query.
const url = new URL(window.location.href)
const incomingSession = url.searchParams.get('session_id')
if (incomingSession) {
  setSessionId(incomingSession)
  url.searchParams.delete('session_id')
}
const incomingError = url.searchParams.get('oauth_error')
if (incomingError) {
  setOAuthError(incomingError)
  url.searchParams.delete('oauth_error')
}
if (incomingSession || incomingError) {
  window.history.replaceState({}, '', url.toString())
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: ArtistSearchView },
    { path: '/setlist', component: SetlistPreviewView },
    { path: '/result', component: PlaylistResultView },
    { path: '/festival', component: FestivalSearchView },
    { path: '/festival/mode', component: FestivalModeView },
    { path: '/festival/result', component: FestivalResultView },
  ],
})

createApp(App).use(router).mount('#app')
