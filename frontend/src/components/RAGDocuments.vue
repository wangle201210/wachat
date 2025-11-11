<template>
  <div v-if="documents && documents.length > 0" class="mt-3 border-t border-gray-200 pt-3">
    <button
      @click="isExpanded = !isExpanded"
      class="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 transition-colors"
    >
      <svg
        :class="['w-4 h-4 transition-transform', isExpanded ? 'rotate-90' : '']"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
      <span class="font-medium">检索到 {{ documents.length }} 个相关文档</span>
    </button>

    <transition
      enter-active-class="transition-all duration-200 ease-out"
      leave-active-class="transition-all duration-150 ease-in"
      enter-from-class="max-h-0 opacity-0"
      leave-to-class="max-h-0 opacity-0"
      enter-to-class="max-h-[500px] opacity-100"
      leave-from-class="max-h-[500px] opacity-100"
    >
      <div v-if="isExpanded" class="mt-3 space-y-2 overflow-hidden">
        <div
          v-for="(doc, index) in documents"
          :key="doc.id || index"
          class="p-3 bg-blue-50 rounded-lg border border-blue-100"
        >
          <div class="flex items-start justify-between gap-2 mb-1">
            <span class="text-xs font-medium text-blue-700">文档 {{ index + 1 }}</span>
            <span v-if="getScore(doc)" class="text-xs text-blue-600">
              相关度: {{ (getScore(doc) * 100).toFixed(1) }}%
            </span>
          </div>
          <p class="text-sm text-gray-700 whitespace-pre-wrap break-words">
            {{ truncateContent(doc.content, 200) }}
          </p>
          <div v-if="doc.meta_data" class="mt-2 pt-2 border-t border-blue-200">
            <div class="flex flex-wrap gap-2">
              <span
                v-for="(value, key) in displayableMetadata(doc.meta_data)"
                :key="key"
                class="text-xs px-2 py-1 bg-blue-100 text-blue-700 rounded"
              >
                {{ key }}: {{ value }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { RAGDocument } from '../composables/useChat'

interface Props {
  documents?: RAGDocument[]
}

const props = defineProps<Props>()
const isExpanded = ref(false)

function getScore(doc: RAGDocument): number | null {
  if (!doc.meta_data) return null
  // Try to get score from metadata
  const score = doc.meta_data.score || doc.meta_data.Score
  return typeof score === 'number' ? score : null
}

function truncateContent(content: string, maxLength: number): string {
  if (content.length <= maxLength) return content
  return content.substring(0, maxLength) + '...'
}

function displayableMetadata(metadata: Record<string, any>): Record<string, any> {
  // Filter out internal metadata and score (already displayed separately)
  const filtered: Record<string, any> = {}
  for (const [key, value] of Object.entries(metadata)) {
    if (key !== 'score' && key !== 'Score' && typeof value !== 'object') {
      filtered[key] = value
    }
  }
  return filtered
}
</script>
