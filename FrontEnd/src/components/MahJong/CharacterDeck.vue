<template>
  <div class="box">
    <img class="selectedimg" :src="headurl" alt="" v-if="isReady" />
    <div class="characters" v-if="!isReady">
      <div class="controls">
        <div class="grid" v-if="characters.length">
          <button
            v-for="character in characters"
            :key="character.name"
            class="thumb"
            :class="{ active: character.name === status.tempSelect }"
            @click="status.setTempSelect(character.name)"
            type="button"
            title="选择图片"
          >
            <div class="thumb-img">
              <img :src="character.head" :alt="character.name" />
            </div>
            <div class="caption">{{ character.name }}</div>
          </button>
        </div>
      </div>
    </div>
    <div class="myinfo" v-if="isReady">
      <h3>{{ myname }}</h3>
      <h4>{{ charactername }}, {{ orgnaization }}</h4>
      <p>Score {{ score }}</p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { statusStore } from '@/stores/status'

const status = statusStore()

const myinfo = computed(() => status.getInfoBySid(status.mysid))
const myname = computed(() => (myinfo.value ? myinfo.value.name : ''))
const orgnaization = computed(() => (myinfo.value?.decorator ? myinfo.value.decorator.org : ''))
const charactername = computed(() => (myinfo.value?.decorator ? myinfo.value.decorator.chara : ''))
const isReady = computed(() => (myinfo.value ? myinfo.value.ready : false))
const score = computed(() =>
  status.isGaming ? status.getGameInfo().score : (myinfo.value?.score ?? 0),
)
const headurl = computed(() =>
  isReady.value ? status.getCharacterHead(orgnaization.value, charactername.value) : 'tilesvgs/Regular/Blank.svg',
)

const characters = computed(() => status.getCharacters(status.tempOrg))
</script>

<style scoped>
@media (prefers-color-scheme: dark) {
  .characters {
    background-color: rgb(24, 24, 24) !important;
  }
}

.box {
  z-index: 1;
  display: flex;
  margin: 0vh 1vw;
  caret-color: transparent;
  gap: 2vw;
}
.selectedimg {
  width: 20vh;
  height: 20vh;
  margin-top: 5vh;
  border: 0.5vh solid #111;
  border-radius: 0.5vh;
  object-fit: cover;
  box-shadow: 0.6vh 0.6vh 0 #111;
}
.characters {
  z-index: 1;
  border: 0.5vh solid #111;
  border-radius: 0.5vh;
  width: 50vw;
  height: 35vh;
  box-sizing: border-box;
  overflow: hidden;
  background-color: #fff;
  box-shadow: 0.8vh 0.8vh 0 #111;
  padding: 1.5vh;
}
.controls {
  display: flex;
  flex-direction: column;
  gap: 1.5vh;
  height: 100%;
}
.grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5vh;
  overflow-y: auto;
  padding-top: 1vh;
}
.thumb {
  display: flex;
  flex-direction: column;
  border: 0.3vh solid #111;
  background-color: #fff;
  padding: 0.5vh;
  cursor: pointer;
  height: 28vh;
  border-radius: 0.3vh;
  transition: all 0.1s;
}
.thumb:hover {
  background-color: #f0f0f0;
  transform: translate(-0.1vh, -0.1vh);
  box-shadow: 0.3vh 0.3vh 0 #111;
}
.thumb.active {
  background-color: #ffeb3b;
  border-width: 0.4vh;
  box-shadow: 0.4vh 0.4vh 0 #111;
}
.thumb-img {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-bottom: 0.2vh solid #111;
  background: #fafafa;
}
.thumb img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.caption {
  font-size: 1.8vh;
  font-weight: 800;
  text-align: center;
  padding-top: 0.5vh;
}
.myinfo {
  z-index: 1;
  display: flex;
  flex-direction: column;
  margin-top: 5vh;
  gap: 1vh;
}

h3 {
  margin: 0;
  font-size: 5vh;
  font-weight: 900;
  color: #111;
  text-decoration: underline 0.5vh #111;
}
h4 {
  margin: 0;
  font-size: 2.4vh;
  font-weight: 700;
  color: #444;
}
p {
  margin: 0;
  font-size: 2vh;
  font-weight: 800;
  color: #111;
}
</style>
