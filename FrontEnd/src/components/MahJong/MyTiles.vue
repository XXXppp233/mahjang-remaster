<script setup>
import MahjongTile from './MahjongTile.vue'
import ActionButton from './ActionButton.vue'
import DiscardButton from './DiscardButton.vue'
import DiscardTiles from './DiscardTiles.vue'
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { statusStore } from '@/stores/status'
import { tilesmapStore } from '@/stores/tilesmap'
import keySounds from '/src/assets/keyboardsounds/MikuTap/main.js'

const status = statusStore()
const tilesmap = tilesmapStore()
// const keySoundType = ref('MikuTap') // default value
// const keyboardsoundsurl = ref('/src/assets/keyboardsounds/MikuTap/main.js')


//const actionNames = ref(['吃 🀙🀚', '吃 🀚🀜', '吃 🀜🀝', '碰 🀛', '杠 🀛', '胡 🀛'])
const username = computed(() => status.username)
const mysid = computed(() => status.mysid)
const myinfo = computed(() => status.getInfoBySid(mysid.value))
const mygameinfo = computed(() => status.getGameInfo())
const isActive = computed(() => mygameinfo.value.activePlayer === mygameinfo.value.id) // 是否应该出牌

const discarded = computed(() => tilesmap.getTilesFont(mygameinfo.value.discarded))
const hands = computed(() => tilesmap.getTilesName(mygameinfo.value.hands))
const newtile = computed(() => tilesmap.getTileName(mygameinfo.value.new))
const locked = computed(() => tilesmap.getTilesName(mygameinfo.value.locked))
const score = computed(() => mygameinfo.value.score ?? 0)
const actionEntries = computed(() => {
  const names = tilesmap.getActionsName(mygameinfo.value.actions)
  const data = tilesmap.getActionData(mygameinfo.value.actions)
  return names
    .map((name, index) => ({ name, data: data[index] }))
    .filter((entry) => !entry.name.startsWith('吃'))
})
const actionsName = computed(() => actionEntries.value.map((entry) => entry.name))
const actionsData = computed(() => actionEntries.value.map((entry) => entry.data))
const actionsnumber = computed(() => actionEntries.value.length)

// 游戏开始后早已初始化基本数据
const organization = computed(() => (myinfo.value?.decorator ? myinfo.value.decorator.org : ''))
const charactername = computed(() => (myinfo.value?.decorator ? myinfo.value.decorator.chara : ''))
const headurl = computed(() => status.getCharacterHead(organization.value, charactername.value))
// 有新牌时默认选中新牌，否则默认选中第一张牌
const selectedIndex = ref(0) // 0~15 手牌 16 新牌
// 轮到我出牌时，如果所有牌都未选中且有新牌则选中新牌，否则选中第一张牌
watch(isActive, (newval) => {
  if (newval) {
    if (selectedIndex.value >= hands.value.length && newtile.value) {
      selectedIndex.value = hands.value.length
    } else if (selectedIndex.value >= hands.value.length) {
      selectedIndex.value = 0
    }
  } else {
  }
})
const select = (index) => {
  if (index === selectedIndex.value) {
    selectedIndex.value = hands.value.length // 再次点击取消选择，选中新牌位置（如果有新牌）
  } else if (index >= hands.value.length) {
    selectedIndex.value = hands.value.length // 选中新牌位置
  } else {
    selectedIndex.value = index // 选中手牌
  }
}
const selectAction = ref(6) // 6 为无效选择，0~5 为吃碰杠胡等操作

// 按键监听
const handleKeyPress = (event) => {
  // 只在 isTyping 为 false 时监听按键
  if (status.isTyping) return

  if (event.key.toLowerCase() === 'q') {
    // 在这里添加 Q 键的处理逻辑
    // 例如：选择上一张牌
    if (selectedIndex.value > 0) {
      selectedIndex.value -= 1
    }
    keySounds.Play(event) // 播放 Q 键音效
  } else if (event.key.toLowerCase() === 'e') {
    // 在这里添加 E 键的处理逻辑
    // 例如：选择下一张牌
    const maxIndex = newtile.value ? hands.value.length : hands.value.length - 1
    if (selectedIndex.value < maxIndex) {
      selectedIndex.value += 1
    }
    keySounds.Play(event) // 播放 E 键音效
  }
  else if (event.keyCode >= 49 && event.keyCode <= 54) { // 49-54 是 '1'-'6' 的 keyCode
    const actionIndex = event.keyCode - 49 // 转换为 0-5 的索引
    if (actionsnumber.value === 0) return // 没有操作时不响应数字键
    if (actionIndex < actionsnumber.value) {
      if(selectAction.value === actionIndex){
        selectAction.value = 6 // 再次按相同数字键取消选择
      }else{
        selectAction.value = actionIndex
      } // 0-5 的选择
      // 在这里添加对应操作的处理逻辑
    }
    keySounds.Play(event) // 播放数字键音效
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeyPress)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyPress)
})

</script>

<template>
  <div class="box" :class="{ active: isActive }">
    <div class="my-dashboard">
      <div class="my-profile">
        <div class="avatar-frame">
          <img
            class="characterhead"
            :src="headurl"
            :alt="`${charactername}`"
          />
        </div>
        <div class="profile-text">
          <h3>{{ username }}</h3>
          <h4>{{ charactername }}, {{ organization }}</h4>
          <p>Score {{ score }}</p>
        </div>
      </div>
      <div id="my-actions" class="my-center">
        <div class="action-buttons">
          <ActionButton
            v-for="(action, index) in actionsName"
            :label="action"
            :actionid="index"
            :selected="selectAction === index"
            :length="actionsnumber"
            :data="actionsData[index]"
          />
        </div>
        <DiscardTiles class="discardedtiles" :discarded="discarded" />
      </div>
      <div class="discard-command">
        <DiscardButton :selectedIndex="selectedIndex" :active="isActive" />
      </div>
    </div>

    <div id="my-tiles" class="my-tiles">
      <span>
        <MahjongTile
          v-if="locked.length"
          v-for="(tile, index) in locked"
          :index="index"
          :tile="tile"
          :locked="true"
        />
      </span>
      <span>
        <MahjongTile
          v-for="(tile, index) in hands"
          :index="index"
          :tile="tile"
          :selected="selectedIndex === index"
          @click="select"
        />
      </span>
      <span>
        <MahjongTile
          v-if="newtile"
          :tile="newtile"
          :index="hands.length"
          :selected="selectedIndex === hands.length"
          @click="select"
        />
      </span>
    </div>
  </div>
</template>

<style scoped>
@font-face {
  font-family: 'Segoe UI Symbol';
  src: url('/src/assets/fonts/segoe-ui-symbol.ttf') format('truetype');
}
.box {
  z-index: 1;
  padding: 1.5vh 2vh;
  border-radius: 1vh;
  transition: all 0.3s;
}
.box.active {
  background-color: #ffeb3b; /* 明黄色激活状态 */
  border: 0.5vh solid #111;
  box-shadow: 0.8vh 0.8vh 0 #111;
}
/* span 元素目前仅对下方自己的手排生效 */
span {
  z-index: 1;
  display: flex;
  flex-wrap: nowrap;
  gap: 0.2vw;
}

.my-dashboard {
  z-index: 1;
  display: grid;
  grid-template-columns: 24vw minmax(45vw, 1fr) 9vw;
  gap: 1vw;
  align-items: stretch;
}
.avatar-frame {
  z-index: 1;
  border: 0.4vh solid #111;
  height: 100%;
  width: 100%;
  padding: 0;
  border-radius: 0.5vh;
  background-color: #fff;
  overflow: hidden;
  box-shadow: 0.5vh 0.5vh 0 #111;
}

.my-profile {
  z-index: 1;
  caret-color: transparent;
  display: grid;
  grid-template-columns: 6.5vw minmax(0, 1fr);
  grid-template-rows: repeat(3, 1fr);
  gap: 0 1vw;
  min-height: 16vh;
  align-items: center;
}
.my-profile .avatar-frame {
  grid-row: 1 / 4;
  width: 6.5vw;
  height: 6.5vw;
  align-self: center;
}
.characterhead {
  z-index: 1;
  width: 100%;
  height: 100%;
  object-fit: cover;
  image-rendering: auto;
}

.profile-text {
  grid-column: 2;
  grid-row: 1 / 4;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

h3 {
  z-index: 1;
  cursor: pointer;
  margin: 0;
  font-size: 3.6vh;
  font-weight: 900;
  text-decoration: underline 0.4vh #111;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
h3:hover {
  background: rgba(0,0,0,0.05);
}
h4 {
  z-index: 1;
  cursor: pointer;
  margin: 0.35vh 0;
  font-size: 2.2vh;
  font-weight: 700;
  color: #444;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
p {
  z-index: 1;
  margin: 0;
  font-size: 2vh;
  font-weight: 800;
  color: #111;
}
.my-center {
  min-width: 0;
  display: grid;
  grid-template-rows: 6vh minmax(8vh, 1fr);
  gap: 1vh;
}
.action-buttons {
  z-index: 1;
  display: flex;
  flex-wrap: nowrap;
  gap: 0.5vw;
  min-width: 0;
  overflow: hidden;
}
.discard-command {
  display: flex;
  align-items: stretch;
  justify-content: center;
  min-width: 0;
}

.my-tiles {
  z-index: 1;
  gap: 1.5vw;
  margin: 2vh 0;
  display: flex;
}
</style>
