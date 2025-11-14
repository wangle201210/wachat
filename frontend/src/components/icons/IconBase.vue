<template>
  <svg
    :class="className"
    :width="size"
    :height="size"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
    xmlns="http://www.w3.org/2000/svg"
  >
    <slot />
  </svg>
</template>

<script setup lang="ts">
defineProps<{
  size?: string | number
  className?: string
}>()
</script>
