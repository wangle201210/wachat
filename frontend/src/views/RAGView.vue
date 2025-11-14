<template>
  <div class="flex flex-col h-screen bg-white">
    <!-- Top Navigation Bar -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-gray-50">
      <!-- Back Button -->
      <button
        @click="router.push('/')"
        class="flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded transition-colors"
      >
        <IconArrowLeft class="w-4 h-4" />
        <span>返回聊天</span>
      </button>

      <!-- RAG Settings and Config Button -->
      <div class="flex items-center gap-3">
        <!-- RAG Settings - Show when installed -->
        <div v-if="status.installed" class="flex items-center gap-2 text-sm border-gray-300 pl-3">
          <template v-if="!editingSettings">
            <!-- Display Mode -->
            <div class="flex items-center gap-2 text-gray-600">
              <span class="text-xs">TopK: <strong class="text-gray-900">{{ ragSettings.topK }}</strong></span>
              <span class="text-xs text-gray-300">|</span>
              <span class="text-xs">知识库: <strong class="text-gray-900">{{ ragSettings.defaultKnowledgeBase || '未设置' }}</strong></span>
              <button
                @click="toggleEditSettings"
                class="p-1 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                title="编辑设置"
              >
                <IconEdit class="w-3.5 h-3.5" />
              </button>
            </div>
          </template>
          <template v-else>
            <!-- Edit Mode -->
            <div class="flex items-center gap-2">
              <div class="flex items-center gap-1">
                <label class="text-xs text-gray-600">TopK:</label>
                <input
                  v-model.number="ragSettings.topK"
                  type="number"
                  min="1"
                  max="100"
                  class="w-16 px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div class="flex items-center gap-1">
                <label class="text-xs text-gray-600">知识库:</label>
                <select
                  v-model="ragSettings.defaultKnowledgeBase"
                  :disabled="loadingKnowledgeBases"
                  class="w-40 px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50"
                >
                  <option value="">-- 请选择 --</option>
                  <option v-for="kb in knowledgeBases" :key="kb" :value="kb">{{ kb }}</option>
                </select>
                <span v-if="loadingKnowledgeBases" class="text-xs text-gray-500">加载中...</span>
              </div>
              <button
                @click="saveRAGSettings"
                :disabled="savingSettings"
                class="px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors disabled:opacity-50"
              >
                {{ savingSettings ? '保存中...' : '保存' }}
              </button>
              <button
                @click="toggleEditSettings"
                :disabled="savingSettings"
                class="px-2 py-1 text-xs bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition-colors disabled:opacity-50"
              >
                取消
              </button>
            </div>
          </template>
        </div>

        <!-- Config Button -->
        <button
          v-if="status.installed"
          @click="showConfigEditor = true"
          class="p-1.5 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded transition-colors"
          title="编辑配置"
        >
          <IconSettings class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-hidden">
      <!-- Initial Loading State -->
      <div v-if="initialLoading" class="h-full flex items-center justify-center">
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p class="text-gray-600">正在检查 RAG 服务状态...</p>
        </div>
      </div>

      <!-- Not Installed - Show Download Section -->
      <div v-else-if="!status.installed" class="h-full flex items-center justify-center">
        <div class="text-center max-w-md px-4">
          <IconDownload class="w-20 h-20 text-gray-400 mx-auto mb-4" />
          <h3 class="text-xl font-medium text-gray-900 mb-2">RAG 服务未安装</h3>
          <p class="text-sm text-gray-600 mb-6">
            点击下方按钮下载并安装 go-rag 服务，用于知识库管理和文档检索
          </p>

          <!-- Download Progress -->
          <DownloadProgress
            v-if="downloading"
            :progress="downloadProgress"
          />

          <!-- Error Message -->
          <div v-if="downloadError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
            {{ downloadError }}
          </div>

          <!-- Download Button -->
          <button
            v-if="!downloading"
            @click="handleDownload"
            class="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            <IconDownload class="w-5 h-5 inline-block mr-2" />
            下载并安装
          </button>
        </div>
      </div>

      <!-- Installed but Not Running - Show Service Cards -->
      <div v-else-if="!status.running" class="h-full flex items-center justify-center p-6 overflow-auto">
        <div class="w-full max-w-2xl space-y-6">
          <!-- Qdrant Service Card -->
          <ServiceCard
            title="Qdrant 向量数据库"
            description="Qdrant 是 RAG 服务的向量数据库依赖项，用于存储和检索文档向量"
            :status="qdrantStatus"
            :downloading="downloadingQdrant"
            :download-progress="qdrantDownloadProgress"
            :download-error="qdrantDownloadError"
            download-button-text="下载并安装 Qdrant"
            download-button-class="bg-purple-500 hover:bg-purple-600"
            progress-bar-class="bg-purple-600"
            icon-bg-class="bg-purple-100"
            icon-class="text-purple-600"
            :info-text="qdrantStatus.running ? '✓ Qdrant 正在运行' : 'ℹ️ 启动 RAG 时会自动启动 Qdrant'"
            @download="handleQdrantDownload"
          >
            <template #icon>
              <IconServer class="w-6 h-6 text-purple-600" />
            </template>
          </ServiceCard>

          <!-- RAG Service Card -->
          <div class="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
            <div class="text-center">
              <IconPlay class="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <h3 class="text-xl font-medium text-gray-900 mb-2">RAG 服务未运行</h3>
              <p class="text-sm text-gray-600 mb-6">
                点击下方按钮启动 go-rag 服务，启动后即可进行知识库管理
              </p>

              <!-- Start Progress -->
              <div v-if="starting" class="mb-6">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-2"></div>
                <p class="text-sm text-gray-600">{{ startProgress }}</p>
              </div>

              <!-- Error Message -->
              <div v-if="startError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                {{ startError }}
              </div>

              <!-- Start Button -->
              <button
                v-if="!starting"
                @click="handleStart"
                :disabled="!qdrantStatus.installed"
                class="px-6 py-3 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <IconPlay class="w-5 h-5 inline-block mr-2" />
                启动服务
              </button>

              <!-- Hint -->
              <p v-if="!qdrantStatus.installed" class="mt-4 text-xs text-orange-600">
                ⚠️ 请先下载并安装 Qdrant
              </p>
              <p v-else class="mt-4 text-xs text-gray-500">
                启动 RAG 时会自动检查并启动 Qdrant
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Running but Not Healthy - Show Waiting State -->
      <div v-else-if="!status.healthy" class="h-full flex items-center justify-center">
        <div class="text-center max-w-md px-4">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-yellow-500 mx-auto mb-4"></div>
          <h3 class="text-lg font-medium text-gray-900 mb-2">服务正在启动...</h3>
          <p class="text-sm text-gray-600">请稍候，正在等待服务就绪</p>
          <button
            @click="checkStatus"
            class="mt-4 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded transition-colors"
          >
            刷新状态
          </button>
        </div>
      </div>

      <!-- Healthy - Show iframe -->
      <div v-else class="h-full">
        <iframe
          ref="iframeRef"
          :src="ragServerUrl"
          class="w-full h-full border-0"
          title="RAG Management Interface"
        ></iframe>
      </div>
    </div>

    <!-- Config Editor Modal -->
    <RAGConfigEditor
      v-model="showConfigEditor"
      :is-service-running="status.running"
      @load="handleConfigLoad"
      @save="handleConfigSave"
      ref="configEditorRef"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  GetRAGServerInfo,
  GetRAGStatus,
  DownloadRAG,
  StartRAG,
  StopRAG,
  GetRAGConfig,
  SaveRAGConfig,
  GetQdrantStatus,
  DownloadQdrant,
  GetRAGSettings,
  UpdateRAGSettings,
  GetKnowledgeBases
} from '../../wailsjs/go/main/App'
import { useRuntimeEvents } from '@/composables/useRuntimeEvents'
import {
  IconArrowLeft,
  IconEdit,
  IconSettings,
  IconDownload,
  IconServer,
  IconPlay
} from '@/components/icons'
import DownloadProgress from '@/components/DownloadProgress.vue'
import ServiceCard from '@/components/ServiceCard.vue'
import RAGConfigEditor from '@/components/RAGConfigEditor.vue'

const router = useRouter()
const iframeRef = ref<HTMLIFrameElement | null>(null)
const configEditorRef = ref<InstanceType<typeof RAGConfigEditor> | null>(null)

// State
const initialLoading = ref(true)
const status = ref({ installed: false, running: false, healthy: false })
const ragServerUrl = ref('')

// RAG Settings
const ragSettings = ref({ topK: 5, defaultKnowledgeBase: '' })
const editingSettings = ref(false)
const savingSettings = ref(false)
const knowledgeBases = ref<string[]>([])
const loadingKnowledgeBases = ref(false)

// Qdrant state
const qdrantStatus = ref({ installed: false, running: false, healthy: false })

// Download state
const downloading = ref(false)
const downloadProgress = ref({ downloaded: 0, total: 0, percent: 0, status: '' })
const downloadError = ref('')

// Qdrant download state
const downloadingQdrant = ref(false)
const qdrantDownloadProgress = ref({ downloaded: 0, total: 0, percent: 0, status: '' })
const qdrantDownloadError = ref('')

// Start state
const starting = ref(false)
const startProgress = ref('')
const startError = ref('')

// Config editor state
const showConfigEditor = ref(false)

// Event management
const { onMultiple } = useRuntimeEvents()

// Check RAG status
async function checkStatus() {
  try {
    const ragStatus = await GetRAGStatus()
    status.value = {
      installed: ragStatus.installed || false,
      running: ragStatus.running || false,
      healthy: ragStatus.healthy || false
    }

    if (!ragServerUrl.value && status.value.healthy) {
      const info = await GetRAGServerInfo()
      if (info.enabled && info.url) {
        ragServerUrl.value = info.url
      }
    }
  } catch (err) {
    console.error('Failed to check RAG status:', err)
  }
}

// Check Qdrant status
async function checkQdrantStatus() {
  try {
    const status = await GetQdrantStatus()
    qdrantStatus.value = {
      installed: status.installed || false,
      running: status.running || false,
      healthy: status.healthy || false
    }
  } catch (err) {
    console.error('Failed to check Qdrant status:', err)
  }
}

// Handle download
async function handleDownload() {
  downloading.value = true
  downloadError.value = ''
  downloadProgress.value = { downloaded: 0, total: 0, percent: 0, status: '准备下载...' }

  try {
    await DownloadRAG()
    await checkStatus()
    downloading.value = false
  } catch (err: any) {
    downloadError.value = err.toString()
    downloading.value = false
  }
}

// Handle Qdrant download
async function handleQdrantDownload() {
  downloadingQdrant.value = true
  qdrantDownloadError.value = ''
  qdrantDownloadProgress.value = { downloaded: 0, total: 0, percent: 0, status: '准备下载 Qdrant...' }

  try {
    await DownloadQdrant()
    await checkQdrantStatus()
    downloadingQdrant.value = false
  } catch (err: any) {
    qdrantDownloadError.value = err.toString()
    downloadingQdrant.value = false
  }
}

// Handle start
async function handleStart() {
  starting.value = true
  startError.value = ''
  startProgress.value = '正在启动服务...'

  try {
    await StartRAG()
    await checkStatus()
    starting.value = false
  } catch (err: any) {
    startError.value = err.toString()
    starting.value = false
  }
}

// Config editor handlers
async function handleConfigLoad() {
  try {
    const content = await GetRAGConfig()
    configEditorRef.value?.setConfigContent(content)
  } catch (err: any) {
    configEditorRef.value?.setError('读取配置失败: ' + err.toString())
  }
}

async function handleConfigSave(content: string) {
  const wasRunning = status.value.running

  try {
    await SaveRAGConfig(content)

    if (wasRunning) {
      try {
        await StopRAG()
        await new Promise(resolve => setTimeout(resolve, 2000))
        await StartRAG()
        alert('配置保存成功！\n\nRAG 服务已自动重启，新配置已生效。')
      } catch (restartErr: any) {
        alert('配置保存成功，但重启服务失败！\n\n错误：' + restartErr.toString() + '\n\n请手动重启服务。')
      }
    } else {
      alert('配置保存成功！\n\n下次启动 RAG 服务时将使用新配置。')
    }

    configEditorRef.value?.close()
    await checkStatus()
  } catch (err: any) {
    configEditorRef.value?.setError('保存配置失败: ' + err.toString())
  } finally {
    configEditorRef.value?.setSaving(false)
  }
}

// Load RAG settings
async function loadRAGSettings() {
  try {
    const settings = await GetRAGSettings()
    ragSettings.value = {
      topK: settings.topK || 5,
      defaultKnowledgeBase: settings.defaultKnowledgeBase || ''
    }
  } catch (err: any) {
    console.error('Failed to load RAG settings:', err)
  }
}

// Load knowledge bases
async function loadKnowledgeBases() {
  if (!status.value.healthy) return

  loadingKnowledgeBases.value = true
  try {
    const bases = await GetKnowledgeBases()
    knowledgeBases.value = bases || []
  } catch (err: any) {
    console.error('Failed to load knowledge bases:', err)
    knowledgeBases.value = []
  } finally {
    loadingKnowledgeBases.value = false
  }
}

// Toggle settings editing
async function toggleEditSettings() {
  if (!editingSettings.value) {
    editingSettings.value = true
    await loadKnowledgeBases()
  } else {
    editingSettings.value = false
    loadRAGSettings()
  }
}

// Save RAG settings
async function saveRAGSettings() {
  if (ragSettings.value.topK < 1 || ragSettings.value.topK > 100) {
    alert('TopK 值必须在 1-100 之间')
    return
  }

  savingSettings.value = true
  try {
    await UpdateRAGSettings(ragSettings.value.topK, ragSettings.value.defaultKnowledgeBase)
    alert('设置保存成功！')
    editingSettings.value = false
  } catch (err: any) {
    alert('保存设置失败: ' + err.toString())
  } finally {
    savingSettings.value = false
  }
}

// Lifecycle
onMounted(async () => {
  // Setup event listeners
  onMultiple({
    'rag:download:progress': (data: any) => {
      downloadProgress.value = {
        downloaded: data.downloaded || 0,
        total: data.total || 0,
        percent: data.percent || 0,
        status: data.status || ''
      }
    },
    'rag:download:complete': () => {
      downloading.value = false
      checkStatus()
    },
    'rag:download:error': (data: any) => {
      downloadError.value = data.error || '下载失败'
      downloading.value = false
    },
    'rag:start:progress': (data: any) => {
      startProgress.value = data.status || '正在启动...'
    },
    'rag:start:complete': () => {
      starting.value = false
      checkStatus()
    },
    'rag:start:error': (data: any) => {
      startError.value = data.error || '启动失败'
      starting.value = false
    },
    'qdrant:download:progress': (data: any) => {
      qdrantDownloadProgress.value = {
        downloaded: data.downloaded || 0,
        total: data.total || 0,
        percent: data.percent || 0,
        status: data.status || ''
      }
    },
    'qdrant:download:complete': () => {
      downloadingQdrant.value = false
      checkQdrantStatus()
    },
    'qdrant:download:error': (data: any) => {
      qdrantDownloadError.value = data.error || '下载 Qdrant 失败'
      downloadingQdrant.value = false
    }
  })

  await checkStatus()
  await checkQdrantStatus()
  await loadRAGSettings()
  initialLoading.value = false

  // Poll status
  setInterval(async () => {
    if (!status.value.healthy) await checkStatus()
    await checkQdrantStatus()
  }, 5000)
})
</script>
