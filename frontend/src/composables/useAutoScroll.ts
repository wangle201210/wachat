import { ref, watch, nextTick } from 'vue'

export function useAutoScroll() {
  const messagesContainer = ref<HTMLElement | null>(null)

  // Check if user is near the bottom of messages
  function isNearBottom(): boolean {
    if (!messagesContainer.value) return true
    const element = messagesContainer.value
    const threshold = 10 // pixels from bottom
    return element.scrollHeight - element.scrollTop - element.clientHeight <= threshold
  }

  // Scroll to bottom of messages
  function scrollToBottom() {
    if (!messagesContainer.value) return
    nextTick(() => {
      messagesContainer.value!.scrollTop = messagesContainer.value!.scrollHeight
    })
  }

  // Setup watchers for auto-scroll
  function setupAutoScroll(streamingMessage: any, messagesLength: any) {
    // Auto-scroll when streaming message updates
    watch(streamingMessage, () => {
      if (isNearBottom()) {
        scrollToBottom()
      }
    })

    // Auto-scroll when new message is added
    watch(messagesLength, () => {
      nextTick(() => {
        if (isNearBottom()) {
          scrollToBottom()
        }
      })
    })
  }

  return {
    messagesContainer,
    isNearBottom,
    scrollToBottom,
    setupAutoScroll
  }
}
