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
<!--      <h1 class="text-lg font-semibold text-gray-900">知识库管理</h1>-->

      <!-- Status Badge and Settings -->
      <div class="flex items-center gap-3">
        <!-- Status Badge -->
<!--        <div class="flex items-center gap-2">-->
<!--          <span v-if="status.healthy" class="px-2 py-1 text-xs bg-green-100 text-green-800 rounded">运行中</span>-->
<!--          <span v-else-if="status.running" class="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded">启动中</span>-->
<!--          <span v-else-if="status.installed" class="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded">已安装</span>-->
<!--          <span v-else class="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded">未安装</span>-->
<!--        </div>-->

        <!-- RAG Settings - Show when installed -->
        <div v-if="status.installed" class="flex items-center gap-2 text-sm  border-gray-300 pl-3">
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
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
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
                  <option v-for="kb in knowledgeBases" :key="kb" :value="kb">
                    {{ kb }}
                  </option>
                </select>
                <span v-if="loadingKnowledgeBases" class="text-xs text-gray-500">加载中...</span>
              </div>
              <button
                @click="saveRAGSettings"
                :disabled="savingSettings"
                class="px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors disabled:opacity-50"
                title="保存"
              >
                {{ savingSettings ? '保存中...' : '保存' }}
              </button>
              <button
                @click="toggleEditSettings"
                :disabled="savingSettings"
                class="px-2 py-1 text-xs bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition-colors disabled:opacity-50"
                title="取消"
              >
                取消
              </button>
            </div>
          </template>
        </div>

        <!-- Config Button - Only show when installed -->
        <button
          v-if="status.installed"
          @click="openConfigEditor"
          class="p-1.5 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded transition-colors"
          title="编辑配置"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
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

      <!-- Not Installed - Show Download Button -->
      <div v-else-if="!status.installed" class="h-full flex items-center justify-center">
        <div class="text-center max-w-md px-4">
          <svg class="w-20 h-20 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
          </svg>
          <h3 class="text-xl font-medium text-gray-900 mb-2">RAG 服务未安装</h3>
          <p class="text-sm text-gray-600 mb-6">
            点击下方按钮下载并安装 go-rag 服务，用于知识库管理和文档检索
          </p>

          <!-- Download Progress -->
          <div v-if="downloading" class="mb-6">
            <div class="mb-2">
              <div class="flex justify-between text-sm text-gray-600 mb-1">
                <span>{{ downloadProgress.status }}</span>
                <span>{{ downloadProgress.percent.toFixed(1) }}%</span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2.5">
                <div
                  class="bg-blue-600 h-2.5 rounded-full transition-all duration-300"
                  :style="{ width: downloadProgress.percent + '%' }"
                ></div>
              </div>
            </div>
            <p class="text-xs text-gray-500">
              {{ formatBytes(downloadProgress.downloaded) }} / {{ formatBytes(downloadProgress.total) }}
            </p>
          </div>

          <!-- Error Message -->
          <div v-if="downloadError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
            {{ downloadError }}
          </div>

          <!-- Download Button -->
          <button
            v-if="!downloading"
            @click="handleDownload"
            :disabled="downloading"
            class="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg class="w-5 h-5 inline-block mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            下载并安装
          </button>
        </div>
      </div>

      <!-- Installed but Not Running - Show Start Button -->
      <div v-else-if="!status.running" class="h-full flex items-center justify-center p-6 overflow-auto">
        <div class="w-full max-w-2xl space-y-6">
          <!-- Qdrant Dependency Card -->
          <div class="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
            <div class="flex items-start gap-4">
              <div class="p-3 bg-purple-100 rounded-lg flex-shrink-0">
                <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                </svg>
              </div>
              <div class="flex-1">
                <div class="flex items-center justify-between mb-2">
                  <h4 class="font-semibold text-gray-900">Qdrant 向量数据库</h4>
                  <span v-if="qdrantStatus.healthy" class="px-2 py-1 text-xs bg-green-100 text-green-800 rounded">运行中</span>
                  <span v-else-if="qdrantStatus.running" class="px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded">启动中</span>
                  <span v-else-if="qdrantStatus.installed" class="px-2 py-1 text-xs bg-gray-100 text-gray-800 rounded">已安装</span>
                  <span v-else class="px-2 py-1 text-xs bg-orange-100 text-orange-800 rounded">未安装</span>
                </div>
                <p class="text-sm text-gray-600 mb-4">
                  Qdrant 是 RAG 服务的向量数据库依赖项，用于存储和检索文档向量
                </p>

                <!-- Qdrant Download Button -->
                <div v-if="!qdrantStatus.installed">
                  <div v-if="downloadingQdrant" class="mb-4">
                    <div class="flex justify-between text-sm text-gray-600 mb-1">
                      <span>{{ qdrantDownloadProgress.status }}</span>
                      <span>{{ qdrantDownloadProgress.percent.toFixed(1) }}%</span>
                    </div>
                    <div class="w-full bg-gray-200 rounded-full h-2">
                      <div class="bg-purple-600 h-2 rounded-full transition-all duration-300" :style="{ width: qdrantDownloadProgress.percent + '%' }"></div>
                    </div>
                  </div>
                  <div v-if="qdrantDownloadError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                    {{ qdrantDownloadError }}
                  </div>
                  <button
                    v-if="!downloadingQdrant"
                    @click="handleQdrantDownload"
                    class="px-4 py-2 text-sm bg-purple-500 text-white rounded-lg hover:bg-purple-600 transition-colors"
                  >
                    下载并安装 Qdrant
                  </button>
                </div>

                <!-- Qdrant Info -->
                <div v-else class="text-sm text-gray-600">
                  <p v-if="qdrantStatus.running">✓ Qdrant 正在运行</p>
                  <p v-else>ℹ️ 启动 RAG 时会自动启动 Qdrant</p>
                </div>
              </div>
            </div>
          </div>

          <!-- RAG Service Card -->
          <div class="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
            <div class="text-center">
              <svg class="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
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
                :disabled="starting || (!qdrantStatus.installed && !starting)"
                class="px-6 py-3 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg class="w-5 h-5 inline-block mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
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
    <div
      v-if="showConfigEditor"
      class="fixed inset-0 bg-gray-900 bg-opacity-60 backdrop-blur-sm flex items-center justify-center z-50"
      @click.self="closeConfigEditor"
    >
      <div class="bg-white rounded-xl shadow-2xl w-full max-w-5xl max-h-[90vh] flex flex-col">
        <!-- Modal Header -->
        <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between bg-gradient-to-r from-blue-50 to-indigo-50">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-blue-500 rounded-lg">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <div>
              <h2 class="text-xl font-semibold text-gray-900">编辑 RAG 配置</h2>
              <p class="text-sm text-gray-600">配置文件路径: ~/.wachat/go-rag/config.yaml</p>
            </div>
          </div>
          <button
            @click="closeConfigEditor"
            class="text-gray-400 hover:text-gray-600 transition-colors p-2 hover:bg-white rounded-lg"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Modal Body -->
        <div class="flex-1 overflow-auto p-6 bg-gray-50">
          <div v-if="configLoading" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
              <p class="text-gray-600">正在加载配置...</p>
            </div>
          </div>

          <div v-else>
            <!-- Config Editor -->
            <div class="relative">
              <div class="absolute top-3 right-3 text-xs text-gray-500 bg-gray-200 px-2 py-1 rounded z-10">YAML</div>
              <textarea
                v-model="configContent"
                class="w-full h-96 font-mono text-sm p-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none bg-white"
                placeholder="配置内容..."
                spellcheck="false"
              ></textarea>
            </div>

            <div v-if="configError" class="mt-4 p-4 bg-red-50 border-l-4 border-red-500 rounded text-sm text-red-700 flex items-start gap-3">
              <svg class="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{{ configError }}</span>
            </div>

            <div class="mt-4 p-4 bg-blue-50 border-l-4 border-blue-500 rounded text-sm text-blue-900">
              <p class="font-medium mb-2 flex items-center gap-2">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                提示
              </p>
              <ul class="list-disc list-inside space-y-1 ml-7">
                <li>请确保 YAML 格式正确</li>
                <li>如果 RAG 服务正在运行，保存后会自动重启以应用新配置</li>
                <li>配置文件会自动备份，保存失败时会自动恢复</li>
              </ul>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-6 py-4 border-t border-gray-200 flex items-center justify-between bg-white">
          <div class="text-sm text-gray-500">
            <span v-if="status.running" class="flex items-center gap-2">
              <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              服务运行中，保存后将自动重启
            </span>
          </div>
          <div class="flex items-center gap-3">
            <button
              @click="closeConfigEditor"
              class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
            >
              取消
            </button>
            <button
              @click="saveConfig"
              :disabled="configSaving"
              class="px-6 py-2 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-sm flex items-center gap-2"
            >
              <svg v-if="configSaving" class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4" />
              </svg>
              {{ configSaving ? '保存中...' : '保存配置' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
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

const router = useRouter()
const iframeRef = ref<HTMLIFrameElement | null>(null)

// State
const initialLoading = ref(true)
const status = ref({
  installed: false,
  running: false,
  healthy: false
})
const ragServerUrl = ref('')

// RAG Settings
const ragSettings = ref({
  topK: 5,
  defaultKnowledgeBase: ''
})
const editingSettings = ref(false)
const savingSettings = ref(false)
const knowledgeBases = ref<string[]>([])
const loadingKnowledgeBases = ref(false)

// Qdrant state
const qdrantStatus = ref({
  installed: false,
  running: false,
  healthy: false
})

// Download state
const downloading = ref(false)
const downloadProgress = ref({
  downloaded: 0,
  total: 0,
  percent: 0,
  status: ''
})
const downloadError = ref('')

// Qdrant download state
const downloadingQdrant = ref(false)
const qdrantDownloadProgress = ref({
  downloaded: 0,
  total: 0,
  percent: 0,
  status: ''
})
const qdrantDownloadError = ref('')

// Start state
const starting = ref(false)
const startProgress = ref('')
const startError = ref('')

// Config editor state
const showConfigEditor = ref(false)
const configContent = ref('')
const configLoading = ref(false)
const configSaving = ref(false)
const configError = ref('')

// Event listeners
let cleanupFunctions: (() => void)[] = []

// Format bytes to human readable
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// Check RAG status
async function checkStatus() {
  try {
    const ragStatus = await GetRAGStatus()
    status.value = {
      installed: ragStatus.installed || false,
      running: ragStatus.running || false,
      healthy: ragStatus.healthy || false
    }

    // Get server URL if needed
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
    // Download complete, check status
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
    // Download complete, check status
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
    // Start complete, check status
    await checkStatus()
    starting.value = false
  } catch (err: any) {
    startError.value = err.toString()
    starting.value = false
  }
}

// Open config editor
async function openConfigEditor() {
  showConfigEditor.value = true
  configLoading.value = true
  configError.value = ''

  try {
    const content = await GetRAGConfig()
    configContent.value = content
  } catch (err: any) {
    configError.value = '读取配置失败: ' + err.toString()
  } finally {
    configLoading.value = false
  }
}

// Close config editor
function closeConfigEditor() {
  showConfigEditor.value = false
  configContent.value = ''
  configError.value = ''
}

// Save config with auto-restart
async function saveConfig() {
  if (!configContent.value.trim()) {
    configError.value = '配置内容不能为空'
    return
  }

  configSaving.value = true
  configError.value = ''

  const wasRunning = status.value.running

  try {
    // Save config file
    await SaveRAGConfig(configContent.value)

    // If RAG was running, restart it
    if (wasRunning) {
      console.log('RAG was running, restarting...')

      // Stop the service
      try {
        await StopRAG()
        console.log('RAG stopped successfully')

        // Wait a moment for complete shutdown
        await new Promise(resolve => setTimeout(resolve, 2000))

        // Start the service
        await StartRAG()
        console.log('RAG restarted successfully')

        alert('配置保存成功！\n\nRAG 服务已自动重启，新配置已生效。')
      } catch (restartErr: any) {
        console.error('Failed to restart RAG:', restartErr)
        alert('配置保存成功，但重启服务失败！\n\n错误：' + restartErr.toString() + '\n\n请手动重启服务。')
      }
    } else {
      alert('配置保存成功！\n\n下次启动 RAG 服务时将使用新配置。')
    }

    closeConfigEditor()
    await checkStatus()
  } catch (err: any) {
    configError.value = '保存配置失败: ' + err.toString()
  } finally {
    configSaving.value = false
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

// Load knowledge bases list
async function loadKnowledgeBases() {
  if (!status.value.healthy) {
    // RAG service not healthy, can't load knowledge bases
    return
  }

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
    // Entering edit mode - load knowledge bases
    editingSettings.value = true
    await loadKnowledgeBases()
  } else {
    // Cancel edit - reload settings
    editingSettings.value = false
    loadRAGSettings()
  }
}

// Save RAG settings
async function saveRAGSettings() {
  // Validate topK
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

// Setup event listeners
function setupEventListeners() {
  const runtime = (window as any).runtime

  // Download progress
  const onDownloadProgress = (data: any) => {
    downloadProgress.value = {
      downloaded: data.downloaded || 0,
      total: data.total || 0,
      percent: data.percent || 0,
      status: data.status || ''
    }
  }

  // Download complete
  const onDownloadComplete = () => {
    downloading.value = false
    checkStatus()
  }

  // Download error
  const onDownloadError = (data: any) => {
    downloadError.value = data.error || '下载失败'
    downloading.value = false
  }

  // Start progress
  const onStartProgress = (data: any) => {
    startProgress.value = data.status || '正在启动...'
  }

  // Start complete
  const onStartComplete = () => {
    starting.value = false
    checkStatus()
  }

  // Start error
  const onStartError = (data: any) => {
    startError.value = data.error || '启动失败'
    starting.value = false
  }

  // Qdrant download progress
  const onQdrantDownloadProgress = (data: any) => {
    qdrantDownloadProgress.value = {
      downloaded: data.downloaded || 0,
      total: data.total || 0,
      percent: data.percent || 0,
      status: data.status || ''
    }
  }

  // Qdrant download complete
  const onQdrantDownloadComplete = () => {
    downloadingQdrant.value = false
    checkQdrantStatus()
  }

  // Qdrant download error
  const onQdrantDownloadError = (data: any) => {
    qdrantDownloadError.value = data.error || '下载 Qdrant 失败'
    downloadingQdrant.value = false
  }

  // Register listeners
  runtime.EventsOn('rag:download:progress', onDownloadProgress)
  runtime.EventsOn('rag:download:complete', onDownloadComplete)
  runtime.EventsOn('rag:download:error', onDownloadError)
  runtime.EventsOn('rag:start:progress', onStartProgress)
  runtime.EventsOn('rag:start:complete', onStartComplete)
  runtime.EventsOn('rag:start:error', onStartError)
  runtime.EventsOn('qdrant:download:progress', onQdrantDownloadProgress)
  runtime.EventsOn('qdrant:download:complete', onQdrantDownloadComplete)
  runtime.EventsOn('qdrant:download:error', onQdrantDownloadError)

  // Store cleanup functions
  cleanupFunctions = [
    () => runtime.EventsOff('rag:download:progress'),
    () => runtime.EventsOff('rag:download:complete'),
    () => runtime.EventsOff('rag:download:error'),
    () => runtime.EventsOff('rag:start:progress'),
    () => runtime.EventsOff('rag:start:complete'),
    () => runtime.EventsOff('rag:start:error'),
    () => runtime.EventsOff('qdrant:download:progress'),
    () => runtime.EventsOff('qdrant:download:complete'),
    () => runtime.EventsOff('qdrant:download:error')
  ]
}

// Lifecycle
onMounted(async () => {
  setupEventListeners()
  await checkStatus()
  await checkQdrantStatus()
  await loadRAGSettings()
  initialLoading.value = false

  // Poll status every 5 seconds if not healthy
  const interval = setInterval(async () => {
    if (!status.value.healthy) {
      await checkStatus()
    }
    // Also check Qdrant status
    await checkQdrantStatus()
  }, 5000)

  cleanupFunctions.push(() => clearInterval(interval))
})

onUnmounted(() => {
  // Cleanup all event listeners
  cleanupFunctions.forEach(cleanup => cleanup())
})
</script>
