<template>
  <div class="message-wrapper" :title="fullFormattedTime">
    <div v-if="messageData.type === 'chat'" class="message-item chat-style">
      <div class="message-content">
        <span class="message-name"
          ><button @click="blur" class="blur">{{ messageData.name }}</button></span
        >
        <div class="text-wrapper">
          <span v-if="isBlurred === false">{{ messageData.message }}</span>
          <span v-if="isBlurred">{{ replace(messageData.message) }}</span>
          <span class="message-time">{{ shortTime }}</span>
        </div>
      </div>
    </div>
    <div v-else-if="messageData.type === 'log'" class="message-item log-style">
      <span class="log-text">{{ messageData.message }}</span>
    </div>
  </div>
</template>

<script setup>
// Script部分也无需任何改动
import { computed, ref } from 'vue'

const props = defineProps({
  messageData: { type: Object, required: true },
  timestamp: { type: [String, Number], required: true },
  blursymbol: { type: String, default: '*' },
})
// 🀆 🀄 🃏 🀡
const replace = (str) => {
  return Array(str.length + 1).join(props.blursymbol)
}

const pad = (num) => String(num).padStart(2, '0')

const fullFormattedTime = computed(() => {
  const date = new Date(parseInt(props.timestamp))
  return date
    .toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    })
    .replace(/\//g, '-')
})

const shortTime = computed(() => {
  const date = new Date(parseInt(props.timestamp))
  return `${pad(date.getHours())}:${pad(date.getMinutes())}`
})

// 隐藏文本
const isBlurred = ref(false)
const blur = () => {
  isBlurred.value = !isBlurred.value
}
</script>

<style scoped>
/* 定义本组件所需的所有变量 */
.message-wrapper {
  --bubble-bg: #fff;
  --bubble-border: #111;
  --text-primary: #111;
  --name-color: #111;
  --log-bg: #eee;
  --log-text: #111;
}

.message-item {
  margin-bottom: 0.8vh;
  max-width: 98%;
  position: relative;
}

/* 聊天气泡样式 */
.chat-style {
  position: relative;
  padding: 0.6vh 0.8vw;
  background: var(--bubble-bg);
  border: 0.25vh solid var(--bubble-border);
  border-radius: 0.4vh;
  box-shadow: 0.3vh 0.3vh 0 var(--bubble-border);
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 0.2vh;
}
.message-name {
  caret-color: transparent;
}
.blur {
  cursor: pointer;
  margin: 0;
  border: 0;
  padding: 0;
  font-weight: 800;
  color: var(--name-color);
  font-size: 1.6vh;
  background-color: inherit;
  text-decoration: underline;
  text-underline-offset: 0.2vh;
}
.text-wrapper {
  line-height: 1.4;
  font-size: 1.5vh;
  color: var(--text-primary);
  word-wrap: break-word;
  text-align: left;
  font-weight: 500;
}
.message-time {
  display: block;
  text-align: right;
  font-size: 1vh;
  color: #666;
  margin-top: 0.2vh;
  user-select: none;
  font-weight: 700;
}

/* 日志消息样式 */
.log-style {
  color: var(--log-text);
  background-color: var(--log-bg);
  font-size: 1.2vh;
  text-align: center;
  padding: 0.4vh 1vw;
  margin: 0.5vh auto;
  border: 0.2vh solid var(--bubble-border);
  border-radius: 0.4vh;
  font-weight: 700;
}
.log-text {
  font-style: normal;
}
</style>
