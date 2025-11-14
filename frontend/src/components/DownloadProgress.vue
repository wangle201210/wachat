<template>
  <div class="mb-6">
    <div class="mb-2">
      <div class="flex justify-between text-sm text-gray-600 mb-1">
        <span>{{ progress.status }}</span>
        <span>{{ progress.percent.toFixed(1) }}%</span>
      </div>
      <div class="w-full bg-gray-200 rounded-full h-2.5">
        <div
          :class="['h-2.5 rounded-full transition-all duration-300', progressBarClass]"
          :style="{ width: progress.percent + '%' }"
        ></div>
      </div>
    </div>
    <p class="text-xs text-gray-500">
      {{ formatBytes(progress.downloaded) }} / {{ formatBytes(progress.total) }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { formatBytes } from '@/utils/format'

interface DownloadProgress {
  downloaded: number
  total: number
  percent: number
  status: string
}

interface Props {
  progress: DownloadProgress
  progressBarClass?: string
}

withDefaults(defineProps<Props>(), {
  progressBarClass: 'bg-blue-600'
})
</script>
