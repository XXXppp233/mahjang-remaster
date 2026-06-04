<template>
  <div class="other-info" :class="{ compact }">
    <div class="box" :class="{ active: isActive, compact }">
      <div class="summary-row">
        <div class="playerhead">
          <button class="test-button" @click.ctrl.exact="clicktestbutton">
            <img class="characterhead" :src="`${url}`" :alt="`${charactername}`" />
          </button>
        </div>
        <div class="playerinfo">
          <h3>{{ memberinfo.name }}</h3>
          <h4>{{ charactername }}, {{ organization }}</h4>
          <p class="score">Score {{ score }}</p>
        </div>
      </div>
      <div v-if="isGaming && !compact" class="handstiles">
        <span><b v-for="tile in locked">{{ tile }}</b></span>
        <span><b v-for="tile in Array(handsnumber).fill('🀫')">{{ tile }}</b></span>
        <span><b v-if="hasnew">🀫</b></span>
      </div>
    </div>
    <div v-if="isGaming && !compact" class="discarded-panel">
      <div class="discardedtiles">
        <b
          class="discardedtile"
          :class="{ new: index > shownnum - 1 }"
          v-for="(tile, index) in discarded"
          >{{ tile }}</b
        >
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { tilesmapStore } from '@/stores/tilesmap'
import { statusStore } from '@/stores/status'

const status = statusStore()
const tilesmap = tilesmapStore()
const props = defineProps({
  sid: {
    type: String,
    required: true,
  },
  compact: {
    type: Boolean,
    default: false,
  },
})

const memberinfo = computed(() => status.getInfoBySid(props.sid))
const playerinfo = computed(() => status.getPlayerBySid(props.sid))
const isActive = computed(() => Boolean(playerinfo.value?.active))
const isGaming = computed(() => status.now === 'gaming')
// Room
const organization = computed(() => (memberinfo.value.ready ? memberinfo.value.decorator.org : ''))
const charactername = computed(() => (memberinfo.value.ready ? memberinfo.value.decorator.chara : ''))

// Gaming
// locked discarded 转换为字符后输出
const locked = computed(() =>
  playerinfo.value ? tilesmap.getTilesFont(playerinfo.value.locked) : [],
)
const handsnumber = computed(() => (playerinfo.value ? playerinfo.value.hand_count : 0))
const hasnew = computed(() => (playerinfo.value ? playerinfo.value.hasnew : false))
const score = computed(() => playerinfo.value?.score ?? memberinfo.value?.score ?? 0)
const discarded = computed(() =>
  playerinfo.value ? tilesmap.getTilesFont(playerinfo.value.discarded) : [],
)
const shownnum = ref(0)
watch(discarded, (newVal, oldVal) => {
  if (newVal.length === oldVal.length) {
    // 牌数没变
    return
  } else if (newVal.length >= oldVal.length) {
    // 出牌
    console.log('牌数增加')
    setTimeout(() => {
      shownnum.value = newVal.length
    }, 2000) // 2秒后取消特写
    //shownnum.value += 1
  } else {
    // 牌被杠走
    console.log('牌数减少')
    shownnum.value = newVal.length
  }
})

const url = computed(() =>
  memberinfo.value.ready
    ? status.getCharacterHead(organization.value, charactername.value)
    : 'tilesvgs/Regular/Blank.svg',
)

//后续会改成点击头像会触发一个语音，这里为了测试改为触发特写
const emit = defineEmits(['click-head'])
const clicktestbutton = () => {
  emit('click-head', organization.value, charactername.value)
}

</script>

<style scoped>
@font-face {
  font-family: 'Segoe UI Symbol';
  src: url('/src/assets/fonts/segoe-ui-symbol.ttf') format('truetype');
}
.box {
  z-index: 1;
  caret-color: transparent;
  display: flex;
  flex-direction: column;
  width: 26vw;
  height: 19vh;
  background: #fff;
  border: 0.4vh solid #111;
  border-radius: 3vh;
  box-shadow: 0.6vh 0.6vh 0 #111;
  padding: 1.5vh 1.5vh 0 1.5vh;
  gap: 0.8vh;
}
.box.compact {
  flex-direction: row;
  width: 16vw;
  min-width: 12rem;
  height: 15vh;
  padding: 1vh;
  gap: 1vw;
}
.box.active {
  background-color: #ffeb3b; /* 明黄色的激活状态 */
  transform: translate(-0.2vh, -0.2vh);
  box-shadow: 0.8vh 0.8vh 0 #111;
}
.other-info {
  width: 26vw;
}
.other-info.compact {
  width: auto;
}
.summary-row {
  display: flex;
  gap: 1vw;
  min-height: 5vw;
}
.box.compact .summary-row {
  display: contents;
}
.test-button {
  border: 0.3vh solid #111;
  height: 100%;
  width: 100%;
  padding: 0;
  margin: 0;
  border-radius: 0.3vh;
  background-color: #fff;
  overflow: hidden;
  box-shadow: 0.3vh 0.3vh 0 #111;
}
.playerhead {
  width: 5vw;
  height: 5vw;
  min-width: 4rem;
  min-height: 4rem;
  flex-shrink: 0;
}
.box.compact .playerhead {
  width: 5vw;
  height: 5vw;
  min-width: 4rem;
  min-height: 4rem;
}
.characterhead {
  width: 100%;
  height: 100%;
  object-fit: cover;
  image-rendering: auto;
}

.playerinfo {
  flex: 1;
  display: flex;
  flex-direction: column;
  line-height: 1.2;
  min-width: 0;
}
h3 {
  cursor: pointer;
  margin: 0;
  font-size: 2.2vh;
  font-weight: 800;
  color: #111;
  text-decoration: underline 0.3vh #111;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.box.compact h3 {
  font-size: 2.2vh;
  text-decoration-thickness: 0.2vh;
}
h3:hover {
  color: #000;
  background: #eee;
}
h4 {
  cursor: pointer;
  margin: 0.35vh 0;
  font-size: 1.5vh;
  font-weight: 700;
  color: #444;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.box.compact h4 {
  font-size: 1.5vh;
  margin: 0.4vh 0 0;
}
.handstiles {
  display: flex;
  flex-wrap: nowrap;
  gap: 0.5vw;
  margin-top: auto;
  overflow: hidden;
  white-space: nowrap;
}
.score {
  margin: 0;
  font-size: 1.45vh;
  font-weight: 800;
  color: #111;
}
.box.compact .score {
  font-size: 1.4vh;
}
.discarded-panel {
  margin-top: 1vh;
  width: 26vw;
  min-height: 8vh;
  border-radius: 0.5vh;
  background-color: transparent;
  border: none;
  box-shadow: none;
  padding: none;
}
.discardedtiles {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25vw;
  max-height: 11vh;
  overflow-y: auto;
}
b {
  font-family: 'Segoe UI Symbol';
  font-size: 3vh;
  font-weight: normal;
}
b:hover {
  cursor: pointer;
  color: #f44336;
}
.discardedtile{
  transition: all 0.3s;
}
.discardedtile.new {
  font-size: 8vh;
}
</style>
