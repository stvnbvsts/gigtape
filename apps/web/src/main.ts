import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

import App from './App.vue'
import './styles/gigtape.css'
import ArtistSearchView from './views/ArtistSearchView.vue'
import LandingView from './views/LandingView.vue'
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
  if (url.pathname === '/') {
    url.pathname = '/search'
  }
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
    { path: '/', component: LandingView, meta: { step: '00 / START' } },
    { path: '/search', component: ArtistSearchView, meta: { step: '01 / SEARCH', back: '/' } },
    { path: '/setlist', component: SetlistPreviewView, meta: { step: '02 / EDIT', back: '/search' } },
    { path: '/result', component: PlaylistResultView, meta: { step: '03 / PLAY', back: '/' } },
    { path: '/festival', component: FestivalSearchView, meta: { step: 'FEST / FIND', back: '/' } },
    { path: '/festival/mode', component: FestivalModeView, meta: { step: 'FEST / MIX', back: '/festival' } },
    { path: '/festival/result', component: FestivalResultView, meta: { step: 'FEST / DONE', back: '/festival' } },
  ],
})

createApp(App).use(router).mount('#app')
