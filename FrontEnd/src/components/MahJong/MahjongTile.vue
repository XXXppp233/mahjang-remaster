<script setup>
import { ref, onMounted, computed } from 'vue'
const props = defineProps({
  tile: {
    type: String,
    required: true,
  },
  index: {
    type: Number,
    required: false,
  },
  selected: {
    type: Boolean,
    default: false,
  },
  locked: {
    type: Boolean,
    required: false,
    default: false,
  },
})
const isPressed = ref(false)
const isDarkMode = ref(false)

const emit = defineEmits(['click'])
const ensmall = () => {
  isPressed.value = true
  if (props.locked) return
  emit('click', props.index)
}
const enlarge = () => {
  isPressed.value = false
}

// 深色模式
const checkDarkMode = () => {
  isDarkMode.value = window.matchMedia('(prefers-color-scheme: dark)').matches
}

onMounted(() => {
  checkDarkMode()
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', checkDarkMode)
})

// 计算瓦片颜色
const tileColor = computed(() => (isDarkMode.value ? 'Black' : 'Regular'))
</script>
<template>
  <div>
    <button
      @mouseup="enlarge"
      @mousedown="ensmall"
      @mouseleave="enlarge"
      :class="{ pressed: isPressed, selected: props.selected, locked: props.locked }"
      :disabled="props.locked"
    >
      <img :src="`tilesvgs/${tileColor}/${props.tile}.svg`" />
    </button>
  </div>
</template>
<style scoped>
div {
  width: 5vw;
  height: 13vh;
  margin: 0;
  align-items: flex-end;
  justify-content: center;
}
@media (prefers-color-scheme: dark) {
  button {
    border-color: #5c5c5c !important;
  }
}

div {
  height: 13vh;
}

button {
  width: 5vw;
  height: 13vh;
  border: 0.35vw solid #111;
  border-radius: 0.3vw;
  padding: 0.2vw;
  caret-color: transparent; /* 隐藏光标 */
  transition: all 0.1s;
  box-shadow: 0.3vh 0.3vh 0 #111;
}

button.pressed {
  transform: translate(0.1vh, 0.1vh);
  box-shadow: 0.1vh 0.1vh 0 #111;
}
button:hover {
  transform: scale(1.05) translate(-0.1vh, -0.1vh);
  box-shadow: 0.5vh 0.5vh 0 #111;
}
button.selected {
  transform: translateY(-2vh);
  border-color: #f44336;
  box-shadow: 0.6vh 0.6vh 0 #111;
  background: #fffbe6;
}
button.selected:hover {
  transform: scale(1.05) translateY(-2vh);
}
button.locked {
  background: #eee;
  border-color: #666;
  box-shadow: 0.2vh 0.2vh 0 #666;
}
button.locked:hover {
  transform: none;
  box-shadow: 0.2vh 0.2vh 0 #666;
}

img {
  width: 100%;
  height: 100%;
}
</style>
