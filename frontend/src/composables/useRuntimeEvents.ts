import { onMounted, onUnmounted } from 'vue'

type EventHandler = (data: any) => void

/**
 * Composable for managing Wails runtime events
 */
export function useRuntimeEvents() {
  const runtime = (window as any).runtime
  const cleanupFunctions: (() => void)[] = []

  /**
   * Register an event listener
   */
  function on(eventName: string, handler: EventHandler) {
    runtime.EventsOn(eventName, handler)
    cleanupFunctions.push(() => runtime.EventsOff(eventName))
  }

  /**
   * Register multiple event listeners
   */
  function onMultiple(events: Record<string, EventHandler>) {
    Object.entries(events).forEach(([eventName, handler]) => {
      on(eventName, handler)
    })
  }

  /**
   * Cleanup all registered event listeners
   */
  function cleanup() {
    cleanupFunctions.forEach(fn => fn())
    cleanupFunctions.length = 0
  }

  // Auto cleanup on unmount
  onUnmounted(() => {
    cleanup()
  })

  return {
    on,
    onMultiple,
    cleanup
  }
}

/**
 * Create a typed event handler for service events
 */
export function createServiceEventHandlers<T = any>(options: {
  onProgress?: (data: T) => void
  onComplete?: (data?: T) => void
  onError?: (error: string) => void
}) {
  return {
    progress: options.onProgress || (() => {}),
    complete: options.onComplete || (() => {}),
    error: options.onError || (() => {})
  }
}
