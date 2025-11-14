<template>
  <div class="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
    <div class="flex items-start gap-4">
      <!-- Icon Slot -->
      <div :class="['p-3 rounded-lg flex-shrink-0', iconBgClass]">
        <slot name="icon">
          <IconServer :class="['w-6 h-6', iconClass]" />
        </slot>
      </div>

      <div class="flex-1">
        <!-- Header: Title and Status Badge -->
        <div class="flex items-center justify-between mb-2">
          <h4 class="font-semibold text-gray-900">{{ title }}</h4>
          <span :class="['px-2 py-1 text-xs rounded', statusBadgeClass]">
            {{ statusText }}
          </span>
        </div>

        <!-- Description -->
        <p class="text-sm text-gray-600 mb-4">{{ description }}</p>

        <!-- Content Slot (for download button, progress, etc.) -->
        <slot></slot>

        <!-- Download Section (if not installed) -->
        <div v-if="!status.installed && showDownload">
          <!-- Download Progress -->
          <DownloadProgress
            v-if="downloading"
            :progress="downloadProgress"
            :progress-bar-class="progressBarClass"
          />

          <!-- Download Error -->
          <div v-if="downloadError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
            {{ downloadError }}
          </div>

          <!-- Download Button -->
          <button
            v-if="!downloading"
            @click="$emit('download')"
            :class="['px-4 py-2 text-sm text-white rounded-lg transition-colors', downloadButtonClass]"
          >
            {{ downloadButtonText }}
          </button>
        </div>

        <!-- Info Text (if installed but not running) -->
        <div v-else-if="status.installed && !status.running && infoText" class="text-sm text-gray-600">
          <p>{{ infoText }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { IconServer } from '@/components/icons'
import DownloadProgress from './DownloadProgress.vue'

interface ServiceStatus {
  installed: boolean
  running: boolean
  healthy: boolean
}

interface DownloadProgress {
  downloaded: number
  total: number
  percent: number
  status: string
}

interface Props {
  title: string
  description: string
  status: ServiceStatus
  downloading?: boolean
  downloadProgress?: DownloadProgress
  downloadError?: string
  downloadButtonText?: string
  downloadButtonClass?: string
  progressBarClass?: string
  iconBgClass?: string
  iconClass?: string
  showDownload?: boolean
  infoText?: string
}

const props = withDefaults(defineProps<Props>(), {
  downloading: false,
  downloadProgress: () => ({ downloaded: 0, total: 0, percent: 0, status: '' }),
  downloadError: '',
  downloadButtonText: '下载并安装',
  downloadButtonClass: 'bg-blue-500 hover:bg-blue-600',
  progressBarClass: 'bg-blue-600',
  iconBgClass: 'bg-blue-100',
  iconClass: 'text-blue-600',
  showDownload: true,
  infoText: ''
})

defineEmits<{
  download: []
}>()

const statusText = computed(() => {
  if (props.status.healthy) return '运行中'
  if (props.status.running) return '启动中'
  if (props.status.installed) return '已安装'
  return '未安装'
})

const statusBadgeClass = computed(() => {
  if (props.status.healthy) return 'bg-green-100 text-green-800'
  if (props.status.running) return 'bg-yellow-100 text-yellow-800'
  if (props.status.installed) return 'bg-gray-100 text-gray-800'
  return 'bg-orange-100 text-orange-800'
})
</script>
