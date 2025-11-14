<template>
  <IconBase :size="size" :className="className">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
  </IconBase>
</template>

<script setup lang="ts">
import IconBase from './IconBase.vue'

defineProps<{
  size?: string | number
  className?: string
}>()
</script>
