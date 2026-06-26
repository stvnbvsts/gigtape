<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const stepTag = computed(() => String(route.meta.step || ''))
const backTarget = computed(() => route.meta.back as string | undefined)

function goBack() {
  if (backTarget.value) router.push(backTarget.value)
}
</script>

<template>
  <main class="gt-desk">
    <div class="gt-sheet">
      <div class="gt-grain"></div>
      <header class="gt-topbar">
        <button
          type="button"
          class="gt-topbar__back gt-topbar-button"
          :class="{ 'gt-topbar__back--hidden': !backTarget }"
          @click="goBack"
        >
          ← back
        </button>
        <button type="button" class="gt-wordmark gt-topbar-button" @click="router.push('/')">
          GIGTAPE
        </button>
        <div class="gt-step">{{ stepTag }}</div>
      </header>
      <div class="gt-rule-dashed"></div>
      <router-view v-slot="{ Component }">
        <section class="gt-content">
          <component :is="Component" />
        </section>
      </router-view>
    </div>
  </main>
</template>

<style>
html,
body,
#app {
  min-height: 100%;
  margin: 0;
}

body {
  margin: 0;
}

button,
input {
  font: inherit;
}

.gt-topbar-button {
  border: 0;
  background: transparent;
  padding: 0;
}
</style>
