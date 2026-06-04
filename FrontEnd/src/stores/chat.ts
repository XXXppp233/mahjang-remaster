import { ref } from 'vue'
import { defineStore } from 'pinia'

// 1. 定义消息的类型
interface ChatMessage {
  type: 'chat'
  name: string
  message: string
}

interface LogMessage {
  type: 'log'
  level: 'info' | 'warning' | 'error'
  message: string
}

type MessageEntry = ChatMessage | LogMessage
type ChatLog = {
  [key: string]: MessageEntry
} // 666 神仙语法

export const useChatStore = defineStore('chat', () => {
  const data = ref<ChatLog>({
    0: { type: 'chat', name: 'Server', message: 'F11 以全屏游玩' },
    1: { type: 'chat', name: 'Server', message: '点击 Server 以隐藏该消息。' },
    2: { type: 'chat', name: 'Server', message: 'Q/E 选择牌，空格键 出牌' },
    3: { type: 'chat', name: 'Server', message: '1-6 选择操作，W 键 确认操作' },
    4: { type: 'log', level: 'info', message: '房间已创建' },

  })
  function addMessage(message: MessageEntry) {
    const timestamp = Date.now().toString()
    data.value[timestamp] = message
  }

  return { data, addMessage }
})
