<template>
  <div class="border-t border-gray-200 p-4 bg-white">
    <textarea
      v-model="inputText"
      @keydown="handleKeydown"
      :placeholder="disabled ? '正在等待回复...' : '输入消息... (Enter 发送, ⌘+Enter 换行)'"
      class="w-full px-4 py-3 resize-none outline-none focus:outline-none"
      rows="3"
    ></textarea>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  modelValue: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'send': [message: string]
}>()

const inputText = ref(props.modelValue)

watch(() => props.modelValue, (newVal) => {
  inputText.value = newVal
})

watch(inputText, (newVal) => {
  emit('update:modelValue', newVal)
})

function handleKeydown(event: KeyboardEvent) {
  // Command+Enter 或 Ctrl+Enter 换行
  if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
    // 允许默认换行行为
    return
  }

  // Enter 发送消息
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    handleSend()
  }
}

function handleSend() {
  if (inputText.value.trim() && !props.disabled) {
    emit('send', inputText.value.trim())
  }
}
</script>
