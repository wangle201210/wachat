<template>
  <div v-if="show" class="w-64 border-r border-gray-200 flex flex-col">
    <div class="p-2 border-b border-gray-200">
      <button
        @click="$emit('create-conversation')"
        class="w-full px-3 py-1.5 bg-blue-500 text-white rounded hover:bg-blue-600 text-xs font-medium"
      >
        新对话
      </button>
    </div>
    <div class="flex-1 overflow-y-auto p-1.5">
      <div
        v-for="conv in conversations"
        :key="conv.id"
        @click="$emit('select-conversation', conv.id)"
        :class="[
          'px-2.5 py-1.5 mb-0.5 rounded cursor-pointer text-xs',
          activeConversationId === conv.id
            ? 'bg-blue-50 text-blue-700'
            : 'hover:bg-gray-50 text-gray-700'
        ]"
      >
        <div class="truncate">{{ conv.title }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Conversation {
  id: string
  title: string
  messages: any[]
  createdAt: Date
  updatedAt: Date
}

defineProps<{
  show: boolean
  conversations: Conversation[]
  activeConversationId: string | null
}>()

defineEmits<{
  'create-conversation': []
  'select-conversation': [id: string]
}>()
</script>
