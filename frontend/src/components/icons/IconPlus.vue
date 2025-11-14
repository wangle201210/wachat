<template>
  <IconBase :size="size" :className="className">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
  </IconBase>
</template>

<script setup lang="ts">
import IconBase from './IconBase.vue'

defineProps<{
  size?: string | number
  className?: string
}>()
</script>
