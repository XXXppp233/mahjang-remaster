<template>
  <div class="chat-container">
    <div class="chat-header">
      <h3>聊天</h3>
    </div>

    <div class="chat-messages" ref="messagesContainer">
      <Message v-for="(item, id) in chatLog.data" :key="id" :messageData="item" :timestamp="id" />
    </div>

    <div class="chat-input">
      <input
        v-model="newMessage"
        @keyup.enter="sendMessage"
        @focus="onInputFocus"
        @blur="onInputBlur"
        placeholder="What can I say ?"
        class="message-input"
      />
      <button @click="sendMessage" class="send-button">发送</button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import { statusStore } from '@/stores/status'
import Message from './Message.vue'

const props = defineProps({
  myname: { type: String, default: 'me' },
  inheritedlog: { type: Object, default: () => ({}) },
})
const messagesContainer = ref(null)
const newMessage = ref('')
const chatLog = useChatStore()
const status = statusStore()

const onInputFocus = () => {
  status.setTyping(true)
}

const onInputBlur = () => {
  status.setTyping(false)
}

async function sendMessage() {
  if (newMessage.value.trim() === '') return

  await status.apiFetch(
    `/rooms/chat?roomid=${encodeURIComponent(status.roomid)}&uuid=${encodeURIComponent(status.mysid)}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: newMessage.value }),
    },
  )

  newMessage.value = ''
}

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

watch(
  chatLog.data,
  () => {
    nextTick(scrollToBottom)
  },
  { deep: true },
)
onMounted(() => {
  scrollToBottom()
})

// Incoming messages are now handled via SSE in sse.ts which calls chatLog.addMessage
</script>

<style scoped>
.chat-container {
  --bg-main: #fff;
  --bg-secondary: #fff;
  --bg-header: #fff;
  --text-primary: #111;
  --text-header: #111;
  --border-color: #111;
  --input-bg: #fff;
  --input-text: #111;
  --btn-primary-bg: #fff;
  --btn-primary-hover: #eee;

  z-index: 1;
  width: 18vw;
  height: 60vh;
  border: 0.4vh solid var(--border-color);
  background: var(--bg-main);
  border-radius: 0.5vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0.6vh 0.6vh 0 var(--border-color);
  font-family: 'Microsoft YaHei', sans-serif;
  overflow: hidden;
}

.chat-header {
  height: 4vh;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 0.5vw;
  background: var(--bg-header);
  border-bottom: 0.4vh solid var(--border-color);
  flex-shrink: 0;
}

.chat-header h3 {
  margin: 0;
  font-size: 2.2vh;
  font-weight: 800;
}
.chat-messages {
  flex: 1;
  padding: 0.5vw;
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  background: var(--bg-secondary);
}
.chat-input {
  height: 6vh;
  display: flex;
  padding: 0.8vh;
  border-top: 0.4vh solid var(--border-color);
  background: var(--bg-main);
  flex-shrink: 0;
  gap: 0.5vw;
}
.message-input {
  flex: 1;
  min-width: 0;
  height: 100%;
  font-size: 1.8vh;
  padding: 0 1vw;
  border: 0.3vh solid var(--border-color);
  background-color: var(--input-bg);
  color: var(--input-text);
  border-radius: 0.3vh;
  outline: none;
  font-weight: 600;
}
.message-input:focus {
  background: #fdfdfd;
}
.send-button {
  flex-shrink: 0;
  font-size: 1.8vh;
  padding: 0 1.5vw;
  background: #111;
  color: white;
  border: 0.3vh solid #111;
  border-radius: 0.3vh;
  cursor: pointer;
  font-weight: 700;
  transition: all 0.1s;
}
.send-button:hover:not(:disabled) {
  background: #333;
}
.send-button:active:not(:disabled) {
  transform: scale(0.95);
}
.message-input:disabled,
.send-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
.chat-messages::-webkit-scrollbar {
  display: none;
}
</style>
