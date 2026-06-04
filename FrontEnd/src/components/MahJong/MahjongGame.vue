<!-- 游戏开始后展示此页面 -->
<script setup>
import Chat from './Chat.vue'
import OtherInfo from './OtherInfo.vue'
import MyTiles from './MyTiles.vue'
import { ref, nextTick, computed, watch } from 'vue'
import { statusStore } from '@/stores/status'

const status = statusStore()

const showVideo = ref(false)
const videoPlayer = ref(null)
const focusurl = ref('')

const onVideoEnded = () => {
  showVideo.value = false
}

const FocusAni = async (organization, charactername) => {
  if (!status.roomRule.ShowCritical) return

  const critical = status.getCharacterCritical(organization, charactername)
  if (!critical || !critical.toLowerCase().endsWith('.webm')) return

  focusurl.value = critical
  try {
    const response = await fetch(focusurl.value, { method: 'HEAD' })
    const contentType = response.headers.get('Content-Type')
    if (response.ok && contentType && contentType.startsWith('video')) {
      console.log('video exists:', focusurl.value)
      showVideo.value = true
      await nextTick()
      videoPlayer.value?.play()
    } else {
      console.error(
        'Video not found or incorrect content type:',
        focusurl.value,
        'Content-Type:',
        contentType,
      )
    }
  } catch (error) {
    console.error('Error checking video existence:', error)
  }
}

const otherPlayerSids = computed(() => status.getOtherPlayerSidsBySeat())
const chowChoices = computed(() => {
  const chow = status.getGameInfo().actions?.chow
  return Array.isArray(chow) ? chow : []
})
const tileColor = computed(() =>
  window.matchMedia('(prefers-color-scheme: dark)').matches ? 'Black' : 'Regular',
)
const pendingChowIndex = ref(-1)

const allowOtherInfo = computed(() => (status.players ? true : false))
const allowMytiles = computed(() => status.getGameInfo().init)

watch(
  () => status.actionEvent,
  (event) => {
    if (!event) return
    const member = status.getInfoBySid(event.uuid)
    const organization = member?.decorator?.org ?? ''
    const charactername = member?.decorator?.chara ?? ''
    FocusAni(organization, charactername)
  },
)

function tileSortValue(tile) {
  const matched = /^(Man|Pin|Sou)([1-9])$/.exec(tile)
  if (!matched) return 100
  const suitBase = { Man: 0, Pin: 10, Sou: 20 }
  return suitBase[matched[1]] + Number(matched[2])
}

function chowTiles(choice) {
  if (!choice) return []
  return [...(choice.tiles ?? []), choice.tile]
    .filter(Boolean)
    .sort((a, b) => tileSortValue(a) - tileSortValue(b))
}

async function submitChow(choice, index) {
  if (!choice || pendingChowIndex.value !== -1) return
  pendingChowIndex.value = index
  try {
    await status.apiFetch('/gaming/chow', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        PlayerIndex: status.myid,
        Selec: choice.selec ?? choice.tiles ?? [],
      }),
    })
    status.reSetActions()
  } finally {
    pendingChowIndex.value = -1
  }
}
</script>

<template>
  <div class="first-half">
    <Chat />
    <div class="table-side">
      <div class="other-players">
        <template v-if="allowOtherInfo">
          <OtherInfo
            @click-head="FocusAni"
            v-for="sid in otherPlayerSids"
            :key="sid"
            :sid="sid"
          />
        </template>
      </div>
      <div class="chow-choices">
        <div
          v-for="(_, index) in otherPlayerSids"
          :key="`chow-${index}`"
          class="chow-slot"
        >
          <button
            v-if="chowChoices[index]"
            class="chow-choice"
            :disabled="pendingChowIndex !== -1"
            type="button"
            @click="submitChow(chowChoices[index], index)"
          >
            <span
              class="chow-tile"
              :class="{ eaten: tile === chowChoices[index].tile }"
              v-for="tile in chowTiles(chowChoices[index])"
              :key="tile"
            >
              <img :src="`tilesvgs/${tileColor}/${tile}.svg`" :alt="tile" />
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
  <MyTiles @click-head="FocusAni" v-if="allowMytiles" />
  <video
    v-if="showVideo"
    ref="videoPlayer"
    :src="`${focusurl}`"
    @ended="onVideoEnded"
    autoplay
  ></video>
</template>
<style>
video {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100vw;
  height: 100vh;
  object-fit: cover;
  z-index: 9999;
}
.first-half {
  display: flex;
  height: 60vh;
}
.table-side {
  height: 60vh;
  margin: 0 1vw;
  display: flex;
  flex-direction: column;
}
.other-players {
  height: 39vh;
  display: flex;
  gap: 1vw;
  align-items: flex-start;
}
.chow-choices {
  height: 21vh;
  display: flex;
  gap: 1vw;
  align-items: flex-start;
}
.chow-slot {
  width: 26vw;
  height: 21vh;
}
.chow-choice {
  width: 26vw;
  height: 18vh;
  box-sizing: border-box;
  padding: 1.2vh 1vw;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.8vw;
  border: 0.4vh solid #111;
  border-radius: 3vh;
  cursor: pointer;
}
.chow-choice:hover {
  background: #ffeb3b;
}
.chow-choice:disabled {
  cursor: wait;
  opacity: 0.7;
}
.chow-tile {
  width: 5vw;
  height: 13vh;
  box-sizing: border-box;
  padding: 0.2vw;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0.35vw solid #111;
  border-radius: 0.3vw;
  box-shadow: none;
}
.chow-tile.eaten {
  background: #ffc107;
  border-width: 0.45vw;
}
.chow-tile img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>
