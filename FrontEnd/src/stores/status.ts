import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

type membersType = {
  [sid: string]: {
    name: string
    username: string
    ip: string
    ready: boolean
    score: number
    decorator: {
      org: string
      chara: string
    }
  }
}

type playerType = {
  active: boolean
  hasnew: boolean
  uuid: string
  id: number
  name: string
  hand_count: number
  locked: string[]
  discarded: string[]
  score: number
}

export type roomType = {
  id: string
  name: string
  owner: string
  player: number
  watcher: number
  playing: boolean
  game: string
  members: number
  max_members: number
  has_password: boolean
  status: string
}

type GameInfoType = {
  init: boolean
  id: number
  activePlayer: number
  wallcount: number
  hands: string[]
  locked: string[]
  new: string
  discarded: string[]
  score: number
  actions: object
}

type UserInfo = {
  Uuid?: string
  Name?: string
  Room?: string
  Grade?: {
    '1st'?: number
    '2nd'?: number
    '3rd'?: number
    '4th'?: number
  }
  IP?: string
}

type CharacterInfo = {
  name: string
  head: string
  full: string
  confirm: string
  critical: string
}

type RuleOption = {
  index: number
  name: string
}

const playerStorageKey = 'mahjong.uuid'
const storageActions = ['confirm', 'critical', 'full', 'head'] as const

function loadPlayerUuid() {
  return sessionStorage.getItem(playerStorageKey) ?? ''
}

function savePlayerUuid(uuid: string) {
  if (!uuid) return
  sessionStorage.setItem(playerStorageKey, uuid)
  localStorage.removeItem(playerStorageKey)
}

function clearPlayerUuid() {
  sessionStorage.removeItem(playerStorageKey)
  localStorage.removeItem(playerStorageKey)
}

function characterNameFromKey(part = '') {
  return decodeURIComponent(part).replace(/\.[^.]+$/, '')
}

function parseCharacterUrls(urls: string[]): Record<string, CharacterInfo[]> {
  const groups: Record<string, Record<string, CharacterInfo>> = {}

  urls.forEach((rawUrl) => {
    let parts: string[]
    try {
      parts = new URL(rawUrl).pathname.split('/').filter((part) => part)
    } catch {
      parts = rawUrl.split('/').filter((part) => part)
    }
    if (parts.length < 3) return

    const action = parts[0] as (typeof storageActions)[number]
    if (!storageActions.includes(action)) return

    const org = decodeURIComponent(parts[1])
    const name = characterNameFromKey(parts.slice(2).join('/'))
    if (!org || !name) return

    groups[org] ??= {}
    groups[org][name] ??= {
      name,
      head: 'tilesvgs/Regular/Blank.svg',
      full: '',
      confirm: '',
      critical: '',
    }
    groups[org][name][action] = rawUrl
  })

  return Object.fromEntries(
    Object.entries(groups).map(([org, users]) => [
      org,
      Object.values(users).sort((a, b) => a.name.localeCompare(b.name)),
    ]),
  )
}

type ActionChoice = {
  tile: string
  tiles?: string[]
  selec: number[]
}

type LocalActions = {
  hu?: string | boolean
  kong?: ActionChoice[]
  pong?: ActionChoice
  chow?: ActionChoice[]
}

type GameActionEvent = {
  action: 'chow' | 'pong' | 'kong' | 'hu'
  tiles: string[]
  uuid: string
  nonce: number
}

const blankCharacter: CharacterInfo = {
  name: '',
  head: 'tilesvgs/Regular/Blank.svg',
  full: '',
  confirm: '',
  critical: '',
}

function normalizeRoom(raw: any): roomType {
  const id = raw.Id ?? raw.id ?? ''
  const name = raw.Name ?? raw.name ?? id
  const owner = raw.Owner ?? raw.owner ?? raw.OwnerUuid ?? raw.ownerUuid ?? ''
  const player = Number(raw.Player ?? raw.player ?? raw.members ?? 0)
  const watcher = Number(raw.Watcher ?? raw.watcher ?? 0)
  const playing = Boolean(raw.Playering ?? raw.playing ?? raw.status === 'playing')
  return {
    id,
    name,
    owner,
    player,
    watcher,
    playing,
    game: raw.game ?? '福州麻将',
    members: player,
    max_members: Number(raw.max_members ?? 4),
    has_password: Boolean(raw.has_password),
    status: raw.status ?? (playing ? 'playing' : 'waiting'),
  }
}

function parseTile(tile: string) {
  const matched = /^(Man|Pin|Sou)([1-9])$/.exec(tile)
  if (!matched) return null
  return {
    suit: matched[1],
    num: Number(matched[2]),
  }
}

function sortTileValue(tile: string) {
  const parsed = parseTile(tile)
  if (parsed) {
    const suitBase: Record<string, number> = { Man: 0, Pin: 10, Sou: 20 }
    return suitBase[parsed.suit] + parsed.num
  }
  const honors = ['Golden', 'Ton', 'Nan', 'Shaa', 'Pei', 'Chun', 'Hatsu', 'Haku']
  const honorIndex = honors.indexOf(tile)
  return honorIndex >= 0 ? honorIndex === 0 ? 0 : 30 + honorIndex : 100
}

function sortTiles(tiles: string[]) {
  return [...tiles].sort((a, b) => sortTileValue(a) - sortTileValue(b) || a.localeCompare(b))
}

function findSameIndices(tiles: string[], tile: string, limit: number) {
  const result: number[] = []
  tiles.forEach((handTile, index) => {
    if (handTile === tile && result.length < limit) result.push(index)
  })
  return result
}

function findTileIndex(tiles: string[], tile: string, used: Set<number>) {
  return tiles.findIndex((handTile, index) => handTile === tile && !used.has(index))
}

function tileFromSuit(suit: string, num: number) {
  return `${suit}${num}`
}

function removeFirstCopies(tiles: string[], tile: string, count: number) {
  const next = [...tiles]
  for (let i = 0; i < count; i += 1) {
    const index = next.indexOf(tile)
    if (index < 0) return null
    next.splice(index, 1)
  }
  return next
}

function canFormSets(tiles: string[], goldenCount: number): boolean {
  const sorted = sortTiles(tiles)
  if (sorted.length === 0) return goldenCount % 3 === 0
  const first = sorted[0]
  const sameCount = sorted.filter((tile) => tile === first).length

  for (let usedSame = Math.min(3, sameCount); usedSame >= 1; usedSame -= 1) {
    const needGolden = 3 - usedSame
    if (needGolden > goldenCount) continue
    const next = removeFirstCopies(sorted, first, usedSame)
    if (next && canFormSets(next, goldenCount - needGolden)) return true
  }

  const parsed = parseTile(first)
  if (!parsed || parsed.num > 7) return false

  let next = [...sorted]
  next.splice(0, 1)
  let needGolden = 0
  for (const needed of [parsed.num + 1, parsed.num + 2]) {
    const tile = tileFromSuit(parsed.suit, needed)
    const index = next.indexOf(tile)
    if (index >= 0) {
      next.splice(index, 1)
    } else {
      needGolden += 1
    }
  }
  return needGolden <= goldenCount && canFormSets(next, goldenCount - needGolden)
}

function canHu(tiles: string[], golden: string) {
  const goldenCount = tiles.filter((tile) => tile === golden).length
  const normalTiles = tiles.filter((tile) => tile && tile !== golden)
  if (goldenCount >= 3) return true

  const uniqueTiles = [...new Set(normalTiles)]
  for (const tile of uniqueTiles) {
    const sameCount = normalTiles.filter((handTile) => handTile === tile).length
    if (sameCount >= 2) {
      const next = removeFirstCopies(normalTiles, tile, 2)
      if (next && canFormSets(next, goldenCount)) return true
    }
    if (sameCount >= 1 && goldenCount >= 1) {
      const next = removeFirstCopies(normalTiles, tile, 1)
      if (next && canFormSets(next, goldenCount - 1)) return true
    }
  }

  return goldenCount >= 2 && canFormSets(normalTiles, goldenCount - 2)
}

function getChowChoices(hands: string[], out: string, golden: string): ActionChoice[] {
  const parsed = parseTile(out)
  if (!parsed || out === golden) return []

  return [
    [parsed.num - 2, parsed.num - 1],
    [parsed.num - 1, parsed.num + 1],
    [parsed.num + 1, parsed.num + 2],
  ].reduce<ActionChoice[]>((choices, nums) => {
    if (nums.some((num) => num < 1 || num > 9)) return choices

    const used = new Set<number>()
    const selec: number[] = []
    const tiles: string[] = []
    for (const num of nums) {
      const tile = tileFromSuit(parsed.suit, num)
      if (tile === golden) return choices
      const index = findTileIndex(hands, tile, used)
      if (index < 0) return choices
      used.add(index)
      selec.push(index)
      tiles.push(tile)
    }
    choices.push({ tile: out, tiles, selec })
    return choices
  }, [])
}

function computeAvailableActions(state: any, myInfo: any): LocalActions {
  const myIndex = Number(myInfo?.Index)
  if (!Number.isFinite(myIndex)) return {}

  const hands: string[] = myInfo?.Hands ?? []
  const newTile: string = myInfo?.New ?? ''
  const out: string = state.Out ?? ''
  const golden = 'Golden'
  const currentUser = Number(state.CurrentUser)
  const waitingResponse = Boolean(state.WaitingResponse)
  const actions: LocalActions = {}

  if (waitingResponse) {
    if (!out || out === golden || currentUser === myIndex) return actions

    if (canHu([...hands, out], golden)) actions.hu = out

    const pongSelec = findSameIndices(hands, out, 2)
    if (pongSelec.length === 2) actions.pong = { tile: out, selec: pongSelec }

    const kongSelec = findSameIndices(hands, out, 3)
    if (kongSelec.length === 3) actions.kong = [{ tile: out, selec: kongSelec }]

    if ((currentUser + 1) % 4 === myIndex) {
      const chow = getChowChoices(hands, out, golden)
      if (chow.length) actions.chow = chow
    }

    return actions
  }

  if (currentUser !== myIndex) return actions

  const selfHuTiles = newTile ? [...hands, newTile] : hands
  if (selfHuTiles.length && canHu(selfHuTiles, golden)) actions.hu = newTile || true

  const kongChoices = [...new Set(hands)]
    .filter((tile) => tile !== golden && findSameIndices(hands, tile, 4).length === 4)
    .map((tile) => ({ tile, selec: findSameIndices(hands, tile, 4) }))
  if (kongChoices.length) actions.kong = kongChoices

  return actions
}

export const statusStore = defineStore('status', () => {
  const isTyping = ref(false)
  const connected = ref(false)
  const isLogin = ref(false)
  const isRoomList = ref(false)
  const isRoom = ref(false)
  const isGaming = ref(false)
  const loginRequire = ref(false)
  const showIPLocation = ref(true)
  const characterGroups = ref<Record<string, CharacterInfo[]>>({})
  const tempOrg = ref('')
  const tempSelect = ref('')
  const username = ref('')
  const mysid = ref(loadPlayerUuid())
  const myid = ref(0)
  const userIP = ref('')
  const grades = ref({ first: 0, second: 0, third: 0, fourth: 0 })
  const members = ref<membersType>({})
  const players = ref<playerType[] | undefined>(undefined)
  const gameinfo = ref<GameInfoType>({
    init: false,
    id: 0,
    activePlayer: 0,
    wallcount: 0,
    hands: [],
    locked: [],
    new: '',
    discarded: [],
    score: 0,
    actions: {},
  })
  const roomlist = ref<roomType[]>([])
  const roomid = ref('')
  const ownerSid = ref('')
  const roomRule = ref({
    Rule: 0,
    ShowCritical: false,
    SkipOffline: false,
    MaxWaiting: 10,
  })
  const ruleList = ref<RuleOption[]>([])
  const actionEvent = ref<GameActionEvent | null>(null)

  const now = computed(() => {
    if (!isLogin.value) return 'nologin'
    if (isGaming.value) return 'gaming'
    if (isRoom.value) return 'room'
    if (isRoomList.value) return 'roomlist'
    return 'login'
  })
  const needsReconnect = computed(() => Boolean(roomid.value && !isRoom.value))
  const isOwner = computed(() => mysid.value !== '' && mysid.value === ownerSid.value)

  function authHeaders(): HeadersInit {
    if (!mysid.value) return {}
    return { 'X-Player-UUID': mysid.value }
  }

  async function apiFetch(url: string, options: RequestInit = {}) {
    const headers = new Headers(options.headers)
    if (mysid.value) {
      headers.set('X-Player-UUID', mysid.value)
    }
    const apiUrl = url.startsWith('/api') ? url : `/api${url}`
    return fetch(apiUrl, {
      ...options,
      headers,
    })
  }

  function applyUser(user?: UserInfo | null) {
    if (!user) return
    username.value = user.Name ?? username.value
    mysid.value = user.Uuid ?? mysid.value
    roomid.value = user.Room ?? roomid.value
    userIP.value = user.IP ?? userIP.value
    grades.value = {
      first: Number(user.Grade?.['1st'] ?? 0),
      second: Number(user.Grade?.['2nd'] ?? 0),
      third: Number(user.Grade?.['3rd'] ?? 0),
      fourth: Number(user.Grade?.['4th'] ?? 0),
    }
    if (mysid.value) savePlayerUuid(mysid.value)
    if (username.value) isLogin.value = true
  }

  async function loadConfig() {
    const res = await apiFetch('/config')
    if (!res.ok) return
    const data = await res.json()
    loginRequire.value = Boolean(data.LoginRequire)
    showIPLocation.value = Boolean(data.ShowIPLocation)
    characterGroups.value = parseCharacterUrls(Array.isArray(data.CharactersMap) ? data.CharactersMap : [])
    ensureTempOrg()
    applyUser(data.User)
  }

  async function loadCharacterMap() {
    await loadConfig()
  }

  function getCharacterGroups() {
    return Object.keys(characterGroups.value)
  }

  function getCharacters(group: string) {
    return characterGroups.value[group] ?? []
  }

  function getCharacter(group: string, name: string) {
    return characterGroups.value[group]?.find((character) => character.name === name) ?? blankCharacter
  }

  function getCharacterHead(group: string, name: string) {
    return getCharacter(group, name).head
  }

  function getCharacterCritical(group: string, name: string) {
    return getCharacter(group, name).critical
  }

  function setTempOrg(org: string) {
    tempOrg.value = org
    tempSelect.value = ''
  }

  function setTempSelect(name: string) {
    tempSelect.value = tempSelect.value === name ? '' : name
  }

  function resetTempCharacter() {
    tempSelect.value = ''
    ensureTempOrg()
  }

  function ensureTempOrg() {
    const groups = Object.keys(characterGroups.value)
    if (groups.length && (!tempOrg.value || !groups.includes(tempOrg.value))) {
      tempOrg.value = groups[0]
    }
  }

  async function login(account: string, password = '') {
    const res = await apiFetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ Account: account, Password: password }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '登录失败')
    }
    applyUser(data.User)
    isRoomList.value = false
    isRoom.value = false
    isGaming.value = false
    return data
  }

  async function logoutRemote() {
    await apiFetch('/logout', { method: 'POST' }).catch(() => undefined)
    logout()
  }

  async function enterLobby() {
    await refreshRooms()
    isRoomList.value = true
    isRoom.value = false
    isGaming.value = false
  }

  async function refreshRooms() {
    const res = await apiFetch('/rooms')
    const data = await res.json().catch(() => ({}))
    if (res.ok && data.status !== false) {
      updateRoomList(data.roomlist ?? [])
    }
  }

  async function createRoom(name: string) {
    const res = await apiFetch('/rooms/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ Name: name }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '创建房间失败')
    }
    const room = normalizeRoom(data.Room ?? {})
    updateRuleList(data.RuleList ?? [])
    setRoomOwner(room.owner || mysid.value)
    joinRoom(room.id)
    applyRoomState(data.State)
    return room
  }

  async function joinRoomRemote(id: string) {
    const res = await apiFetch(`/rooms/join?roomid=${encodeURIComponent(id)}`, { method: 'POST' })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '加入房间失败')
    }
    const room = normalizeRoom(data.Room ?? { Id: id })
    updateRuleList(data.RuleList ?? [])
    setRoomOwner(room.owner)
    joinRoom(room.id)
    applyRoomState(data.State)
    return room
  }

  async function leaveRoomRemote() {
    const res = await apiFetch('/rooms/user?action=leave', { method: 'POST' })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '离开房间失败')
    }
    leaveRoom('')
  }

  async function unready() {
    const res = await apiFetch('/rooms/user?action=unready', { method: 'POST' })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '取消准备失败')
    }
  }

  async function kickPlayer(targetUuid: string) {
    const res = await apiFetch(`/rooms/user?action=kick&target=${encodeURIComponent(targetUuid)}`, {
      method: 'POST',
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '踢出玩家失败')
    }
  }

  async function updateRule(rule: any) {
    const res = await apiFetch('/rooms/user?action=update_rule', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(rule),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '更新规则失败')
    }
  }

  async function startGameRemote() {
    const res = await apiFetch('/rooms/start', { method: 'POST' })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || data.status === false) {
      throw new Error(data.message ?? data.error ?? '开始游戏失败')
    }
  }

  function connectToServer(sid: string) {
    connected.value = true
    mysid.value = sid
    savePlayerUuid(sid)
  }

  function lostConnection() {
    connected.value = false
    isLogin.value = false
    isRoom.value = false
    isRoomList.value = false
    isGaming.value = false
    username.value = ''
    mysid.value = ''
    roomid.value = ''
    ownerSid.value = ''
    clearPlayerUuid()
    alert('Lost connection to server. Please refresh the page.')
  }

  function sucLogin(name: string) {
    username.value = name
    isLogin.value = true
  }

  function logout() {
    username.value = ''
    mysid.value = ''
    roomid.value = ''
    ownerSid.value = ''
    members.value = {}
    players.value = undefined
    myid.value = 0
    roomlist.value = []
    reSetGameInfo()
    isLogin.value = false
    isRoomList.value = false
    isRoom.value = false
    isGaming.value = false
    clearPlayerUuid()
  }

  function joinRoom(id: string) {
    roomid.value = id
    if (!ownerSid.value && Object.keys(members.value).length === 0) ownerSid.value = mysid.value
    resetTempCharacter()
    isRoom.value = true
    isRoomList.value = false
    updateRoomUser({
      uuid: mysid.value,
      name: username.value,
      ready: false,
    })
  }

  function leaveRoom(reason: string) {
    roomid.value = ''
    ownerSid.value = ''
    resetTempCharacter()
    players.value = undefined
    myid.value = 0
    members.value = {}
    isRoom.value = false
    isRoomList.value = true
    isGaming.value = false
    reSetGameInfo()
    if (reason) alert(reason)
  }

  function startGame() {
    isGaming.value = true
    isRoom.value = true
    isRoomList.value = false
  }

  function endGame() {
    isGaming.value = false
    players.value = undefined
    myid.value = 0
    reSetGameInfo()
  }

  function updateRoomList(data: any[]) {
    roomlist.value = data.map(normalizeRoom)
  }

  function updateRuleList(data: any[]) {
    ruleList.value = data
      .map((rule: any) => ({
        index: Number(rule.Index ?? rule.index ?? 0),
        name: String(rule.Name ?? rule.name ?? ''),
      }))
      .filter((rule) => rule.name)
  }

  function updateRoomUser(data: any) {
    if (!data.uuid) return
    setRoomOwner(data.Owner ?? data.owner ?? data.OwnerUuid ?? data.ownerUuid ?? '')
    if (data.left) {
      if (data.uuid === mysid.value) {
        leaveRoom(data.reason ?? data.message ?? '您已离开房间')
        return
      }
      delete members.value[data.uuid]
      return
    }
    const old = members.value[data.uuid]
    members.value[data.uuid] = {
      name: data.name ?? old?.name ?? data.uuid,
      username: data.name ?? old?.username ?? data.uuid,
      ip: old?.ip ?? '',
      ready: Boolean(data.ready ?? old?.ready ?? false),
      score: Number(data.Score ?? data.score ?? old?.score ?? 0),
      decorator: {
        org: data.decorator?.org ?? old?.decorator?.org ?? '',
        chara: data.decorator?.chara ?? old?.decorator?.chara ?? '',
      },
    }
  }

  function applyRoomState(state: any) {
    if (!state) return
    updateGameState(state)
  }

  function getMembers(sid = ''): object {
    const allMembers = { ...members.value }
    if (sid) {
      delete allMembers[sid]
    }
    return allMembers
  }

  function getInfoBySid(sid: string): object {
    return members.value[sid]
  }

  function getMemberidBySid(sid: string): number {
    return Object.keys(members.value).indexOf(sid)
  }

  function getPlayerById(id: number): playerType | undefined {
    return players.value?.find((player) => player.id === id)
  }

  function getPlayerBySid(sid: string): playerType | undefined {
    return players.value?.find((player) => player.uuid === sid)
  }

  function getOtherPlayerSidsBySeat(): string[] {
    if (!players.value) return []
    return [1, 2, 3]
      .map((offset) => (myid.value + offset) % 4)
      .map((index) => players.value?.find((player) => player.id === index)?.uuid ?? '')
      .filter((sid) => sid)
  }

  function setTyping(status: boolean) {
    isTyping.value = status
  }

  function reSetGameInfo() {
    gameinfo.value = {
      init: false,
      id: 0,
      activePlayer: 0,
      hands: [],
      locked: [],
      new: '',
      discarded: [],
      score: 0,
      actions: {},
      wallcount: 0,
    }
  }

  function reSetActions() {
    gameinfo.value.actions = {}
  }

  function updateGameState(newState: any) {
    if (newState.RoomRule) {
      roomRule.value = {
        Rule: Number(newState.RoomRule.Rule ?? 0),
        ShowCritical: Boolean(newState.RoomRule.ShowCritical),
        SkipOffline: Boolean(newState.RoomRule.SkipOffline),
        MaxWaiting: Number(newState.RoomRule.MaxWaiting ?? 10),
      }
    }
    const playerInfo = newState.PlayerInfo ?? []
    const myInfo = Array.isArray(newState.MyInfo) ? newState.MyInfo[0] : newState.MyInfo
    const allInfo = myInfo?.Uuid ? [myInfo, ...playerInfo] : playerInfo
    if (allInfo.length > 0) {
      const owner = allInfo.find((p: any) => Number(p.Index) === 0) ?? allInfo[0]
      setRoomOwner(owner.Uuid)
    }
    const activePlayer = !newState.WaitingResponse
      ? playerInfo.find((p: any) => Number(p.Index) === Number(newState.CurrentUser))
      : undefined
    const activePlayerUuid = activePlayer?.Uuid ?? ''
    const nextPlayers = playerInfo.map((p: any) => {
      members.value[p.Uuid] = {
        name: p.Name,
        username: p.Name,
        ip: '',
        ready: Boolean(p.Ready),
        score: Number(p.Score ?? 0),
        decorator: { org: p.CharactersGroup || '', chara: p.Character || '' },
      }
      if (p.Uuid === mysid.value) myid.value = p.Index
      const isCurrentDrawing = p.Uuid === activePlayerUuid
      return {
        active: isCurrentDrawing,
        hasnew: isCurrentDrawing,
        uuid: p.Uuid,
        id: p.Index,
        name: p.Name,
        hand_count: Number(p.HandsCount ?? p.Hands?.length ?? 0),
        locked: p.Lock ?? [],
        discarded: p.Discarded ?? [],
        score: Number(p.Score ?? 0),
      }
    })
    if (myInfo?.Uuid) {
      members.value[myInfo.Uuid] = {
        name: myInfo.Name,
        username: myInfo.Name,
        ip: '',
        ready: Boolean(myInfo.Ready),
        score: Number(myInfo.Score ?? 0),
        decorator: { org: myInfo.CharactersGroup || '', chara: myInfo.Character || '' },
      }
      myid.value = Number(myInfo.Index ?? 0)
    }

    if (!newState.Playing) {
      players.value = undefined
      reSetGameInfo()
      isGaming.value = false
      isRoom.value = true
      isRoomList.value = false
      return
    }

    players.value = nextPlayers
    gameinfo.value.init = true
    gameinfo.value.id = myid.value
    gameinfo.value.activePlayer =
      !newState.WaitingResponse && Number(newState.CurrentUser) === Number(myInfo?.Index)
        ? Number(myInfo.Index)
        : 5
    gameinfo.value.wallcount = Number(newState.WallCount ?? 0)
    gameinfo.value.hands = myInfo?.Hands ?? []
    gameinfo.value.locked = myInfo?.Lock ?? []
    gameinfo.value.discarded = myInfo?.Discarded ?? []
    gameinfo.value.new = myInfo?.New ?? ''
    gameinfo.value.score = Number(myInfo?.Score ?? 0)
    gameinfo.value.actions = computeAvailableActions(newState, myInfo)
    startGame()
  }

  function getGameInfo() {
    return gameinfo.value
  }

  function setRoomOwner(owner: string) {
    if (owner) ownerSid.value = owner
  }

  function updateActionEvent(data: any) {
    const action = data.action
    if (!['chow', 'pong', 'kong', 'hu'].includes(action) || !data.uuid) return
    actionEvent.value = {
      action,
      tiles: Array.isArray(data.tiles) ? data.tiles : [],
      uuid: data.uuid,
      nonce: Date.now(),
    }
  }

  return {
    now,
    connected,
    isLogin,
    isRoomList,
    isRoom,
    isGaming,
    loginRequire,
    showIPLocation,
    characterGroups,
    tempOrg,
    tempSelect,
    username,
    mysid,
    myid,
    userIP,
    grades,
    roomlist,
    roomid,
    ownerSid,
    roomRule,
    ruleList,
    actionEvent,
    needsReconnect,
    isOwner,
    members,
    players,
    isTyping,

    authHeaders,
    apiFetch,
    loadConfig,
    loadCharacterMap,
    login,
    logoutRemote,
    enterLobby,
    refreshRooms,
    createRoom,
    joinRoomRemote,
    leaveRoomRemote,
    unready,
    kickPlayer,
    updateRule,
    startGameRemote,
    connectToServer,
    lostConnection,
    sucLogin,
    logout,
    updateRoomList,
    updateRoomUser,
    updateGameState,
    joinRoom,
    leaveRoom,
    startGame,
    endGame,
    reSetActions,

    getMembers,
    getInfoBySid,
    getMemberidBySid,
    getPlayerById,
    getPlayerBySid,
    getOtherPlayerSidsBySeat,
    getGameInfo,
    getCharacterGroups,
    getCharacters,
    getCharacter,
    getCharacterHead,
    getCharacterCritical,
    setTempOrg,
    setTempSelect,
    resetTempCharacter,
    setTyping,
    updateActionEvent,
  }
})
