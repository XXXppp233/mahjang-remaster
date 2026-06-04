<template>
  <div class="roomlist-page">
    <header class="lobby-header">
      <button class="ghost-btn" @click="backToIntro">返回</button>
      <h1>房间列表</h1>
      <button class="ghost-btn" @click="refreshRooms">刷新</button>
    </header>

    <section class="create-line">
      <input
        v-model="newRoomName"
        class="create-input"
        type="text"
        maxlength="24"
        placeholder="房间名称"
        @keydown.enter="createRoom"
      />
      <button
        class="main-btn"
        :disabled="creating || status.needsReconnect || !newRoomName.trim()"
        @click="createRoom"
      >
        创建房间
      </button>
    </section>

    <main class="room-grid" :class="{ empty: roomsForGrid.length === 0 }">
      <p v-if="roomsForGrid.length === 0" class="empty-text">暂无房间</p>

      <article
        v-for="room in roomsForGrid"
        :key="room.id"
        class="room-card"
        :class="{ reconnect: room.id === status.roomid && status.needsReconnect }"
      >
        <div>
          <h2>{{ room.name }}</h2>
          <p>{{ room.player }} / {{ room.max_members }}</p>
          <p>{{ room.playing ? '对局中' : '等待中' }}</p>
        </div>
        <button
          class="main-btn"
          :disabled="joiningRoomId !== '' || joinDisabled(room)"
          @click="joinRoom(room.id)"
        >
          {{ room.id === status.roomid && status.needsReconnect ? '重连' : '加入' }}
        </button>
      </article>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { statusStore } from '@/stores/status'

const status = statusStore()
const newRoomName = ref('')
const creating = ref(false)
const joiningRoomId = ref('')

const reconnectRoom = computed(() => {
  if (!status.needsReconnect) return null
  return (
    status.roomlist.find((room) => room.id === status.roomid) ?? {
      id: status.roomid,
      name: `未完成的对局 ${status.roomid.slice(0, 8)}`,
      owner: status.ownerSid,
      player: 0,
      watcher: 0,
      playing: true,
      game: '福州麻将',
      members: 0,
      max_members: 4,
      has_password: false,
      status: 'playing',
    }
  )
})

const roomsForGrid = computed(() => {
  const rooms = status.roomlist.filter((room) => room.id !== status.roomid)
  return reconnectRoom.value ? [reconnectRoom.value, ...rooms] : rooms
})

function joinDisabled(room) {
  return status.needsReconnect && room.id !== status.roomid
}

async function refreshRooms() {
  await status.refreshRooms()
}

async function createRoom() {
  const name = newRoomName.value.trim()
  if (!name || creating.value) return
  creating.value = true
  try {
    await status.createRoom(name)
  } finally {
    creating.value = false
    newRoomName.value = ''
  }
}

async function joinRoom(roomId) {
  if (joinDisabled({ id: roomId }) || joiningRoomId.value) return
  joiningRoomId.value = roomId
  try {
    await status.joinRoomRemote(roomId)
  } finally {
    joiningRoomId.value = ''
  }
}

function backToIntro() {
  status.isRoomList = false
}

onMounted(() => {
  // refreshRooms() // Removed redundant call
})
</script>

<style scoped>
.roomlist-page {
  min-height: 100vh;
  padding: 5vh 8vw;
  color: #111;
  background: #f7f7f7;
}

.lobby-header {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  margin-bottom: 5vh;
}

.lobby-header h1 {
  margin: 0;
  font-size: 7vh;
  line-height: 1;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: -0.2vh;
}

.ghost-btn,
.main-btn {
  border: 0.5vh solid #111;
  border-radius: 0.5vh;
  background: #fff;
  color: #111;
  cursor: pointer;
  transition: all 0.1s ease;
  box-shadow: 0.4vh 0.4vh 0 #111;
  font-weight: 700;
}

.ghost-btn {
  justify-self: start;
  padding: 1vh 2vw;
  font-size: 2.5vh;
}

.lobby-header .ghost-btn:last-child {
  justify-self: end;
}

.main-btn {
  padding: 1vh 2vw;
  font-size: 2.5vh;
}

.ghost-btn:hover,
.main-btn:hover:not(:disabled) {
  transform: translate(-0.2vh, -0.2vh);
  box-shadow: 0.6vh 0.6vh 0 #111;
}

.ghost-btn:active,
.main-btn:active:not(:disabled) {
  transform: translate(0.2vh, 0.2vh);
  box-shadow: 0.2vh 0.2vh 0 #111;
}

.main-btn:disabled {
  color: #888;
  border-color: #777;
  box-shadow: 0.2vh 0.2vh 0 #777;
  cursor: not-allowed;
  background: #eee;
}

.create-line {
  display: flex;
  justify-content: center;
  gap: 2vw;
  margin-bottom: 5vh;
}

.create-input {
  width: min(46vw, 520px);
  border: 0.5vh solid #111;
  border-radius: 0.5vh;
  background: #fff;
  color: #111;
  text-align: center;
  font-size: 3vh;
  padding: 1vh;
  box-shadow: 0.4vh 0.4vh 0 #111;
  font-weight: 600;
}

.create-input:focus {
  outline: none;
  border-color: #000;
  box-shadow: 0.6vh 0.6vh 0 #000;
}

.room-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(260px, 32vw));
  justify-content: center;
  gap: 4vh 4vw;
}

.room-grid.empty {
  display: flex;
  justify-content: center;
}

.empty-text {
  margin-top: 12vh;
  color: #111;
  font-size: 5vh;
  font-weight: 800;
}

.room-card {
  min-height: 24vh;
  border: 0.5vh solid #111;
  border-radius: 0.5vh;
  padding: 3vh 2vw;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  background: #fff;
  box-shadow: 0.8vh 0.8vh 0 #111;
}

.room-card.reconnect {
  grid-column: 1 / -1;
  width: min(67vw, 760px);
  justify-self: center;
  background: #fffbe6;
}

.room-card h2 {
  margin: 0 0 2vh;
  font-size: 4vh;
  line-height: 1.1;
  font-weight: 500;
  overflow-wrap: anywhere;
}

.room-card p {
  margin: 0.8vh 0;
  font-size: 2.5vh;
  color: #555;
}

@media (max-width: 760px) {
  .roomlist-page {
    padding: 4vh 5vw;
  }

  .lobby-header {
    grid-template-columns: 1fr;
    gap: 2vh;
    text-align: center;
  }

  .lobby-header h1 {
    font-size: 5vh;
  }

  .ghost-btn,
  .lobby-header .ghost-btn:last-child {
    justify-self: center;
  }

  .create-line {
    flex-direction: column;
    align-items: center;
  }

  .create-input {
    width: 86vw;
    font-size: 3.5vh;
  }

  .room-grid {
    grid-template-columns: minmax(0, 88vw);
  }

  .room-card.reconnect {
    width: auto;
  }
}
</style>
