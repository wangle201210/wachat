<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 bg-gray-900 bg-opacity-60 backdrop-blur-sm flex items-center justify-center z-50"
    @click.self="close"
  >
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-5xl max-h-[90vh] flex flex-col">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between bg-gradient-to-r from-blue-50 to-indigo-50">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-blue-500 rounded-lg">
            <IconSettings class="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 class="text-xl font-semibold text-gray-900">编辑 RAG 配置</h2>
            <p class="text-sm text-gray-600">配置文件路径: ~/.wachat/go-rag/config.yaml</p>
          </div>
        </div>
        <button
          @click="close"
          class="text-gray-400 hover:text-gray-600 transition-colors p-2 hover:bg-white rounded-lg"
        >
          <IconClose class="w-6 h-6" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="flex-1 overflow-auto p-6 bg-gray-50">
        <div v-if="loading" class="flex items-center justify-center h-64">
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

          <div v-if="error" class="mt-4 p-4 bg-red-50 border-l-4 border-red-500 rounded text-sm text-red-700 flex items-start gap-3">
            <IconAlert class="w-5 h-5 flex-shrink-0 mt-0.5" />
            <span>{{ error }}</span>
          </div>

          <div class="mt-4 p-4 bg-blue-50 border-l-4 border-blue-500 rounded text-sm text-blue-900">
            <p class="font-medium mb-2 flex items-center gap-2">
              <IconInfo class="w-5 h-5" />
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
          <span v-if="isServiceRunning" class="flex items-center gap-2">
            <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
            服务运行中，保存后将自动重启
          </span>
        </div>
        <div class="flex items-center gap-3">
          <button
            @click="close"
            class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            @click="save"
            :disabled="saving || loading"
            class="px-6 py-2 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-sm flex items-center gap-2"
          >
            <IconSave :class="['w-4 h-4', { 'animate-spin': saving }]" />
            {{ saving ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { IconSettings, IconClose, IconAlert, IconInfo, IconSave } from '@/components/icons'

interface Props {
  modelValue: boolean
  isServiceRunning?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isServiceRunning: false
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'save': [content: string]
  'load': []
}>()

const configContent = ref('')
const loading = ref(false)
const saving = ref(false)
const error = ref('')

// Watch for modal opening to trigger loading
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    loadConfig()
  } else {
    // Reset state when closing
    error.value = ''
  }
})

function loadConfig() {
  loading.value = true
  error.value = ''
  emit('load')
}

// This should be called from parent after loading is complete
function setConfigContent(content: string) {
  configContent.value = content
  loading.value = false
}

function setError(errorMessage: string) {
  error.value = errorMessage
  loading.value = false
  saving.value = false
}

function setSaving(isSaving: boolean) {
  saving.value = isSaving
}

function close() {
  if (!saving.value) {
    emit('update:modelValue', false)
    configContent.value = ''
    error.value = ''
  }
}

function save() {
  if (!configContent.value.trim()) {
    error.value = '配置内容不能为空'
    return
  }

  error.value = ''
  saving.value = true
  emit('save', configContent.value)
}

// Expose methods for parent to call
defineExpose({
  setConfigContent,
  setError,
  setSaving,
  close
})
</script>
