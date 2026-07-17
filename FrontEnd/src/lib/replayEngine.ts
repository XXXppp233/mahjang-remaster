import { shuffleMahjongWall } from './mahjongShuffle'
import type { ReplayData, ReplayEvent } from '@/stores/status'

export type ReplayPlayerState = {
  uuid: string
  name: string
  group: string
  chara: string
  locked: string[]
  hands: string[]
  newTile: string
  discarded: string[]
}

export type ReplayState = {
  players: ReplayPlayerState[]
  wall: string[]
  wallPosition: number
  currentPlayer: number
  pendingDiscard: { player: number; tile: string } | null
  lastEvent?: ReplayEvent
}

export const replayMetadataActions = new Set(['Ready', 'SetRule', 'Shuffle', 'Golden'])

const tileOrder = [
  ...Array.from({ length: 9 }, (_, index) => `Man${index + 1}`),
  ...Array.from({ length: 9 }, (_, index) => `Pin${index + 1}`),
  ...Array.from({ length: 9 }, (_, index) => `Sou${index + 1}`),
  'Ton', 'Nan', 'Shaa', 'Pei', 'Chun', 'Hatsu', 'Haku',
]

function sortTiles(tiles: string[]) {
  return [...tiles].sort((a, b) => tileOrder.indexOf(a) - tileOrder.indexOf(b))
}

function parseReady(event: ReplayEvent): ReplayPlayerState {
  const separator = event.value.indexOf(':')
  const name = separator >= 0 ? event.value.slice(0, separator) : event.user
  const character = separator >= 0 ? event.value.slice(separator + 1) : ''
  const slash = character.indexOf('/')
  return {
    uuid: event.user,
    name,
    group: slash >= 0 ? character.slice(0, slash) : '',
    chara: slash >= 0 ? character.slice(slash + 1) : character,
    locked: [], hands: [], newTile: '', discarded: [],
  }
}

function buildWall(golden: string, seed: number) {
  const base = [
    ...Array.from({ length: 9 }, (_, index) => `Pin${index + 1}`),
    ...Array.from({ length: 9 }, (_, index) => `Sou${index + 1}`),
    ...Array.from({ length: 9 }, (_, index) => `Man${index + 1}`),
    'Ton', 'Nan', 'Shaa', 'Pei', 'Chun', 'Hatsu', 'Haku',
  ].map((tile) => tile === golden ? 'Golden' : tile)
  return shuffleMahjongWall([...base, ...base, ...base, ...base], seed)
}

function removeTiles(source: string[], selected: string[]) {
  const result = [...source]
  selected.forEach((tile) => {
    const index = result.indexOf(tile)
    if (index >= 0) result.splice(index, 1)
  })
  return result
}

function draw(state: ReplayState, playerIndex: number) {
  const tile = state.wall[state.wallPosition]
  if (!tile) return
  state.players[playerIndex].newTile = tile
  state.wallPosition += 1
  state.currentPlayer = playerIndex
}

function removeClaimedDiscard(state: ReplayState) {
  if (!state.pendingDiscard) return
  const source = state.players[state.pendingDiscard.player].discarded
  const index = source.lastIndexOf(state.pendingDiscard.tile)
  if (index >= 0) source.splice(index, 1)
}

function applyAction(state: ReplayState, event: ReplayEvent) {
  const playerIndex = state.players.findIndex((player) => player.uuid === event.user)
  if (playerIndex < 0) return
  const player = state.players[playerIndex]
  const selected = event.value.split(',').filter((tile) => tileOrder.includes(tile) || tile === 'Golden')

  const continuesAfterUnclaimedDiscard = event.action === 'Discard'
    || (event.action === 'Kong' && selected.length === 4)
    || (event.action === 'Hu' && event.value === event.user)
  if (state.pendingDiscard && continuesAfterUnclaimedDiscard) {
    state.pendingDiscard = null
    draw(state, (state.currentPlayer + 1) % state.players.length)
  }

  if (event.action === 'Discard') {
    const tile = event.value
    if (player.newTile === tile) {
      player.newTile = ''
    } else {
      player.hands = removeTiles(player.hands, [tile])
      if (player.newTile) player.hands.push(player.newTile)
      player.newTile = ''
      player.hands = sortTiles(player.hands)
    }
    player.discarded.push(tile)
    state.currentPlayer = playerIndex
    state.pendingDiscard = { player: playerIndex, tile }
    return
  }

  if (event.action === 'Chow' || event.action === 'Pong') {
    const claimed = state.pendingDiscard?.tile
    removeClaimedDiscard(state)
    player.hands = sortTiles(removeTiles(player.hands, selected))
    player.locked.push(...(claimed ? [claimed, ...selected] : selected))
    state.currentPlayer = playerIndex
    state.pendingDiscard = null
    return
  }

  if (event.action === 'Kong') {
    const claimed = selected.length === 3 ? state.pendingDiscard?.tile : undefined
    if (claimed) removeClaimedDiscard(state)
    player.hands = removeTiles(player.hands, selected)
    if (player.newTile) player.hands.push(player.newTile)
    player.newTile = ''
    player.hands = sortTiles(player.hands)
    player.locked.push(...(claimed ? [claimed, ...selected] : selected))
    state.pendingDiscard = null
    draw(state, playerIndex)
    return
  }

  if (event.action === 'Hu') state.pendingDiscard = null
}

export function replayActions(data: ReplayData) {
  return data.events.filter((event) => !replayMetadataActions.has(event.action))
}

export function reconstructReplay(data: ReplayData, step: number): ReplayState {
  const players = data.events.filter((event) => event.action === 'Ready').map(parseReady)
  const golden = data.events.find((event) => event.action === 'Golden')?.value ?? ''
  const seed = Number(data.events.find((event) => event.action === 'Shuffle')?.value.match(/Seed:([^;]+)/)?.[1] ?? 0)
  const wall = buildWall(golden, seed)
  let wallPosition = 0
  players.forEach((player) => {
    player.hands = sortTiles(wall.slice(wallPosition, wallPosition + 16))
    wallPosition += 16
  })
  const state: ReplayState = { players, wall, wallPosition, currentPlayer: 0, pendingDiscard: null }
  if (players.length) draw(state, 0)
  const actions = replayActions(data).slice(0, step)
  actions.forEach((event) => { applyAction(state, event); state.lastEvent = event })
  return state
}
