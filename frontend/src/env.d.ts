/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// Declare electron global
interface Window {
  electron: {
    ipcRenderer: {
      send(channel: string, ...args: any[]): void
      on(channel: string, listener: (event: any, ...args: any[]) => void): void
      once(channel: string, listener: (event: any, ...args: any[]) => void): void
      removeListener(channel: string, listener: Function): void
      removeAllListeners(channel: string): void
      invoke(channel: string, ...args: any[]): Promise<any>
    }
  }
  api: any
}
