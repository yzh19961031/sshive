// frontend/src/stores/terminal.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface TermTab {
  id: string
  name: string
  hostId: number
  connected: boolean
  everConnected: boolean
  outputBuf: (Uint8Array | string)[]
  ws?: WebSocket
}

const BUF_LIMIT = 300

export const useTerminalStore = defineStore('terminal', () => {
  const tabs = ref<TermTab[]>([])
  const activeId = ref('')

  function getTab(id: string) {
    return tabs.value.find(t => t.id === id)
  }

  function addTab(tab: TermTab) {
    tabs.value.push(tab)
    activeId.value = tab.id
  }

  function removeTab(id: string) {
    const tab = getTab(id)
    if (tab?.ws && tab.ws.readyState !== WebSocket.CLOSED) tab.ws.close()
    tabs.value = tabs.value.filter(t => t.id !== id)
    if (activeId.value === id) {
      activeId.value = tabs.value[tabs.value.length - 1]?.id ?? ''
    }
  }

  function appendOutput(id: string, data: Uint8Array | string) {
    const tab = getTab(id)
    if (!tab) return
    tab.outputBuf.push(data)
    if (tab.outputBuf.length > BUF_LIMIT) tab.outputBuf.shift()
  }

  function clearAll() {
    tabs.value.forEach(t => {
      if (t.ws && t.ws.readyState !== WebSocket.CLOSED) t.ws.close()
    })
    tabs.value = []
    activeId.value = ''
  }

  return { tabs, activeId, getTab, addTab, removeTab, appendOutput, clearAll }
})
