import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

import App from './App.vue'
import ArtistSearchView from './views/ArtistSearchView.vue'
import SetlistPreviewView from './views/SetlistPreviewView.vue'
import PlaylistResultView from './views/PlaylistResultView.vue'
import FestivalSearchView from './views/FestivalSearchView.vue'
import FestivalModeView from './views/FestivalModeView.vue'
import FestivalResultView from './views/FestivalResultView.vue'
import { setSessionId } from './api/client'

// When /auth/callback redirects back with ?session_id=..., lift the value out of
// the URL into the module-level session store used by the API client.
const url = new URL(window.location.href)
const incoming = url.searchParams.get('session_id')
if (incoming) {
  setSessionId(incoming)
  url.searchParams.delete('session_id')
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
