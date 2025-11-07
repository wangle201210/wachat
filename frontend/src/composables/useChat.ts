import { ref, computed } from 'vue'
import { CreateConversation, SendMessageStream, ListConversations, GetConversation } from '../../wailsjs/go/main/App'

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

export interface Conversation {
  id: string
  title: string
  messages: Message[]
  createdAt: Date
  updatedAt: Date
}

export function useChat() {
  const conversations = ref<Conversation[]>([])
  const activeConversationId = ref<string | null>(null)
  const streamingMessage = ref('')
  const isSending = ref(false)

  const currentConversation = computed(() => {
    return conversations.value.find(c => c.id === activeConversationId.value)
  })

  const currentMessages = computed(() => {
    return currentConversation.value?.messages || []
  })

  async function loadConversations() {
    try {
      const convs = await ListConversations()
      conversations.value = convs || []
    } catch (error) {
      console.error('Failed to load conversations:', error)
    }
  }

  async function createNewConversation() {
    try {
      const now = new Date()
      const title = `新对话 ${now.getMonth() + 1}/${now.getDate()}`
      const conv = await CreateConversation(title)
      conversations.value.unshift(conv)
      activeConversationId.value = conv.id
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
  }

  async function selectConversation(id: string) {
    activeConversationId.value = id

    // Load conversation with messages
    try {
      const conv = await GetConversation(id)
      const existingConv = conversations.value.find(c => c.id === id)
      if (existingConv && conv) {
        existingConv.messages = conv.messages || []
      }
    } catch (error) {
      console.error('Failed to load conversation messages:', error)
    }
  }

  async function sendMessage(message: string) {
    if (!message.trim() || !activeConversationId.value || isSending.value) {
      return
    }

    isSending.value = true
    streamingMessage.value = ''

    try {
      const userMessage: Message = {
        id: Date.now().toString(),
        role: 'user',
        content: message,
        timestamp: new Date()
      }

      const conv = currentConversation.value
      if (conv) {
        conv.messages.push(userMessage)
      }

      await SendMessageStream(activeConversationId.value, message)
    } catch (error) {
      console.error('Failed to send message:', error)
      isSending.value = false
      streamingMessage.value = ''
    }
  }

  function setupEventListeners() {
    const runtime = (window as any).runtime
    if (runtime && runtime.EventsOn) {
      runtime.EventsOn('stream:start', (data: any) => {
        console.log('Stream started:', data)
        streamingMessage.value = ''
      })

      runtime.EventsOn('stream:response', (data: any) => {
        console.log('Stream response:', data)
        if (data.chunk) {
          streamingMessage.value += data.chunk
        }
      })

      runtime.EventsOn('stream:end', (data: any) => {
        console.log('Stream ended:', data)
        const conv = currentConversation.value
        if (conv && data.message) {
          conv.messages.push(data.message)
        }
        streamingMessage.value = ''
        isSending.value = false
      })

      runtime.EventsOn('stream:error', (data: any) => {
        console.error('Stream error:', data)
        alert('发送消息失败: ' + data.error)
        streamingMessage.value = ''
        isSending.value = false
      })

      runtime.EventsOn('conversation:title-updated', (data: any) => {
        console.log('Title updated:', data)
        const conv = conversations.value.find(c => c.id === data.conversationId)
        if (conv) {
          conv.title = data.title
        }
      })
    } else {
      console.error('Wails runtime not available')
    }
  }

  return {
    conversations,
    activeConversationId,
    currentConversation,
    currentMessages,
    streamingMessage,
    isSending,
    loadConversations,
    createNewConversation,
    selectConversation,
    sendMessage,
    setupEventListeners
  }
}
