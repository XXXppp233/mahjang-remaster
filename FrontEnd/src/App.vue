<script setup>
import { Analytics } from "@vercel/analytics/vue"
import { SpeedInsights } from "@vercel/speed-insights/vue"
import { computed, onMounted, watch } from 'vue'
import { statusStore } from '@/stores/status'
import { connectSSE, disconnectSSE } from '@/stores/sse'
import MahjongGame from './components/MahJong/MahjongGame.vue'
import Intro from './components/MahJong/Intro.vue'
import RoomList from './components/MahJong/RoomList.vue'
import Room from './components/MahJong/Room.vue'

const status = statusStore()
const now = computed(() => status.now)
const sseKey = computed(() => {
  return status.isLogin && status.mysid ? status.mysid : ''
})

onMounted(() => {
  status.loadConfig()
})

watch(
  sseKey,
  (key) => {
    if (key) {
      connectSSE()
      return
    }
    disconnectSSE()
  },
)
</script>

<template>
  <Analytics/>
  <SpeedInsights/>
  <Intro v-if="now === 'nologin' || now === 'login'" />
  <RoomList v-if="now === 'roomlist'" />
  <Room v-if="now === 'room'" />
  <MahjongGame v-if="now === 'gaming'" />
</template>

<style></style>
