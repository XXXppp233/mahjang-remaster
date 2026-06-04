import { reactive } from 'vue'
import { statusStore } from './status'
import { useChatStore } from './chat'

export const sseState = reactive({
  connected: false,
})

let eventSource: EventSource | null = null
let currentPlayer = ''

export function connectSSE() {
  const status = statusStore()
  if (!status.mysid) return
  if (eventSource && currentPlayer === status.mysid) return
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }

  currentPlayer = status.mysid
  status.connected = false
  sseState.connected = false
  const query = `?player=${encodeURIComponent(status.mysid)}`
  eventSource = new EventSource(`/api/stream${query}`)

  eventSource.onopen = () => {
    sseState.connected = true
    status.connected = true
  }

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.type === 'chat') {
        useChatStore().addMessage({
          type: 'chat',
          name: data.name,
          message: data.message,
        })
        return
      }
      if (data.type === 'log') {
        useChatStore().addMessage({
          type: 'log',
          level: data.level ?? 'info',
          message: data.message,
        })
        return
      }
      if (data.type === 'room_user') {
        status.updateRoomUser(data)
        return
      }
      if (data.type === 'action' || data.action) {
        status.updateActionEvent(data)
        return
      }
      status.updateGameState(data)
    } catch (e) {
      console.error('Failed to parse SSE message', e)
    }
  }

  eventSource.onerror = () => {
    sseState.connected = false
    status.connected = false
    eventSource?.close()
    eventSource = null
    currentPlayer = ''
  }
}

export function disconnectSSE() {
  const status = statusStore()
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  currentPlayer = ''
  sseState.connected = false
  status.connected = false
}
