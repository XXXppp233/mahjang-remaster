<template>
  <span>
    <button
      v-if="props.label"
      :class="{ pressed: ispressed, left: isleft, right: isright, selected: props.selected, red: props.label.startsWith('胡') }"
      :disabled="pending"
      @mousedown="ispressed = true"
      @mouseup="ispressed = false"
      @click="handleAction"
    >
      {{ props.label }}
    </button>
  </span>
</template>
<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import keySounds from '/src/assets/keyboardsounds/MikuTap/main.js'
import { statusStore } from '@/stores/status'

const props = defineProps({
  label: {
    type: String,
    required: true,
  },
  actionid: {
    type: Number,
    required: true,
  },
  selected: {
    type: Boolean,
    default: false,
  },
  length: {
    type: Number,
    required: true,
  },
  data: {
    type: [Array, Boolean],
    default: null,
  }
})
const emit = defineEmits(['click'])
const ispressed = ref(false)
const pending = ref(false)

async function handleAction() {
  if (pending.value) return
  const status = statusStore()
  let action = ''
  let body = { PlayerIndex: status.myid }

  if (props.label.startsWith('胡')) {
    action = 'hu'
  } else if (props.label.startsWith('杠')) {
    action = 'kong'
    body = { ...body, Selec: props.data } // Assuming data contains indices
  } else if (props.label.startsWith('碰')) {
    action = 'pong'
    body = { ...body, Selec: props.data }
  // } else if (props.label.startsWith('吃')) {
  //   action = 'chow'
  //   body = { ...body, Selec: props.data }
  } else {
    return
  }

  pending.value = true
  try {
    await status.apiFetch(`/gaming/${action}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    })
  } finally {
    status.reSetActions()
    pending.value = false
  }
}

const isleft = ref(props.actionid === 0)
const isright = ref(props.actionid === props.length - 1)

const handleKeyDown = (event) => {
  if (props.selected){
    if (event.keyCode === 87) { // W 键
      ispressed.value = true
      keySounds.Play(event)
    }
  }
}
const handleKeyUp = (event) => {
  if (props.selected){
    if (event.keyCode === 87) { // W 键
      ispressed.value = false
      handleAction()
    }
  }
}

onMounted(() => {
  addEventListener('keydown', handleKeyDown)
  addEventListener('keyup', handleKeyUp)
})
onUnmounted(() => {
  removeEventListener('keydown', handleKeyDown)
  removeEventListener('keyup', handleKeyUp)
})

</script>
<style scoped>
@media (prefers-color-scheme: dark) {
  button {
    background-color: #333;
    color: white;
  }
}

span {
  z-index: 1;
  caret-color: transparent;
  width: 7vw;
  height: 6vh;
  display: inline-block;
}
button {
  border: 0.4vh solid #111;
  width: 7vw;
  height: 6vh;
  font-size: 1.8vw;
  font-family: 'Segoe UI Symbol';
  font-weight: 900;
  cursor: pointer;
  transition: all 0.1s;
  box-shadow: 0.4vh 0.4vh 0 #111;
  border-radius: 0.3vh;
}
button.red {
  background-color: #f44336;
  color: white;
}
button:hover {
  background-color: #ffeb3b;
  transform: translate(-0.2vh, -0.2vh);
  box-shadow: 0.6vh 0.6vh 0 #111;
}
button.pressed {
  transform: translate(0.2vh, 0.2vh);
  box-shadow: 0.2vh 0.2vh 0 #111;
}
button.selected {
  background-color: #ffc107;
  color: #111;
  border-width: 0.5vh;
}
button.left {
  /* Remove large radius */
  border-top-left-radius: 0.3vh;
  border-bottom-left-radius: 0.3vh;
}
button.right {
  /* Remove large radius */
  border-top-right-radius: 0.3vh;
  border-bottom-right-radius: 0.3vh;
}
</style>
