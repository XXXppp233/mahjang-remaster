<!-- 进入房间后展示此组件 -->

<template>
  <div class="room-page">
    <div class="main-content">
      <div class="first-half">
        <Chat :myname="status.username" />
        <div class="room-center">
          <div class="other-players">
            <OtherInfo v-for="(member, sid) in members" :key="sid" :sid="sid" compact />
          </div>

          <section v-if="activePanel" class="center-panel">
            <div v-if="activePanel === 'org'" class="panel-content">
              <CharacterOrgSelect />
            </div>

            <div v-if="activePanel === 'rule'" class="panel-content rule-config">
              <div class="rule-row">
                <h3>出牌时间</h3>
                <div class="time-control">
                  <button class="square-btn" :disabled="isUnlimitedWaiting" @click="adjustWaiting(-1)">
                    -
                  </button>
                  <input
                    v-model.number="limitedWaiting"
                    type="number"
                    min="5"
                    max="20"
                    :disabled="isUnlimitedWaiting"
                    @change="normalizeWaiting"
                  />
                  <span>s</span>
                  <button class="square-btn" :disabled="isUnlimitedWaiting" @click="adjustWaiting(1)">
                    +
                  </button>
                  <label class="check-line">
                    <input type="checkbox" v-model="isUnlimitedWaiting" />
                    无限
                  </label>
                </div>
              </div>

              <div class="rule-row">
                <h3>跳过掉线玩家</h3>
                <label class="check-line">
                  <input
                    type="checkbox"
                    v-model="roomRule.SkipOffline"
                    :disabled="!status.loginRequire"
                  />
                  <span v-if="!status.loginRequire">当前为无登陆模式，必须跳过</span>
                </label>
              </div>

              <div class="rule-row">
                <h3>暴击动画</h3>
                <label class="check-line">
                  <input type="checkbox" v-model="roomRule.ShowCritical" />
                  会在吃碰杠时播放
                </label>
              </div>

              <div class="rule-row">
                <h3>麻将规则</h3>
                <div class="rule-options">
                  <label v-for="rule in rulesForPanel" :key="rule.index" class="radio-line">
                    <input type="radio" name="mahjong-rule" :value="rule.index" v-model.number="roomRule.Rule" />
                    {{ rule.name }}
                  </label>
                </div>
              </div>

              <div class="panel-footer">
                <button class="bold-btn" @click="updateRule">保存规则</button>
              </div>
            </div>
          </section>
        </div>
      </div>

      <CharacterDeck />
    </div>

    <aside v-if="!status.isGaming" class="room-actions">
      <div class="action-group">
        <button class="bold-btn danger" @click="leaveRoom">退出房间</button>
        <button v-if="isMeReady" class="bold-btn" @click="unready">取消准备</button>
      </div>

      <div v-if="!isMeReady" class="action-group">
        <button class="bold-btn" :class="{ selected: activePanel === 'org' }" @click="togglePanel('org')">
          {{ status.tempOrg || '组织选择' }}
        </button>
        <button class="bold-btn" :disabled="!status.tempSelect" @click="confirmSelection">
          确定
        </button>
      </div>

      <div v-if="status.isOwner" class="action-group owner-actions">
        <h3>房主工具</h3>
        <button class="bold-btn" :disabled="!canStartGame" @click="startGame">
          开始游戏
        </button>
        <button class="bold-btn" :class="{ selected: activePanel === 'rule' }" @click="togglePanel('rule')">
          更新规则
        </button>

        <button class="bold-btn" @click="showKickList = !showKickList">踢出玩家</button>
        <div v-if="showKickList" class="kick-panel">
          <button
            v-for="(member, sid) in members"
            :key="sid"
            class="bold-btn small danger"
            @click="kickPlayer(sid)"
          >
            踢出 {{ member.name }}
          </button>
          <p v-if="Object.keys(members).length === 0" class="empty-hint">房间里没别人了</p>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup>
import Chat from './Chat.vue'
import CharacterDeck from './CharacterDeck.vue'
import CharacterOrgSelect from './CharacterOrgSelect.vue'
import OtherInfo from './OtherInfo.vue'
import { computed, ref, reactive, watch } from 'vue'
import { statusStore } from '@/stores/status'

const status = statusStore()
const members = computed(() => status.getMembers(status.mysid))
const isMeReady = computed(() => status.members[status.mysid]?.ready)
const canStartGame = computed(() => {
  const allMembers = Object.values(status.members)
  return status.isOwner && allMembers.length === 4 && allMembers.every((member) => member.ready)
})
const activePanel = ref('')
const showKickList = ref(false)
const isUnlimitedWaiting = ref(false)
const limitedWaiting = ref(10)

const roomRule = reactive({
  Rule: 0,
  ShowCritical: false,
  SkipOffline: true,
  MaxWaiting: 10,
})

const rulesForPanel = computed(() =>
  status.ruleList.length ? status.ruleList : [{ index: 0, name: '福州麻将' }],
)

watch(
  () => status.roomRule,
  (newVal) => {
    Object.assign(roomRule, newVal)
    isUnlimitedWaiting.value = Number(newVal.MaxWaiting) > 20
    limitedWaiting.value = clampWaiting(Number(newVal.MaxWaiting) > 20 ? 10 : Number(newVal.MaxWaiting))
  },
  { immediate: true, deep: true },
)

watch(isUnlimitedWaiting, (next) => {
  roomRule.MaxWaiting = next ? 21 : limitedWaiting.value
})

watch(limitedWaiting, (next) => {
  if (!isUnlimitedWaiting.value) roomRule.MaxWaiting = clampWaiting(next)
})

watch(
  () => status.loginRequire,
  (loginRequire) => {
    if (!loginRequire) roomRule.SkipOffline = true
  },
  { immediate: true },
)

watch(isMeReady, (ready) => {
  if (ready && activePanel.value === 'org') activePanel.value = ''
})

function togglePanel(panel) {
  activePanel.value = activePanel.value === panel ? '' : panel
}

function clampWaiting(value) {
  const next = Number.isFinite(value) ? value : 10
  return Math.min(20, Math.max(5, next))
}

function normalizeWaiting() {
  limitedWaiting.value = clampWaiting(limitedWaiting.value)
}

function adjustWaiting(delta) {
  limitedWaiting.value = clampWaiting(limitedWaiting.value + delta)
}

async function leaveRoom() {
  try {
    await status.leaveRoomRemote()
  } catch (e) {
    alert(e.message)
  }
}

async function unready() {
  try {
    await status.unready()
  } catch (e) {
    alert(e.message)
  }
}

async function confirmSelection() {
  if (!status.tempSelect) return
  try {
    await status.apiFetch('/rooms/user?action=ready', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ready: true,
        decorator: {
          org: status.tempOrg,
          chara: status.tempSelect,
        },
      }),
    })
  } catch (e) {
    alert(e.message)
  }
}

async function kickPlayer(sid) {
  if (!confirm(`确定要踢出 ${status.members[sid]?.name} 吗？`)) return
  try {
    await status.kickPlayer(sid)
  } catch (e) {
    alert(e.message)
  }
}

async function updateRule() {
  try {
    normalizeWaiting()
    await status.updateRule({ ...roomRule })
    activePanel.value = ''
  } catch (e) {
    alert(e.message)
  }
}

async function startGame() {
  if (!canStartGame.value) return
  try {
    await status.startGameRemote()
  } catch (e) {
    alert(e.message)
  }
}
</script>

<style scoped>
.room-page {
  display: flex;
  width: 100%;
  height: 100vh;
  padding: 2vh;
  gap: 2vw;
  background: #f0f0f0;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.first-half {
  display: flex;
  margin-bottom: 2vh;
}

.room-center {
  display: flex;
  flex-direction: column;
  width: 64vw;
  min-width: 0;
  margin: 0 1vw;
  gap: 1.5vh;
}

.other-players {
  display: flex;
  gap: 1vw;
  height: 15vh;
  flex-wrap: wrap;
}

.center-panel {
  width: 60vw;
  height: 40vh;
  border: 0.5vh solid #111;
  border-radius: 0.5vh;
  background: #fff;
  box-shadow: 0.6vh 0.6vh 0 #111;
  padding: 1.5vh;
  overflow: hidden;
}

.panel-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 1.5vh;
}

.panel-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: auto;
}

.rule-config {
  overflow-y: auto;
}

.rule-row {
  display: grid;
  grid-template-columns: 12vw 1fr;
  align-items: center;
  gap: 1.5vw;
}

.rule-row h3 {
  margin: 0;
  font-size: 2vh;
  font-weight: 900;
}

.time-control,
.rule-options {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 1vw;
}

.time-control input {
  width: 5vw;
  border: 0.3vh solid #111;
  border-radius: 0.3vh;
  padding: 0.5vh 0.5vw;
  font-size: 1.8vh;
  font-weight: 800;
  text-align: center;
}

.square-btn {
  width: 4vh;
  height: 4vh;
  border: 0.3vh solid #111;
  border-radius: 0.3vh;
  background: #fff;
  font-size: 2vh;
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0.25vh 0.25vh 0 #111;
}

.check-line,
.radio-line {
  display: inline-flex;
  align-items: center;
  gap: 0.6vw;
  font-size: 1.8vh;
  font-weight: 700;
}

.room-actions {
  width: 18vw;
  display: flex;
  flex-direction: column;
  gap: 3vh;
  padding: 2vh;
  background: #fff;
  border: 0.4vh solid #111;
  box-shadow: 0.6vh 0.6vh 0 #111;
  border-radius: 0.5vh;
  overflow-y: auto;
}

.action-group {
  display: flex;
  flex-direction: column;
  gap: 1.5vh;
}

.action-group h3 {
  margin: 0 0 1vh;
  font-size: 2.2vh;
  font-weight: 800;
  border-bottom: 0.3vh solid #111;
  padding-bottom: 0.5vh;
}

.bold-btn {
  padding: 1.2vh 1vw;
  border: 0.4vh solid #111;
  background: #fff;
  color: #111;
  font-weight: 800;
  font-size: 1.8vh;
  cursor: pointer;
  box-shadow: 0.4vh 0.4vh 0 #111;
  transition: all 0.1s;
  border-radius: 0.3vh;
}

.bold-btn:hover,
.bold-btn.selected {
  transform: translate(-0.2vh, -0.2vh);
  box-shadow: 0.6vh 0.6vh 0 #111;
  background: #ffeb3b;
}

.bold-btn:active {
  transform: translate(0.2vh, 0.2vh);
  box-shadow: 0.2vh 0.2vh 0 #111;
}

.bold-btn:disabled,
.square-btn:disabled,
.time-control input:disabled {
  background: #ccc;
  border-color: #999;
  color: #666;
  cursor: not-allowed;
  transform: none;
  box-shadow: 0.4vh 0.4vh 0 #999;
}

.bold-btn.danger {
  background: #ff4d4d;
  color: #fff;
}

.bold-btn.small {
  padding: 0.5vh 0.8vw;
  font-size: 1.5vh;
  border-width: 0.3vh;
}

.kick-panel {
  display: flex;
  flex-direction: column;
  gap: 1vh;
  padding: 1vh;
  border: 0.3vh solid #111;
  background: #fafafa;
  border-radius: 0.3vh;
}

.empty-hint {
  font-size: 1.4vh;
  color: #666;
  text-align: center;
}
</style>
