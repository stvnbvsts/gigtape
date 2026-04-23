<script setup lang="ts">
import type { Track } from '../api/client'

defineProps<{ tracks: Track[] }>()
defineEmits<{ (e: 'remove', index: number): void }>()
</script>

<template>
  <ol class="track-list">
    <li v-for="(t, i) in tracks" :key="i">
      <span>{{ t.title }}</span>
      <button type="button" @click="$emit('remove', i)">Remove</button>
    </li>
  </ol>
</template>

<style scoped>
.track-list li {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
}
</style>
