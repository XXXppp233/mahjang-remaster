<template>
  <div class="intro" @mouseup="focusNameInput" @wheel="mouseWheel">
    <div v-if="!isLogin">
      <h3 class="inputprompt" :class="{ active: !isMahjor }" v-show="inputValue.length === 0">
        {{ promptText }}
        <button
          class="enter"
          @mousedown="pressEnter"
          @mouseup="unPressEnter"
          @mouseleave="isPressed = false"
          :class="{ active: !isMahjor, pressed: isPressed }"
        >
          Enter↵
        </button>
      </h3>
      <input
        autofocus
        id="nameinput"
        class="nameinput"
        :type="loginStep === 'password' ? 'password' : 'text'"
        v-model="inputValue"
        @keydown="handleKeyDown"
        @keyup="handleKeyUp"
        ref="inputRef"
      />
    </div>
    <div v-else class="welcome-container">
      <h2 class="welcome-text">{{ username }}</h2>
      <div class="user-meta">
        <span>一位 {{ status.grades.first }}</span>
        <span>二位 {{ status.grades.second }}</span>
        <span>三位 {{ status.grades.third }}</span>
        <span>四位 {{ status.grades.fourth }}</span>
      </div>
      <p v-if="status.needsReconnect" class="reconnect-note">有未完成的对局</p>
      <button class="enter active big-btn" @click="enterLobby">进入大厅</button>
      <button class="enter active big-btn" @click="handleLogout">退出登录</button>
    </div>
  </div>
  <div class="mahjor" @wheel="mouseWheel">
    <h1 class="mahjor-h" :class="{ active: isMahjor }" title="特级锦标赛">Mahjor</h1>
    <div class="mahjor-content" :class="{ active: isMahjor }"></div>
  </div>
  <div class="ranking" @wheel="mouseWheel">
    <h1 class="ranking-h" :class="{ active: isRanking }" title="天下英雄如过江之鲫">Ranking</h1>
    <div class="ranking-content" :class="{ active: isRanking }">
      <ul class="ranking-list">
        <li
          v-for="(data, index) in rankingData"
          class="ranking-item"
          :class="{ Odd: index % 2, first: index === 0 }"
        >
          <span class="ranking-index">{{ index ? index : ' ' }}</span>
          <span class="ranking-name" :title="data.name">{{ data.name }}</span>
          <span class="ranking-freq" :title="`Frequency`">{{ data.frequency }}</span>
          <span class="ranking-rate" :title="`1st`">{{ data.rate1 }}</span>
          <span class="ranking-rate" :title="`2nd`">{{ data.rate2 }}</span>
          <span class="ranking-rate" :title="`3rd`">{{ data.rate3 }}</span>
          <span class="ranking-rate" :title="`4nd`">{{ data.rate4 }}</span>
          <span class="ranking-perf" :title="`Performances`">{{ data.performances }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { statusStore } from '@/stores/status'
import keySounds from '/src/assets/keyboardsounds/MikuTap/main.js'

const status = statusStore()
const isLogin = computed(() => status.isLogin)
const username = computed(() => status.username)

const rankingData = [
  {
    name: 'Name',
    frequency: 'Freq',
    rate1: '1st Rate',
    rate2: '2nd Rate',
    rate3: '3rd Rate',
    rate4: '4th Rate',
    performances: 'Perf',
  },
]

const scrollheight = ref(0)
const isMahjor = computed(
  () => scrollheight.value >= window.innerHeight && scrollheight.value < window.innerHeight * 2,
)
const isRanking = computed(() => scrollheight.value >= window.innerHeight * 2)

const inputValue = ref('')
const loginStep = ref('account') // account, password, name
const account = ref('')
const password = ref('')

const promptText = computed(() => {
  if (loginStep.value === 'account') return status.loginRequire ? 'Enter your account' : 'Enter your name'
  if (loginStep.value === 'password') return 'Enter your password'
  if (loginStep.value === 'name') return 'Enter your name'
  return ''
})

const inputRef = ref(null)
const isPressed = ref(false)

const pressEnter = () => {
  isPressed.value = true
  keySounds.Play({ key: 'enter', keyCode: 13 })
}
const unPressEnter = () => {
  isPressed.value = false
  keySounds.Play({ key: 'enter', keyCode: 14 })
}
const focusNameInput = () => {
  if (inputRef.value) inputRef.value.focus({ preventScroll: true })
}

const handleKeyUp = (event) => {
  if (event.key === 'Enter') {
    isPressed.value = false
    handleLoginAction()
  }
}
const handleKeyDown = (event) => {
  keySounds.Play(event)
  if (event.key === 'Enter') {
    isPressed.value = false
  }
}

async function handleLoginAction() {
  if (inputValue.value.trim() === '') return

  if (loginStep.value === 'account') {
    account.value = inputValue.value.trim()
    if (status.loginRequire) {
      loginStep.value = 'password'
      inputValue.value = ''
    } else {
      performLogin()
    }
  } else if (loginStep.value === 'password') {
    password.value = inputValue.value.trim()
    performLogin()
  } else if (loginStep.value === 'name') {
    account.value = inputValue.value.trim()
    performLogin()
  }
}

async function performLogin() {
  try {
    await status.login(account.value, password.value)
    inputValue.value = ''
  } catch (e) {
    console.error(e)
    if (status.loginRequire) {
      loginStep.value = 'password'
      inputValue.value = ''
    }
  }
}

function enterLobby() {
  status.enterLobby()
}

async function handleLogout() {
  await status.logoutRemote()
  loginStep.value = 'account'
  inputValue.value = ''
}

const mouseWheel = (event) => {
  event.preventDefault()
  if (event.deltaY < 0) {
    scrollheight.value -= window.innerHeight
    if (scrollheight.value < 0) scrollheight.value = 0
    window.scrollTo({ top: scrollheight.value, behavior: 'smooth' })
  } else if (event.deltaY > 0) {
    scrollheight.value += window.innerHeight
    const maxScrollHeight = document.documentElement.scrollHeight - window.innerHeight
    if (scrollheight.value > maxScrollHeight) scrollheight.value = maxScrollHeight
    window.scrollTo({ top: scrollheight.value, behavior: 'smooth' })
  }
}
</script>

<style scoped>
.intro {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
}
.inputprompt {
  color: transparent;
  text-align: center;
  font-size: 5vh;
  margin-bottom: 2vh;
  transform: scale(0.8);
  transition: 0.5s ease;
}
.inputprompt.active {
  color: lightgray;
  transform: scale(1);
}
.enter {
  font-size: inherit;
  color: #111;
  background-color: #fff;
  border-radius: 0.5vh;
  border: 0.4vh solid #111;
  transition: 0.1s ease;
  transform: scale(1);
  box-shadow: 0.4vh 0.4vh 0 #111;
  font-weight: 800;
}
.enter.active {
  color: #111;
}
.enter:hover {
  cursor: pointer;
  transform: translate(-0.1vh, -0.1vh);
  box-shadow: 0.6vh 0.6vh 0 #111;
}
.enter.pressed {
  transform: translate(0.2vh, 0.2vh);
  box-shadow: 0.2vh 0.2vh 0 #111;
}
.nameinput {
  caret-color: transparent;
  color: inherit;
  background-color: transparent;
  text-align: center;
  font-size: 10vh;
  width: 80vw;
  border: 0;
}
.nameinput:focus {
  outline: none;
}
.welcome-container {
  display: flex;
  flex-direction: column;
  gap: 2vh;
  align-items: center;
}
.welcome-text {
  font-size: 6vh;
  margin-bottom: 1vh;
}
.user-meta {
  display: flex;
  gap: 2vw;
  flex-wrap: wrap;
  justify-content: center;
  color: #666;
  font-size: 2.4vh;
}
.reconnect-note {
  margin: 0;
  color: #333;
  font-size: 2.8vh;
}
.big-btn {
  font-size: 4vh;
  padding: 1vh 4vw;
}

/* Rest of styles unchanged */
.mahjor {
  text-align: center;
  height: 100vh;
}
.mahjor-h {
  line-height: 14vh;
  color: transparent;
  -webkit-text-stroke: 0.3vh inherit;
  font-size: 10vh;
  transition: 0.5s;
}
.mahjor-h.active {
  color: inherit;
  -webkit-text-stroke: 0.3vh inherit;
}
.mahjor-content {
  background-color: antiquewhite;
  border-radius: 3vh;
  margin: 3vh 10vw 3vh 10vw;
  height: 80vh;
  width: 80vw;
  transform: translateY(20vh);
  transform: scale(0.9);
  transition: 0.2s;
}
.mahjor-content.active {
  transform: translateY(0) scale(1);
}
.ranking {
  text-align: center;
  height: 100vh;
}
.ranking-h {
  line-height: 14vh;
  color: transparent;
  -webkit-text-stroke: 0.3vh inherit;
  font-size: 10vh;
  transition: 0.5s;
}
.ranking-h.active {
  color: inherit;
  -webkit-text-stroke: 0.3vh inherit;
}
.ranking-content {
  background-color: whitesmoke;
  border-radius: 3vh;
  margin: 3vh 10vw 3vh 10vw;
  height: 80vh;
  width: 80vw;
  transform: translateY(20vh);
  transform: scale(0.9);
  transition: 0.2s;
}
.ranking-content.active {
  transform: translateY(0) scale(1);
}
.ranking-list {
  padding: 0;
}

.ranking-item {
  background-color: unset;
  display: flex;
  list-style-type: none;
  font-size: 3vh;
  line-height: 4vh;
  background-color: inherit;
}
.ranking-item.first {
  font: bold;
  font-size: 4vh;
  line-height: 6vh;
  border-radius: 3vh 3vh 0 0;
}
.ranking-item.first:hover {
  background-color: inherit;
  cursor: default;
}
.ranking-item.Odd {
  background-color: snow;
}
.ranking-item:hover {
  background-color: white;
  cursor: pointer;
}
.ranking-index {
  width: 5%;
  text-align: center;
}
.ranking-name {
  width: 10%;
  text-align: left;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.ranking-freq {
  width: 10%;
  text-align: right;
}
.ranking-rate {
  width: 14%;
  text-align: right;
}
.ranking-perf {
  width: 19%;
  text-align: center;
}

@media (prefers-color-scheme: dark) {
  .enter.active {
    color: white;
  }
  .mahjor-content {
    background-color: darkslategray;
  }
  .ranking-content {
    background-color: dimgray;
  }
  .ranking-item.Odd {
    background-color: gray;
  }
  .ranking-item:hover {
    background-color: rgb(24, 24, 24);
    cursor: pointer;
  }
}
</style>
