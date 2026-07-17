<script setup lang="ts">
import { computed, ref } from 'vue'
import { statusStore, type ReplayEvent } from '@/stores/status'

const status = statusStore()
const replay = computed(() => status.replayData!)
const step = ref(0)

const players = computed(() => replay.value.events.filter((event) => event.action === 'Ready').map((event) => {
  const separator = event.value.indexOf(':')
  const name = separator >= 0 ? event.value.slice(0, separator) : event.user
  const character = separator >= 0 ? event.value.slice(separator + 1) : ''
  const slash = character.indexOf('/')
  return {
    uuid: event.user,
    name,
    group: slash >= 0 ? character.slice(0, slash) : '',
    chara: slash >= 0 ? character.slice(slash + 1) : character,
  }
}))

const actions = computed(() => replay.value.events.filter((event) => !['Ready', 'SetRule', 'Shuffle', 'Golden'].includes(event.action)))
const total = computed(() => actions.value.length)
const visibleActions = computed(() => actions.value.slice(0, step.value))
const recentActions = computed(() => visibleActions.value.slice(-5).reverse())
const seed = computed(() => replay.value.events.find((event) => event.action === 'Shuffle')?.value.match(/Seed:([^;]+)/)?.[1] ?? '—')
const golden = computed(() => replay.value.events.find((event) => event.action === 'Golden')?.value ?? '—')
const gameTime = computed(() => {
  const timestamp = actions.value[0]?.timestamp ?? replay.value.exported_at
  return new Date(timestamp).toLocaleString()
})

function playerName(uuid: string) { return players.value.find((player) => player.uuid === uuid)?.name ?? uuid }
function latestFor(uuid: string): ReplayEvent | undefined { return [...visibleActions.value].reverse().find((event) => event.user === uuid) }
function previous() { if (step.value > 0) step.value -= 1 }
function next() { if (step.value < total.value) step.value += 1 }
</script>

<template>
  <main class="replay-page">
    <section class="players-area">
      <article v-for="player in players" :key="player.uuid" class="player-card">
        <img :src="status.getCharacterHead(player.group, player.chara)" :alt="player.chara || player.name" />
        <div class="player-copy">
          <h2>{{ player.name }}</h2>
          <p>{{ player.group || 'nameless' }} / {{ player.chara || '佚名' }}</p>
          <small>{{ player.uuid }}</small>
          <strong v-if="latestFor(player.uuid)">
            {{ latestFor(player.uuid)?.action }} · {{ latestFor(player.uuid)?.value }}
          </strong>
        </div>
      </article>
    </section>

    <aside class="replay-actions">
      <button class="side-button" @click="status.closeReplay()">退出录像</button>
      <dl>
        <div><dt>对局时间</dt><dd>{{ gameTime }}</dd></div>
        <div><dt>随机数种子</dt><dd>{{ seed }}</dd></div>
        <div><dt>金牌</dt><dd>{{ golden }}</dd></div>
      </dl>

      <section class="recent-list">
        <h3>最近 5 条操作</h3>
        <p v-if="recentActions.length === 0">尚未开始</p>
        <article v-for="event in recentActions" :key="event.seq">
          <b>{{ playerName(event.user) }}</b>
          <span>{{ event.action }}</span>
          <small>{{ event.value }}</small>
        </article>
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
.replay-page { display: grid; grid-template-columns: 80% 20%; width: 100%; height: 100vh; background: #f0f0f0; color: #111; overflow: hidden; }
.players-area { display: grid; grid-template-rows: repeat(4, minmax(0, 1fr)); gap: 1.5vh; padding: 2vh 2vw; }
.player-card { display: grid; grid-template-columns: 18vh 1fr; min-height: 0; border: .5vh solid #111; border-radius: .5vh; background: #fff; box-shadow: .6vh .6vh 0 #111; overflow: hidden; }
.player-card img { width: 18vh; height: 100%; object-fit: cover; border-right: .4vh solid #111; }
.player-copy { display: flex; flex-direction: column; justify-content: center; min-width: 0; padding: 1vh 1.5vw; }
.player-copy h2 { margin: 0; font-size: 3.2vh; }
.player-copy p, .player-copy small { margin: .35vh 0; overflow-wrap: anywhere; }
.player-copy strong { margin-top: .8vh; color: #1677ff; font-size: 2vh; }
.replay-actions { display: flex; flex-direction: column; gap: 1.5vh; padding: 2vh 1.2vw; border-left: .5vh solid #111; background: #fff; }
.side-button, .step-button { border: .4vh solid #111; border-radius: .4vh; background: #fff; box-shadow: .35vh .35vh 0 #111; padding: 1vh; font-weight: 900; cursor: pointer; }
dl { margin: 0; }
dl div { margin-bottom: 1.2vh; }
dt { font-size: 1.5vh; color: #666; font-weight: 800; }
dd { margin: .2vh 0 0; font-size: 1.8vh; font-weight: 900; overflow-wrap: anywhere; }
.recent-list { min-height: 0; flex: 1; overflow: hidden; }
.recent-list h3 { margin: 0 0 1vh; font-size: 2vh; }
.recent-list article { display: grid; grid-template-columns: 1fr auto; gap: .2vh .5vw; padding: .8vh 0; border-top: .2vh solid #ccc; }
.recent-list small { grid-column: 1 / -1; overflow-wrap: anywhere; }
footer { display: grid; grid-template-columns: 1fr auto 1fr; align-items: center; gap: .8vw; }
.step-button:disabled { opacity: .35; cursor: default; }
footer strong { white-space: nowrap; font-size: 1.8vh; }
@media (max-width: 760px) { .replay-page { grid-template-columns: 72% 28%; } .player-card { grid-template-columns: 11vh 1fr; } .player-card img { width: 11vh; } }
</style>
