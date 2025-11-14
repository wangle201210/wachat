<template>
  <IconBase :size="size" :className="className">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
  </IconBase>
</template>

<script setup lang="ts">
import IconBase from './IconBase.vue'

defineProps<{
  size?: string | number
  className?: string
}>()
</script>
