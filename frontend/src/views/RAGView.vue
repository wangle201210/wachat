<template>
  <div class="flex flex-col h-screen bg-white">
    <!-- Top Navigation Bar -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-gray-50">
      <!-- Back Button -->
      <button
        @click="router.push('/')"
        class="flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded transition-colors"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        <span>返回聊天</span>
      </button>

      <!-- Title -->
      <h1 class="text-lg font-semibold text-gray-900">知识库管理</h1>

      <!-- Spacer for balance -->
      <div class="w-24"></div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-hidden">
      <!-- Loading State -->
      <div v-if="loading" class="h-full flex items-center justify-center">
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p class="text-gray-600">正在加载 RAG 管理界面...</p>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="h-full flex items-center justify-center">
        <div class="text-center max-w-md px-4">
          <svg class="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 class="text-lg font-medium text-gray-900 mb-2">RAG 服务未启用</h3>
          <p class="text-sm text-gray-600 mb-4">{{ error }}</p>
          <p class="text-xs text-gray-500 mb-4">
            请在配置文件中启用 RAG 服务并配置 server.address
          </p>
          <button
            @click="router.push('/')"
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            返回聊天
          </button>
        </div>
      </div>

      <!-- RAG Web Interface (iframe) -->
      <iframe
        v-else-if="ragServerUrl"
        :src="ragServerUrl"
        class="w-full h-full border-0"
        title="RAG Management Interface"
      ></iframe>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { GetRAGServerInfo } from '../../wailsjs/go/main/App'

const router = useRouter()
const loading = ref(true)
const error = ref('')
const ragServerUrl = ref('')

onMounted(async () => {
  try {
    const info = await GetRAGServerInfo()

    if (!info.enabled) {
      error.value = 'RAG 服务未在配置中启用'
      loading.value = false
      return
    }

    if (!info.url) {
      error.value = 'RAG 服务 URL 未配置'
      loading.value = false
      return
    }

    ragServerUrl.value = info.url
    loading.value = false
  } catch (err) {
    error.value = `加载 RAG 服务信息失败: ${err}`
    loading.value = false
  }
})
</script>
