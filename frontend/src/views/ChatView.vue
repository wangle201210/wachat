<template>
  <div class="flex h-screen bg-white">
    <!-- Main Chat Area -->
    <div class="flex-1 flex flex-col">
      <!-- Tab Bar -->
      <div class="flex items-center border-b border-gray-200 bg-gray-50">
        <!-- Tabs Container -->
        <div class="flex items-center overflow-x-auto flex-1">
          <!-- Tabs -->
          <div
            v-for="tabId in openTabs"
            :key="tabId"
            @click="switchToTab(tabId)"
            :class="[
              'group flex items-center gap-2 px-4 py-2 border-r border-gray-200 cursor-pointer min-w-0 max-w-[200px] rounded-t-sm',
              activeConversationId === tabId ? 'bg-white' : 'bg-gray-50 hover:bg-gray-100'
            ]"
          >
            <span :class="[
              'text-sm truncate flex-1',
              activeConversationId === tabId ? 'text-gray-900' : 'text-gray-600'
            ]">
              {{ getConversationTitle(tabId) }}
            </span>
            <button
              v-if="openTabs.length > 1"
              @click.stop="closeTab(tabId)"
              class="opacity-0 group-hover:opacity-100 p-0.5 hover:bg-gray-200 rounded"
            >
              <svg class="w-3.5 h-3.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- New Tab Button -->
          <button
            @click="handleCreateNewTab"
            class="px-4 py-2 hover:bg-gray-100 border-r border-gray-200 flex-shrink-0"
            title="新建对话"
          >
            <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </button>
        </div>

        <!-- Knowledge Base Button (Right) -->
        <button
          v-if="ragEnabled"
          @click="router.push('/rag')"
          class="px-4 py-2 hover:bg-gray-100 flex-shrink-0"
          title="知识库管理"
        >
          <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </button>

        <!-- Config Button (Right) -->
        <button
          @click="toggleEditConfig"
          :class="['px-4 py-2 hover:bg-gray-100 flex-shrink-0', editingConfig ? 'bg-gray-200' : '']"
          title="配置文件"
        >
          <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>

        <!-- History Button (Right) -->
        <button
          @click="showHistory = !showHistory"
          :class="['px-4 py-2 hover:bg-gray-100 flex-shrink-0', showHistory ? 'bg-gray-200' : '']"
          title="历史记录"
        >
          <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>
      </div>

      <!-- Messages Area -->
      <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4">
        <div class="space-y-6">
          <ChatMessage
            v-for="message in currentMessages"
            :key="message.id"
            :message="message"
          />

          <!-- Loading Indicator (AI Thinking) -->
          <div v-if="isLoading && !streamingMessage" class="flex gap-3">
            <AvatarAI />
            <div class="flex items-center gap-2">
              <div class="flex space-x-1">
                <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.1s"></div>
                <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
              </div>
              <span class="text-sm text-gray-500">正在思考...</span>
            </div>
          </div>

          <!-- Streaming Message -->
          <div v-if="streamingMessage" class="flex gap-3">
            <AvatarAI />
            <div class="flex-1">
              <div class="prose prose-sm prose-slate max-w-none">
                <NodeRenderer :content="streamingMessage" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Area -->
      <ChatInput
        v-model="inputMessage"
        :disabled="isSending"
        @send="handleSendMessage"
      />
    </div>

    <!-- History Sidebar -->
    <transition
      enter-active-class="transition-all duration-200"
      leave-active-class="transition-all duration-200"
      enter-from-class="translate-x-full"
      leave-to-class="translate-x-full"
    >
      <div v-if="showHistory" class="w-64 border-l border-gray-200 flex flex-col bg-gray-50">
        <!-- Conversation List -->
        <div class="flex-1 overflow-y-auto p-2">
          <div
            v-for="conv in conversations"
            :key="conv.id"
            @click="handleSelectConversation(conv.id)"
            :class="[
              'group px-2.5 py-2 mb-1 rounded cursor-pointer text-xs transition-colors',
              activeConversationId === conv.id
                ? 'bg-blue-50 text-blue-700'
                : 'hover:bg-gray-100 text-gray-700'
            ]"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="truncate flex-1">{{ conv.title || '新对话' }}</span>
              <button
                v-if="conversations.length > 1"
                @click.stop="deleteConversation(conv.id)"
                class="opacity-0 group-hover:opacity-100 p-0.5 hover:bg-gray-200 rounded"
              >
                <svg class="w-3 h-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>

      </div>
    </transition>

    <!-- Config Editor Sidebar -->
    <transition
      enter-active-class="transition-all duration-200"
      leave-active-class="transition-all duration-200"
      enter-from-class="translate-x-full"
      leave-to-class="translate-x-full"
    >
      <div v-if="editingConfig" class="w-[500px] border-l border-gray-200 flex flex-col bg-white">
        <!-- Header -->
        <div class="border-b border-gray-200 p-4 flex items-center justify-between">
          <h3 class="text-sm font-semibold text-gray-900">配置文件编辑</h3>
          <button
            @click="toggleEditConfig"
            class="text-gray-400 hover:text-gray-600"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Config Editor Content -->
        <div class="flex-1 flex flex-col p-4 overflow-hidden">
          <!-- Loading indicator -->
          <div v-if="loadingConfig" class="text-sm text-gray-500 text-center py-8">
            加载中...
          </div>

          <!-- Config editor -->
          <div v-else class="flex flex-col h-full">
            <textarea
              v-model="configContent"
              class="flex-1 px-3 py-2 text-sm font-mono border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
              placeholder="配置文件内容..."
            />

            <!-- Save Button -->
            <button
              @click="saveConfig"
              :disabled="savingConfig"
              class="w-full mt-4 px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {{ savingConfig ? '保存中...' : '保存配置' }}
            </button>

            <p class="text-xs text-gray-500 mt-3">
              💡 提示：修改配置后需要重启应用才能生效。
            </p>
          </div>
        </div>
      </div>
    </transition>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import NodeRenderer from 'vue-renderer-markdown'
import 'katex/dist/katex.min.css'
import ChatMessage from '../components/ChatMessage.vue'
import ChatInput from '../components/ChatInput.vue'
import AvatarAI from '../components/AvatarAI.vue'
import { useChat } from '../composables/useChat'
import { useAutoScroll } from '../composables/useAutoScroll'
import { DeleteConversation, GetRAGServerInfo, GetConfig, SaveConfig } from '../../wailsjs/go/main/App'

const router = useRouter()
const inputMessage = ref('')
const showHistory = ref(false)
const ragEnabled = ref(false)
const openTabs = ref<string[]>([])
let tempTabCounter = 0 // 用于生成临时tab ID

// Config editing
const editingConfig = ref(false)
const configContent = ref('')
const savingConfig = ref(false)
const loadingConfig = ref(false)

// Chat logic
const {
  conversations,
  activeConversationId,
  currentConversation,
  currentMessages,
  streamingMessage,
  isSending,
  isLoading,
  loadConversations,
  createNewConversation,
  selectConversation,
  sendMessage,
  setupEventListeners
} = useChat()

// Auto-scroll logic
const { messagesContainer, scrollToBottom, setupAutoScroll } = useAutoScroll()

// Setup auto-scroll watchers
setupAutoScroll(
  streamingMessage,
  computed(() => currentMessages.value.length)
)

// Check if a tab is temporary (not saved yet)
function isTempTab(id: string) {
  return id.startsWith('temp-')
}

// Get conversation title by ID
function getConversationTitle(id: string) {
  if (isTempTab(id)) {
    return '新对话'
  }
  const conv = conversations.value.find(c => c.id === id)
  return conv?.title || '新对话'
}

// Switch to tab
async function switchToTab(id: string) {
  if (isTempTab(id)) {
    // 临时tab不需要加载conversation，只需要设置为active
    activeConversationId.value = id
  } else {
    await selectConversation(id)
  }
  scrollToBottom()
}

// Create new tab (temporary, not saved until first message)
async function handleCreateNewTab() {
  const tempId = `temp-${++tempTabCounter}`
  openTabs.value.push(tempId)
  activeConversationId.value = tempId
  scrollToBottom()
}

// Close tab
async function closeTab(id: string) {
  if (openTabs.value.length <= 1) {
    return
  }

  const index = openTabs.value.indexOf(id)
  if (index === -1) return

  openTabs.value.splice(index, 1)

  // If closing active tab, switch to another
  if (activeConversationId.value === id) {
    const newActiveId = openTabs.value[Math.max(0, index - 1)]
    await switchToTab(newActiveId)
  }
}

// Select conversation from history (replaces current tab content)
async function handleSelectConversation(id: string) {
  // Check if already open in a tab
  if (!openTabs.value.includes(id)) {
    // Replace current tab
    const currentIndex = openTabs.value.indexOf(activeConversationId.value!)
    if (currentIndex !== -1) {
      openTabs.value[currentIndex] = id
    }
  }

  await selectConversation(id)
  scrollToBottom()
  showHistory.value = false
}

async function handleSendMessage(message: string) {
  inputMessage.value = ''
  scrollToBottom()

  // 如果当前是临时tab，先创建真实的conversation
  if (activeConversationId.value && isTempTab(activeConversationId.value)) {
    const tempId = activeConversationId.value
    await createNewConversation()

    // 替换openTabs中的临时ID为真实ID
    const index = openTabs.value.indexOf(tempId)
    if (index !== -1 && activeConversationId.value) {
      openTabs.value[index] = activeConversationId.value
    }
  }

  await sendMessage(message)
}

async function deleteConversation(id: string) {
  if (conversations.value.length <= 1) {
    // 如果只剩一个对话，不允许删除
    return
  }

  try {
    await DeleteConversation(id)

    // 如果删除的是当前对话，切换到另一个对话
    if (activeConversationId.value === id) {
      const remainingConvs = conversations.value.filter(c => c.id !== id)
      if (remainingConvs.length > 0) {
        await selectConversation(remainingConvs[0].id)
      }
    }

    // 重新加载对话列表
    await loadConversations()
  } catch (error) {
    console.error('Failed to delete conversation:', error)
  }
}

// Load config content
async function loadConfig() {
  loadingConfig.value = true
  try {
    const content = await GetConfig()
    configContent.value = content
  } catch (err: any) {
    console.error('Failed to load config:', err)
    alert('加载配置失败：' + (err.message || err))
  } finally {
    loadingConfig.value = false
  }
}

// Toggle edit config
async function toggleEditConfig() {
  if (!editingConfig.value) {
    editingConfig.value = true
    await loadConfig()
  } else {
    editingConfig.value = false
  }
}

// Save config content
async function saveConfig() {
  savingConfig.value = true
  try {
    await SaveConfig(configContent.value)
    editingConfig.value = false
    alert('配置已保存！请重启应用使配置生效。')
  } catch (err: any) {
    console.error('Failed to save config:', err)
    alert('保存失败：' + (err.message || err))
  } finally {
    savingConfig.value = false
  }
}

// Initialize
onMounted(async () => {
  setupEventListeners()
  await loadConversations()

  if (conversations.value.length === 0) {
    // 如果没有历史记录，创建临时tab
    const tempId = `temp-${++tempTabCounter}`
    openTabs.value = [tempId]
    activeConversationId.value = tempId
  } else {
    // Open first conversation as initial tab
    const firstConv = conversations.value[0]
    openTabs.value = [firstConv.id]
    await selectConversation(firstConv.id)
  }

  // Check if RAG is enabled
  try {
    const info = await GetRAGServerInfo()
    ragEnabled.value = info.enabled
  } catch (err) {
    console.error('Failed to get RAG server info:', err)
  }

  // Listen for config changes
  const runtime = (window as any).runtime
  runtime.EventsOn('config:changed', (data: any) => {
    console.log('Config changed:', data)
    // Show notification
    alert('配置文件已更新！\n\n' + (data.message || '请重启应用使配置生效'))

    // Reload RAG status
    GetRAGServerInfo().then(info => {
      ragEnabled.value = info.enabled
    }).catch(err => {
      console.error('Failed to refresh RAG status:', err)
    })
  })
})
</script>

<style scoped>
.prose :deep(p) {
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
  color: #374151;
}

.prose :deep(pre) {
  margin-top: 0;
  margin-bottom: 0;
}

.prose :deep(code) {
  font-size: 0.875rem;
}

.prose :deep(h1),
.prose :deep(h2),
.prose :deep(h3),
.prose :deep(h4),
.prose :deep(h5),
.prose :deep(h6) {
  color: #1f2937;
}

.prose :deep(strong) {
  color: #1f2937;
}

.prose :deep(li) {
  color: #374151;
}

.prose :deep(blockquote) {
  color: #6b7280;
}

.prose :deep(hr) {
  border-color: #e5e7eb;
}
</style>
