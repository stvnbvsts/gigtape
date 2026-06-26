<script setup lang="ts">
import type { Track } from '../api/client'

defineProps<{
  tracks: Track[]
  removed?: Set<number>
  removedIndexes?: number[]
}>()
defineEmits<{ (e: 'toggle', index: number): void; (e: 'remove', index: number): void }>()

function isRemoved(propsRemoved: Set<number> | undefined, propsRemovedIndexes: number[] | undefined, index: number) {
  return propsRemoved?.has(index) || propsRemovedIndexes?.includes(index) || false
}

function numberLabel(
  tracks: Track[],
  propsRemoved: Set<number> | undefined,
  propsRemovedIndexes: number[] | undefined,
  index: number,
) {
  if (isRemoved(propsRemoved, propsRemovedIndexes, index)) return '—'
  const keptBefore = tracks
    .slice(0, index + 1)
    .filter((_, i) => !isRemoved(propsRemoved, propsRemovedIndexes, i)).length
  return `${String(keptBefore).padStart(2, '0')}.`
}
</script>

<template>
  <div class="gt-list" role="list">
    <div
      v-for="(t, i) in tracks"
      :key="`${t.title}-${i}`"
      class="gt-track"
      :class="{ 'gt-track--removed': isRemoved(removed, removedIndexes, i) }"
      role="listitem"
    >
      <span class="gt-track__num">{{ numberLabel(tracks, removed, removedIndexes, i) }}</span>
      <span class="gt-track__title">{{ t.title }}</span>
      <span class="gt-track__dur">—</span>
      <button
        type="button"
        class="gt-track__toggle"
        :aria-label="isRemoved(removed, removedIndexes, i) ? `Restore ${t.title}` : `Remove ${t.title}`"
        @click="$emit('toggle', i); $emit('remove', i)"
      >
        {{ isRemoved(removed, removedIndexes, i) ? '↩' : '✗' }}
      </button>
    </div>
    <slot />
  </div>
</template>
