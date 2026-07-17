<template>
  <div :class="{ pressed: ispressed }">
    <button
      :class="{ pressed: ispressed, active: props.active }"
      :disabled="pending || !props.active"
      @mousedown="handleMouseDown"
      @mouseup="ispressed = false"
    >
      出牌
    </button>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { statusStore } from '@/stores/status'
import keySounds from '/src/assets/keyboardsounds/MikuTap/main.js'

const props = defineProps({
  selectedIndex: {
    type: Number,
    default: 0,
  },
  active: {
    type: Boolean,
    default: false,
  },
})

const ispressed = ref(false)
const pending = ref(false)

const handleMouseDown = () => {
  ispressed.value = true
  submitDiscard()
}
const handleSpaceDown = (event) => {
  if (statusStore().isTyping) return
  if (event.code === 'Space') {
    ispressed.value = true
    keySounds.Play(event)
    submitDiscard()
  }
}
const handleSpaceUp = (event) => {
  if (event.code === 'Space') {
    ispressed.value = false
  }
}

async function submitDiscard() {
  if (props.active && !pending.value) {
    const status = statusStore()
    pending.value = true
    try {
      await status.apiFetch('/gaming/discard', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ PlayerIndex: status.myid, Selec: props.selectedIndex })
      })
    } finally {
      pending.value = false
    }
  } else {
  }
}

onMounted(() => {
  addEventListener('keydown', handleSpaceDown)
  addEventListener('keyup', handleSpaceUp)
})
onUnmounted(() => {
  removeEventListener('keydown', handleSpaceDown)
  removeEventListener('keyup', handleSpaceUp)
})
</script>

<style scoped>
div {
  z-index: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  min-height: 15vh;
}
button {
  caret-color: transparent;
  width: 100%;
  height: 100%;
  min-height: 15vh;
  border: 0.4vh solid #111;
  border-radius: 0.5vh;
  background: #fff;
  color: #111;
  font-size: 2.6vh;
  font-weight: 900;
  box-shadow: 0.5vh 0.5vh 0 #111;
  transition: 0.1s;
  cursor: pointer;
}
button.active {
  background-color: #ffeb3b;
}
button:hover:not(:disabled) {
  transform: translate(-0.2vh, -0.2vh);
  box-shadow: 0.7vh 0.7vh 0 #111;
}
button.pressed {
  transform: translate(0.2vh, 0.2vh);
  box-shadow: 0.2vh 0.2vh 0 #111;
}
button:disabled {
  cursor: not-allowed;
  color: #777;
  background: #eee;
  box-shadow: 0.3vh 0.3vh 0 #777;
}

div.pressed {
  transform: none;
}
</style>
