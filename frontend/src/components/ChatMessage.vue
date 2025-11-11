<template>
  <div>
    <!-- User Message -->
    <div v-if="message.role === 'user'" class="flex justify-end gap-3">
      <div class="bg-blue-500 text-white rounded-lg px-4 py-2">
        {{ message.content }}
      </div>
      <AvatarUser />
    </div>

    <!-- AI Message -->
    <div v-else class="flex gap-3">
      <AvatarAI />
      <div class="flex-1">
        <div class="prose prose-sm prose-zinc max-w-none">
          <NodeRenderer :content="message.content" />
        </div>
        <!-- RAG Documents -->
        <RAGDocuments :documents="message.ragDocuments" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import NodeRenderer from 'vue-renderer-markdown'
import 'katex/dist/katex.min.css'
import '../assets/markdown.css'
import AvatarAI from './AvatarAI.vue'
import AvatarUser from './AvatarUser.vue'
import RAGDocuments from './RAGDocuments.vue'
import type { Message } from '../composables/useChat'

defineProps<{
  message: Message
}>()
</script>
