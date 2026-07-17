<script setup lang="ts">
import { computed, ref } from 'vue'
import { statusStore, type ReplayEvent } from '@/stores/status'
import { reconstructReplay, replayActions } from '@/lib/replayEngine'
import { tilesmapStore } from '@/stores/tilesmap'

const status = statusStore()
const tileMap = tilesmapStore()
const replay = computed(() => status.replayData!)
const step = ref(0)
const actions = computed(() => replayActions(replay.value))
const total = computed(() => actions.value.length)
const state = computed(() => reconstructReplay(replay.value, step.value))
const players = computed(() => state.value.players)
const visibleActions = computed(() => actions.value.slice(0, step.value))
const recentActions = computed(() => visibleActions.value.slice(-4).reverse())
const seed = computed(() => replay.value.events.find((event) => event.action === 'Shuffle')?.value.match(/Seed:([^;]+)/)?.[1] ?? '—')
const golden = computed(() => replay.value.events.find((event) => event.action === 'Golden')?.value ?? '')
const gameTime = computed(() => new Date(actions.value[0]?.timestamp ?? replay.value.exported_at).toLocaleString())
const tileColor = computed(() => window.matchMedia('(prefers-color-scheme: dark)').matches ? 'Black' : 'Regular')
const wallRows = computed(() => {
  const remaining = state.value.wall.slice(state.value.wallPosition)
  return [
    Array.from({ length: 9 }, (_, index) => `Man${index + 1}`),
    Array.from({ length: 9 }, (_, index) => `Sou${index + 1}`),
    Array.from({ length: 9 }, (_, index) => `Pin${index + 1}`),
    ['Ton', 'Nan', 'Shaa', 'Pei', 'Chun', 'Hatsu', 'Haku'],
  ].map((tiles) => tiles.map((tile) => ({
    tile,
    count: remaining.filter((remainingTile) => remainingTile === tile).length,
  })))
})

function playerName(uuid: string) { return players.value.find((player) => player.uuid === uuid)?.name ?? uuid }
function actionName(action: string) { return ({ Discard: '出牌', Chow: '吃', Pong: '碰', Kong: '杠', Hu: '胡' } as Record<string, string>)[action] ?? action }
function eventTiles(event: ReplayEvent) { return event.value.split(',').filter((tile) => tileMap.fontsmap[tile] || tile === 'Golden') }
function eventText(event: ReplayEvent) {
  if (event.action === 'Hu') return event.user === event.value ? '自摸' : `胡 ${playerName(event.value)}`
  return ''
}
function previous() { if (step.value > 0) step.value -= 1 }
function next() { if (step.value < total.value) step.value += 1 }
</script>

<template>
  <main class="replay-page">
    <section class="players-area">
      <article v-for="(player, playerIndex) in players" :key="player.uuid" class="player-card">
        <img class="avatar" :src="status.getCharacterHead(player.group, player.chara)" :alt="player.chara || player.name" />
        <div class="player-copy" :class="{ current: state.currentPlayer === playerIndex }">
          <h2 :title="player.uuid">{{ player.name }}</h2>
          <p>{{ player.group || 'nameless' }}/{{ player.chara || '佚名' }}</p>
          <div class="tile-line">
            <span class="tile-group"><span class="svg-tile locked" v-for="(tile, index) in player.locked" :key="`l-${step}-${index}`"><img :src="`tilesvgs/${tileColor}/${tile}.svg`" /></span></span>
            <span class="tile-group"><span class="svg-tile" v-for="(tile, index) in player.hands" :key="`h-${step}-${index}`"><img :src="`tilesvgs/${tileColor}/${tile}.svg`" /></span></span>
            <span class="tile-group new-group"><span v-if="player.newTile" class="svg-tile new" :key="`n-${step}`"><img :src="`tilesvgs/${tileColor}/${player.newTile}.svg`" /></span></span>
          </div>
        </div>
      </article>
    </section>

    <aside class="replay-actions">
      <button class="side-button" @click="status.closeReplay()">退出录像</button>
      <dl>
        <div><dt>对局时间</dt><dd>{{ gameTime }}</dd></div>
        <div><dt>随机数种子</dt><dd>{{ seed }}</dd></div>
        <div class="golden-row"><dd><span class="golden-tile"><img v-if="golden" :src="`tilesvgs/${tileColor}/${golden}.svg`" title="本局金牌" /></span></dd></div>
      </dl>

      <section class="recent-list">
        <h3>最近 4 条操作</h3>
        <p v-if="recentActions.length === 0">尚未开始</p>
        <article v-for="event in recentActions" :key="event.seq">
          <div class="log-copy"><b>{{ playerName(event.user) }}</b><span>{{ actionName(event.action) }}</span></div>
          <div class="log-value">
            <template v-if="eventTiles(event).length">
              <img v-for="(tile, index) in eventTiles(event).filter((tile) => tile === 'Golden')" :key="`g-${index}`" :src="`tilesvgs/${tileColor}/${tile}.svg`" title="本局金牌" />
              <i v-for="(tile, index) in eventTiles(event).filter((tile) => tile !== 'Golden')" :key="`f-${index}`">{{ tileMap.getTileFont(tile) }}</i>
            </template>
            <small v-else>{{ eventText(event) }}</small>
          </div>
        </article>
      </section>

      <section class="wall-panel">
        <div v-for="(row, rowIndex) in wallRows" :key="rowIndex" class="wall-row">
          <div class="wall-faces" :style="{ gridTemplateColumns: `repeat(${row.length}, minmax(0, 1fr))` }"><i v-for="entry in row" :key="entry.tile" :class="{ empty: entry.count === 0 }">{{ tileMap.getTileFont(entry.tile) }}</i></div>
          <div class="wall-counts" :style="{ gridTemplateColumns: `repeat(${row.length}, minmax(0, 1fr))` }"><i v-for="entry in row" :key="entry.tile" :class="{ empty: entry.count === 0 }">{{ entry.count }}</i></div>
        </div>
      </section>

      <footer>
        <button class="step-button" :disabled="step === 0" @click="previous">←</button>
        <strong>{{ step }}/{{ total }}</strong>
        <button class="step-button" :disabled="step === total" @click="next">→</button>
      </footer>
    </aside>
  </main>
</template>

<style scoped>
@font-face { font-family: 'Segoe UI Symbol'; src: url('/src/assets/fonts/segoe-ui-symbol.ttf') format('truetype'); }
* { box-sizing: border-box; }
.replay-page { display: grid; grid-template-columns: 80% 20%; width: 100%; height: 100vh; background: #f0f0f0; color: #111; overflow: hidden; }
.players-area { display: grid; grid-template-rows: repeat(4, minmax(0, 1fr)); gap: 1.3vh; padding: 1.5vh 1.5vw; }
.player-card { display: grid; grid-template-columns: 18vh 1fr; min-height: 0; border: .45vh solid #111; border-radius: .5vh; background: #fff; box-shadow: .5vh .5vh 0 #111; overflow: hidden; }
.avatar { width: 18vh; height: 100%; object-fit: cover; border-right: .4vh solid #111; }
.player-copy { position: relative; display: grid; grid-template-rows: 30% 20% 50%; min-width: 0; padding: .8vh 1vw; overflow: hidden; background: #fff; transition: background-color .22s ease-out; }
.player-copy.current { background: #ffd84d; }
.player-copy h2 { display: flex; align-items: center; margin: 0; font-size: 3.5vh; line-height: 1; cursor: help; }
.player-copy p { display: flex; align-items: center; margin: 0; font-size: 2.35vh; line-height: 1; font-weight: 800; }
.tile-line { display: flex; align-items: center; min-width: 0; overflow: hidden; gap: .75vw; }
.tile-group { display: flex; align-items: center; flex: 0 1 auto; min-width: 0; }
.new-group { flex: 0 0 auto; }
.svg-tile { flex: 0 0 2.8vw; width: 3.5vw; height: 7.2vh; padding: .08vw; border: .14vw solid #111; border-radius: .18vw; background: #fff; margin-right: .1vw; }
.svg-tile img { width: 100%; height: 100%; object-fit: contain; }
.svg-tile.locked { border-color: #555; }
.svg-tile.new { margin-left: .4vw; border-color: #f5a400; box-shadow: .18vh .18vh 0 #f5a400; }
.replay-actions { display: flex; flex-direction: column; gap: 1vh; min-width: 0; padding: 1.5vh 1vw; border-left: .5vh solid #111; background: #fff; }
.side-button, .step-button { border: .35vh solid #111; border-radius: .35vh; background: #fff; box-shadow: .3vh .3vh 0 #111; padding: .8vh; font-weight: 900; cursor: pointer; }
dl { display: grid; grid-template-columns: 1fr; gap: .7vh; margin: 0; }
dt { font-size: 1.25vh; color: #666; font-weight: 800; } dd { margin: .1vh 0 0; font-size: 1.55vh; font-weight: 900; overflow-wrap: anywhere; }
.golden-row dd { display: flex; }.golden-tile { width: 3.6vw; height: 7vh; padding: .12vw; border: .24vw solid #111; border-radius: .25vw; background: #fff; box-shadow: .25vh .25vh 0 #111; }.golden-tile img { width: 100%; height: 100%; object-fit: contain; cursor: help; }
.recent-list { display: grid; grid-template-rows: auto repeat(4, 6.4vh); height: 28.2vh; min-height: 28.2vh; overflow: hidden; }.recent-list h3, .wall-panel h3 { margin: 0 0 .5vh; font-size: 1.7vh; }
.recent-list > p { grid-row: 2 / 6; margin: 0; display: flex; align-items: center; }
.recent-list article { display: grid; grid-template-columns: minmax(0, 42%) minmax(0, 58%); min-height: 0; border-top: .18vh solid #ccc; }
.log-copy { display: grid; grid-template-rows: 1fr 1fr; min-width: 0; padding: .45vh .35vw 0 0; }
.log-copy b, .log-copy span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }.log-copy b { align-self: end; font-size: 1.5vh; }.log-copy span { align-self: start; font-size: 1.35vh; color: #555; font-weight: 800; }
.log-value { display: flex; align-items: center; justify-content: flex-end; height: 100%; min-width: 0; font-family: 'Segoe UI Symbol'; white-space: nowrap; overflow: hidden; }
.log-value i { font-style: normal; font-size: 5.5vh; line-height: 1; }.log-value img { width: 3.1vw; height: 90%; object-fit: contain; vertical-align: middle; }.log-value small { font-size: 1.4vh; font-family: inherit; }
.wall-panel { flex: 1; display: flex; flex-direction: column; justify-content: space-evenly; min-height: 0; padding-top: .4vh; border-top: .3vh solid #111; overflow: hidden; }
.wall-row { width: 100%; overflow: hidden; }.wall-faces, .wall-counts { display: grid; width: 100%; font-family: 'Segoe UI Symbol'; text-align: center; }.wall-faces i { font-style: normal; font-size: clamp(2.4rem, 3.15vw, 4rem); line-height: .9; color: #111; }.wall-counts i { font-style: normal; font-family: ui-monospace, monospace; font-size: clamp(1rem, 1.35vw, 1.6rem); line-height: 1; font-weight: 900; color: #111; }.wall-faces i.empty, .wall-counts i.empty { color: #c5c5c5; }
footer { display: grid; grid-template-columns: 1fr auto 1fr; align-items: center; gap: .6vw; } .step-button:disabled { opacity: .35; cursor: default; } footer strong { white-space: nowrap; font-size: 1.6vh; }
@media (max-width: 900px) { .replay-page { grid-template-columns: 75% 25%; } .player-card { grid-template-columns: 12vh 1fr; } .avatar { width: 12vh; } .svg-tile { flex-basis: 2.8vw; width: 2.8vw; } }
</style>
