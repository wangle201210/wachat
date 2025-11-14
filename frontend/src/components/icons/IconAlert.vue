<template>
  <IconBase :size="size" :className="className">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  </IconBase>
</template>

<script setup lang="ts">
import IconBase from './IconBase.vue'

defineProps<{
  size?: string | number
  className?: string
}>()
</script>
