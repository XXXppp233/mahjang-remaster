import { describe, expect, it } from 'vitest'
import { reconstructReplay } from '../replayEngine'
import type { ReplayData } from '@/stores/status'

const replay: ReplayData = {
  version: 1,
  room_id: 'test',
  exported_at: 1_700_000_000_000,
  events: [
    { seq: 1, timestamp: 1, user: 'u1', action: 'Ready', value: 'A:Org/A' },
    { seq: 2, timestamp: 2, user: 'u2', action: 'Ready', value: 'B:Org/B' },
    { seq: 3, timestamp: 3, user: 'u3', action: 'Ready', value: 'C:Org/C' },
    { seq: 4, timestamp: 4, user: 'u4', action: 'Ready', value: 'D:Org/D' },
    { seq: 5, timestamp: 5, user: 'Server', action: 'SetRule', value: 'Rule:0;Skip:true;MaxW:21' },
    { seq: 6, timestamp: 6, user: 'Server', action: 'Shuffle', value: 'Seed:666;Algo:FY-XorShift32-v1;Wall:v1' },
    { seq: 7, timestamp: 7, user: 'Server', action: 'Golden', value: 'Pei' },
  ],
}

describe('reconstructReplay', () => {
  it('deals four Fuzhou hands and gives player one a new tile', () => {
    const state = reconstructReplay(replay, 0)
    expect(state.players.map((player) => player.hands.length)).toEqual([16, 16, 16, 16])
    expect(state.players[0].newTile).not.toBe('')
    expect(state.wallPosition).toBe(65)
  })

  it('draws for the next player when a discard is unclaimed', () => {
    const initial = reconstructReplay(replay, 0)
    const firstTile = initial.players[0].newTile
    const secondDraw = initial.wall[65]
    const withActions: ReplayData = {
      ...replay,
      events: [...replay.events,
        { seq: 8, timestamp: 1000, user: 'u1', action: 'Discard', value: firstTile },
        { seq: 9, timestamp: 2000, user: 'u2', action: 'Discard', value: secondDraw },
      ],
    }
    const state = reconstructReplay(withActions, 2)
    expect(state.wallPosition).toBe(66)
    expect(state.players[1].discarded).toEqual([secondDraw])
    expect(state.currentPlayer).toBe(1)
  })
})
